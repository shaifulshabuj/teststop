package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shaifulshabuj/teststop/internal/ai"
	"github.com/shaifulshabuj/teststop/internal/mandate"
	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
	"github.com/shaifulshabuj/teststop/internal/reporter"
	"github.com/shaifulshabuj/teststop/pkg/scenario"

	"github.com/spf13/cobra"
)

var (
	runPath      string
	runDepth     string
	runOutput    string
	runThreshold int
	runQuiet     bool
	runNoColor   bool
	runProvider  string
	runModel     string
	runDryRun    bool
	runReportDir string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a full adversarial testing cycle on the current project.",
	Long: `Scan the project, compose the mandate with detected context and memory,
ask the configured AI to generate adversarial-user scenarios, then update
confidence memory and emit a report.

Exit codes:
  0  confidence threshold met — safe to proceed
  1  below threshold — review required
  2  critical failures — do not deploy
  3  teststop internal error

Provider selection:
  --provider claude|openai (or TESTSTOP_AI env var). If unset, teststop
  picks claude when ANTHROPIC_API_KEY is set, otherwise openai when
  OPENAI_API_KEY is set.`,
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringVar(&runPath, "path", ".", "Path to the project under test")
	runCmd.Flags().StringVar(&runDepth, "depth", "normal", "Testing depth: light | normal | aggressive")
	runCmd.Flags().StringVar(&runOutput, "output", "text", "Output format: text | json | markdown")
	runCmd.Flags().IntVar(&runThreshold, "threshold", 85, "Confidence threshold (0-100)")
	runCmd.Flags().BoolVar(&runQuiet, "quiet", false, "Suppress non-essential human output")
	runCmd.Flags().BoolVar(&runNoColor, "no-color", false, "Disable ANSI colors in text output")
	runCmd.Flags().StringVar(&runProvider, "provider", "", "AI provider override: claude | openai")
	runCmd.Flags().StringVar(&runModel, "model", "", "Model identifier override")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Compose and print the mandate without calling an AI")
	runCmd.Flags().StringVar(&runReportDir, "report-dir", "", "Optional directory to write a copy of the JSON report")
}

func runRun(cmd *cobra.Command, _ []string) error {
	absPath, err := filepath.Abs(runPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	if !runQuiet && runOutput == "text" {
		fmt.Fprintf(cmd.OutOrStderr(), "teststop: analysing %s\n", absPath)
	}

	projectCtx, err := reader.Read(absPath)
	if err != nil {
		return fmt.Errorf("read project: %w", err)
	}

	mem, err := memory.Load(absPath)
	if err != nil {
		return fmt.Errorf("load memory: %w", err)
	}
	previousConfidence := mem.OverallConfidence

	prompt, err := mandate.Compose(mandate.Input{
		Project: projectCtx,
		Memory:  mem,
		Depth:   mandate.Depth(strings.ToLower(strings.TrimSpace(runDepth))),
	})
	if err != nil {
		return fmt.Errorf("compose mandate: %w", err)
	}

	if runDryRun {
		_, err := fmt.Fprint(cmd.OutOrStdout(), prompt)
		return err
	}

	adapter, err := ai.New(ai.Config{
		Provider: runProvider,
		Model:    runModel,
	})
	if err != nil {
		return fmt.Errorf("configure ai: %w", err)
	}
	if !runQuiet && runOutput == "text" {
		fmt.Fprintf(cmd.OutOrStderr(), "teststop: provider %s — generating scenarios\n", adapter.Name())
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()
	scenarios, err := adapter.Generate(ctx, prompt)
	if err != nil {
		if errors.Is(err, ai.ErrEmptyResponse) {
			return fmt.Errorf("ai returned no scenarios")
		}
		return fmt.Errorf("generate scenarios: %w", err)
	}
	if len(scenarios) == 0 {
		return fmt.Errorf("ai returned no scenarios")
	}

	// In v0.1 scenarios are not executed; we treat every generated
	// scenario as "passed but unverified" so memory accumulates slowly,
	// and we surface that honestly in the report.
	now := time.Now().UTC()
	results := make([]scenario.Result, 0, len(scenarios))
	areaByScenario := make(map[string]string, len(scenarios))
	for _, s := range scenarios {
		results = append(results, scenario.Result{
			ScenarioID: s.ScenarioID,
			Passed:     true,
			Notes:      "generated — not executed in v0.1",
			RunAt:      now,
		})
		areaByScenario[s.ScenarioID] = s.ConfidenceArea
	}
	mem.Apply(memory.RunOutcome{
		When:           now,
		Results:        results,
		AreaByScenario: areaByScenario,
	})
	retired := mem.Retire(now)
	mem.Revive()
	if err := memory.Save(absPath, mem); err != nil {
		return fmt.Errorf("save memory: %w", err)
	}

	threshold := float64(runThreshold) / 100.0
	run := reporter.Run{
		RunID:              newRunID(),
		Timestamp:          now,
		Project:            projectCtx.Name,
		Language:           projectCtx.Language,
		ProjectType:        projectCtx.Type,
		OverallConfidence:  mem.OverallConfidence,
		PreviousConfidence: previousConfidence,
		ConfidenceDelta:    round2(mem.OverallConfidence - previousConfidence),
		MaturityStage:      mem.MaturityStage,
		Threshold:          threshold,
		ScenariosGenerated: len(scenarios),
		ScenariosPassed:    0,
		ScenariosFailed:    0,
		ScenariosUnknown:   len(scenarios),
		Scenarios:          scenarios,
		Failures:           nil,
		RetiredThisRun:     retired,
		StableAreas:        mem.StableAreas(),
		VolatileAreas:      mem.VolatileAreas(),
		Notes: []string{
			"v0.1: scenarios are generated but not executed. Confidence reflects breadth of probing, not verification.",
		},
	}
	run.ReadyForDeploy = run.ExitCode() == 0

	if runReportDir != "" {
		if err := writeJSONReportCopy(runReportDir, run); err != nil {
			return fmt.Errorf("write report copy: %w", err)
		}
	}

	if err := renderRun(cmd, run); err != nil {
		return err
	}
	if code := run.ExitCode(); code != 0 {
		os.Exit(code)
	}
	return nil
}

func renderRun(cmd *cobra.Command, run reporter.Run) error {
	switch strings.ToLower(runOutput) {
	case "json":
		return reporter.WriteJSON(cmd.OutOrStdout(), run)
	case "markdown", "md":
		return reporter.WriteMarkdown(cmd.OutOrStdout(), run)
	case "text", "":
		return reporter.WriteText(cmd.OutOrStdout(), run)
	default:
		return fmt.Errorf("unknown --output %q (want text|json|markdown)", runOutput)
	}
}

func writeJSONReportCopy(dir string, run reporter.Run) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("teststop-%s.json", run.Timestamp.Format("20060102-150405"))
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer f.Close()
	return reporter.WriteJSON(f, run)
}

func newRunID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(b[:])
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
