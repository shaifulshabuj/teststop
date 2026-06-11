package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/shaifulshabuj/teststop/internal/ai"
	"github.com/shaifulshabuj/teststop/internal/config"
	"github.com/shaifulshabuj/teststop/internal/executor"
	"github.com/shaifulshabuj/teststop/internal/mandate"
	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
	"github.com/shaifulshabuj/teststop/internal/reporter"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

var (
	runPath          string
	runDepth         string
	runOutput        string
	runThreshold     int
	runNoColor       bool
	runQuiet         bool
	runTarget        string
	runConcurrency   int
	runAIConcurrency int
	runExecTimeout   time.Duration
	runMaxRetries    int
)

var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Run adversarial user testing on a project",
	Long:  `Scan the project, compose a mandate, trigger AI to generate test scenarios, update confidence memory, and report results.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCmdE,
}

func init() {
	runCmd.Flags().StringVar(&runPath, "path", ".", "Path to the project to test")
	runCmd.Flags().StringVar(&runDepth, "depth", "normal", "Testing depth: light | normal | aggressive")
	runCmd.Flags().StringVar(&runOutput, "output", "text", "Output format: json | text | markdown")
	runCmd.Flags().IntVar(&runThreshold, "threshold", 80, "Confidence threshold (0-100)")
	runCmd.Flags().BoolVar(&runNoColor, "no-color", false, "Disable ANSI color output (for agents)")
	runCmd.Flags().BoolVar(&runQuiet, "quiet", false, "Minimal output")
	runCmd.Flags().StringVar(&runTarget, "target", "", "Base URL of the running system to execute against (e.g. http://localhost:8080); empty = static validation only")
	runCmd.Flags().IntVar(&runConcurrency, "concurrency", 4, "Max scenarios executed in parallel")
	runCmd.Flags().IntVar(&runAIConcurrency, "ai-concurrency", 1, "Max concurrent AI-mode executions (default 1 to avoid rate-limit exhaustion)")
	runCmd.Flags().DurationVar(&runExecTimeout, "exec-timeout", 10*time.Second, "Per-request execution timeout")
	runCmd.Flags().IntVar(&runMaxRetries, "max-retries", 2, "Retries for transient HTTP execution failures")
}

func runCmdE(cmd *cobra.Command, args []string) error {
	start := time.Now()

	// 1. Resolve path — positional arg overrides --path flag.
	path := runPath
	if len(args) > 0 {
		path = args[0]
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("path: %w", err)
	}

	// 1b. Resolve run settings with precedence: config file < env < CLI flags.
	//     This mutates the package-level flag vars in place so the rest of the
	//     pipeline reads the resolved values transparently.
	if err := resolveRunSettings(cmd, absPath); err != nil {
		return err
	}

	// 2. Scan project.
	ctx, err := reader.ScanProject(absPath)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	// 3. Load memory.
	mem, err := memory.Load(absPath)
	if err != nil {
		return fmt.Errorf("memory: %w", err)
	}

	// 4. Compose mandate.
	mandateStr, err := mandate.Compose(ctx, mem, mandate.Options{Depth: runDepth})
	if err != nil {
		return fmt.Errorf("mandate: %w", err)
	}

	// 5. Detect AI adapter.
	adapter, err := ai.Detect()
	if err != nil {
		return fmt.Errorf("ai: %w", err)
	}

	// 6. Generate scenarios.
	scenarios, err := adapter.GenerateScenarios(mandateStr)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// 7. Execute scenarios (hybrid: HTTP/AI when --target set, else static validation).
	execCfg := executor.Config{
		Target:        runTarget,
		Timeout:       runExecTimeout,
		MaxRetries:    runMaxRetries,
		Concurrency:   runConcurrency,
		AIConcurrency: runAIConcurrency,
		Adapter:       adapter,
	}
	executions := executor.Run(cmd.Context(), execCfg, scenarios)

	// 8. Update memory from real execution outcomes. Skipped results (AI infra
	//    errors, rate limits) carry no verdict about the target, so they must not
	//    move confidence in either direction.
	for _, r := range executions {
		if r.Skipped || r.Area == "" {
			continue
		}
		mem.UpdateArea(r.Area, r.Passed)
	}

	// 9. Build run result (after memory update so reported state reflects this run).
	result := buildRunResult(ctx, scenarios, executions, mem, adapter.Name(), runDepth, runTarget, start)
	result.ExitCode = reporter.ExitCodeFor(result, float64(runThreshold)/100.0)

	// 9. Retire eligible areas.
	retired, err := mem.RetireEligibleAreas(absPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: retire: %v\n", err)
	}
	result.RetiredAreas = retired

	// 10. Save memory.
	if err := mem.Save(absPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: save memory: %v\n", err)
	}

	// 11. Output results.
	if err := writeOutput(cmd.OutOrStdout(), result, runOutput, runNoColor, runQuiet); err != nil {
		return err
	}

	// 12. Save markdown report always (regardless of --output flag).
	if runOutput != "markdown" {
		_, _ = reporter.SaveMarkdownReport(absPath, result)
	}

	// 13. Exit with appropriate code.
	os.Exit(result.ExitCode)
	return nil
}

// buildRunResult constructs the RunResult from the scenario list, execution
// outcomes, and memory state.
func buildRunResult(
	ctx reader.ProjectContext,
	scenarios []scenario.Scenario,
	executions []executor.ExecutionResult,
	mem *memory.Memory,
	adapterName, depth, target string,
	start time.Time,
) reporter.RunResult {
	// Index scenario titles by ID for richer failure reporting.
	titleByID := make(map[string]string, len(scenarios))
	for _, s := range scenarios {
		titleByID[s.ScenarioID] = s.Title
	}

	// Collect failures from real execution outcomes. Skipped (infrastructure)
	// results are not verdicts about the target, so they are never failures.
	var failures []reporter.Failure
	for _, r := range executions {
		if r.Passed || r.Skipped {
			continue
		}
		desc := r.ActualBehavior
		if r.FailureReason != "" {
			desc = r.FailureReason
		}
		failures = append(failures, reporter.Failure{
			ScenarioID:  r.ScenarioID,
			Title:       titleByID[r.ScenarioID],
			Area:        r.Area,
			Priority:    r.Priority,
			Description: desc,
		})
	}

	// Collect stable and volatile area names.
	stable := mem.GetStableAreas()
	volatile := mem.GetVolatileAreas()

	stableNames := make([]string, len(stable))
	for i, a := range stable {
		stableNames[i] = a.Name
	}
	volatileNames := make([]string, len(volatile))
	for i, a := range volatile {
		volatileNames[i] = a.Name
	}

	// Average confidence across all non-retired areas.
	var totalConf float64
	count := 0
	for _, a := range mem.Areas {
		if !a.Retired {
			totalConf += a.Confidence
			count++
		}
	}
	var avgConf float64
	if count > 0 {
		avgConf = totalConf / float64(count)
	}

	return reporter.RunResult{
		ProjectName:     ctx.Name,
		ProjectPath:     ctx.Path,
		Language:        ctx.Language,
		SystemType:      ctx.Type,
		Timestamp:       start,
		Duration:        time.Since(start),
		Scenarios:       scenarios,
		Executions:      executions,
		ExecSummary:     reporter.SummarizeExecutions(executions, target),
		Failures:        failures,
		StableAreas:     stableNames,
		VolatileAreas:   volatileNames,
		ConfidenceScore: avgConf,
		AdapterName:     adapterName,
		Depth:           depth,
	}
}

// resolveRunSettings applies the settings precedence for `teststop run`:
//
//	config file (.teststop/config.yaml)  <  TESTSTOP_RUN_* env vars  <  CLI flags
//
// It works bottom-up. It seeds the package-level flag vars from the config file
// (lowest tier), then overlays env vars, then leaves any flag the user set
// explicitly on the command line untouched (highest tier) — cobra's
// Flags().Changed tells us which those are. Because cobra has already populated
// the flag vars with their built-in defaults, a setting absent from every tier
// keeps that default.
func resolveRunSettings(cmd *cobra.Command, projectPath string) error {
	cfg, err := config.Load(projectPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	changed := func(name string) bool { return cmd.Flags().Changed(name) }

	// Tier 1 (lowest): config file. Only apply when the user did NOT pass the
	// corresponding flag — an explicit flag (tier 3) always wins over the file,
	// and env (tier 2) is layered on top below.
	if cfg.Depth != nil && !changed("depth") {
		runDepth = *cfg.Depth
	}
	if cfg.Output != nil && !changed("output") {
		runOutput = *cfg.Output
	}
	if cfg.Threshold != nil && !changed("threshold") {
		runThreshold = *cfg.Threshold
	}
	if cfg.NoColor != nil && !changed("no-color") {
		runNoColor = *cfg.NoColor
	}
	if cfg.Quiet != nil && !changed("quiet") {
		runQuiet = *cfg.Quiet
	}
	if cfg.Target != nil && !changed("target") {
		runTarget = *cfg.Target
	}
	if cfg.Concurrency != nil && !changed("concurrency") {
		runConcurrency = *cfg.Concurrency
	}
	if cfg.AIConcurrency != nil && !changed("ai-concurrency") {
		runAIConcurrency = *cfg.AIConcurrency
	}
	if cfg.ExecTimeout != nil && !changed("exec-timeout") {
		runExecTimeout = *cfg.ExecTimeout
	}
	if cfg.MaxRetries != nil && !changed("max-retries") {
		runMaxRetries = *cfg.MaxRetries
	}

	// Tier 2: environment variables. These override the file but still yield to
	// an explicit CLI flag. A malformed numeric/bool/duration env value is a
	// hard error so misconfiguration fails loudly.
	return applyRunEnv(changed)
}

// applyRunEnv overlays TESTSTOP_RUN_* environment variables onto the flag vars,
// skipping any setting the user passed explicitly on the command line.
func applyRunEnv(changed func(string) bool) error {
	if v, ok := os.LookupEnv("TESTSTOP_RUN_DEPTH"); ok && !changed("depth") {
		runDepth = v
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_OUTPUT"); ok && !changed("output") {
		runOutput = v
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_THRESHOLD"); ok && !changed("threshold") {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("config: TESTSTOP_RUN_THRESHOLD=%q is not an integer: %w", v, err)
		}
		runThreshold = n
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_NO_COLOR"); ok && !changed("no-color") {
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("config: TESTSTOP_RUN_NO_COLOR=%q is not a boolean: %w", v, err)
		}
		runNoColor = b
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_QUIET"); ok && !changed("quiet") {
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("config: TESTSTOP_RUN_QUIET=%q is not a boolean: %w", v, err)
		}
		runQuiet = b
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_TARGET"); ok && !changed("target") {
		runTarget = v
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_CONCURRENCY"); ok && !changed("concurrency") {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("config: TESTSTOP_RUN_CONCURRENCY=%q is not an integer: %w", v, err)
		}
		runConcurrency = n
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_AI_CONCURRENCY"); ok && !changed("ai-concurrency") {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("config: TESTSTOP_RUN_AI_CONCURRENCY=%q is not an integer: %w", v, err)
		}
		runAIConcurrency = n
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_EXEC_TIMEOUT"); ok && !changed("exec-timeout") {
		d, err := time.ParseDuration(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("config: TESTSTOP_RUN_EXEC_TIMEOUT=%q is not a duration: %w", v, err)
		}
		runExecTimeout = d
	}
	if v, ok := os.LookupEnv("TESTSTOP_RUN_MAX_RETRIES"); ok && !changed("max-retries") {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("config: TESTSTOP_RUN_MAX_RETRIES=%q is not an integer: %w", v, err)
		}
		runMaxRetries = n
	}
	return nil
}

// writeOutput dispatches to the correct reporter format.
func writeOutput(w io.Writer, result reporter.RunResult, output string, noColor, quiet bool) error {
	switch output {
	case "json":
		return reporter.WriteJSON(w, result)
	case "markdown":
		return reporter.WriteMarkdown(w, result)
	default: // "text"
		if quiet {
			switch result.ExitCode {
			case 0:
				fmt.Fprintln(w, "OK")
			case 1:
				fmt.Fprintln(w, "REVIEW")
			case 2:
				fmt.Fprintln(w, "CRITICAL")
			default:
				fmt.Fprintln(w, "ERROR")
			}
			return nil
		}
		return reporter.WriteText(w, result, noColor)
	}
}
