package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/shaifulshabuj/teststop/internal/ai"
	"github.com/shaifulshabuj/teststop/internal/mandate"
	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
	"github.com/shaifulshabuj/teststop/internal/reporter"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

var (
	runPath      string
	runDepth     string
	runOutput    string
	runThreshold int
	runNoColor   bool
	runQuiet     bool
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

	// 7. Build run result.
	result := buildRunResult(ctx, scenarios, mem, adapter.Name(), runDepth, start)
	result.ExitCode = reporter.ExitCodeFor(result, float64(runThreshold)/100.0)

	// 8. Update memory — in v0.1 we don't execute scenarios, so all areas get a pass.
	for _, s := range scenarios {
		if s.ConfidenceArea != "" {
			mem.UpdateArea(s.ConfidenceArea, true)
		}
	}

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

// buildRunResult constructs the RunResult from the scenario list and memory state.
func buildRunResult(
	ctx reader.ProjectContext,
	scenarios []scenario.Scenario,
	mem *memory.Memory,
	adapterName, depth string,
	start time.Time,
) reporter.RunResult {
	// Collect failures: critical-priority scenarios that have failure modes.
	var failures []reporter.Failure
	for _, s := range scenarios {
		if len(s.FailureModes) > 0 && s.Priority == "critical" {
			failures = append(failures, reporter.Failure{
				ScenarioID:  s.ScenarioID,
				Title:       s.Title,
				Area:        s.ConfidenceArea,
				Priority:    s.Priority,
				Description: strings.Join(s.FailureModes, "; "),
			})
		}
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
		Failures:        failures,
		StableAreas:     stableNames,
		VolatileAreas:   volatileNames,
		ConfidenceScore: avgConf,
		AdapterName:     adapterName,
		Depth:           depth,
	}
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
