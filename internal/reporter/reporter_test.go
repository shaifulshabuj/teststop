package reporter_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/internal/executor"
	"github.com/shaifulshabuj/teststop/internal/reporter"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

func makeResult() reporter.RunResult {
	return reporter.RunResult{
		ProjectName:     "testapp",
		ProjectPath:     "/tmp/testapp",
		Language:        "Go",
		SystemType:      "api",
		Timestamp:       time.Now(),
		Duration:        2500 * time.Millisecond,
		ConfidenceScore: 0.85,
		Depth:           "normal",
		AdapterName:     "claude",
		ExecSummary: reporter.ExecSummary{
			Executed: true,
			Count:    1,
			Passed:   0,
			Failed:   1,
			Target:   "http://localhost:8080",
		},
		Scenarios: []scenario.Scenario{
			{
				ScenarioID:     "s-001",
				Title:          "Empty auth submission",
				Priority:       "high",
				ConfidenceArea: "auth",
				IsEdgeCase:     false,
			},
		},
		Failures: []reporter.Failure{
			{
				ScenarioID:  "s-001",
				Title:       "Empty auth submission",
				Area:        "auth",
				Priority:    "high",
				Description: "Form accepts empty credentials",
			},
		},
		StableAreas:   []string{"static-assets"},
		VolatileAreas: []string{"auth"},
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	result := makeResult()
	if err := reporter.WriteJSON(&buf, result); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"project_name"`) {
		t.Error("JSON output missing project_name")
	}
	if !strings.Contains(out, `"testapp"`) {
		t.Error("JSON output missing project name value")
	}
	if !strings.Contains(out, `"failures"`) {
		t.Error("JSON output missing failures")
	}
}

func TestWriteText_noColor(t *testing.T) {
	var buf bytes.Buffer
	result := makeResult()
	if err := reporter.WriteText(&buf, result, true); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "testapp") {
		t.Error("text output missing project name")
	}
	if !strings.Contains(out, "SCENARIOS") {
		t.Error("text output missing SCENARIOS header")
	}
	if !strings.Contains(out, "FAILURES") {
		t.Error("text output missing FAILURES header")
	}
	// No ANSI codes in no-color mode
	if strings.Contains(out, "\033[") {
		t.Error("no-color mode should not contain ANSI codes")
	}
}

func TestWriteMarkdown(t *testing.T) {
	var buf bytes.Buffer
	result := makeResult()
	if err := reporter.WriteMarkdown(&buf, result); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "# teststop Report") {
		t.Error("markdown missing H1 header")
	}
	if !strings.Contains(out, "testapp") {
		t.Error("markdown missing project name")
	}
	if !strings.Contains(out, "| Priority |") {
		t.Error("markdown missing scenarios table")
	}
}

// makePredictedResult is a run with no --target: predictions only, no failures.
func makePredictedResult() reporter.RunResult {
	r := makeResult()
	r.ExecSummary = reporter.ExecSummary{Executed: false, Count: 1, Passed: 1, Failed: 0}
	r.Failures = nil
	return r
}

func TestWriteText_predictedFraming(t *testing.T) {
	var buf bytes.Buffer
	if err := reporter.WriteText(&buf, makePredictedResult(), true); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "PREDICTED RISKS") {
		t.Error("predicted run should show PREDICTED RISKS header")
	}
	if !strings.Contains(out, "PREDICTED CONFIDENCE") {
		t.Error("predicted run should label confidence as PREDICTED CONFIDENCE")
	}
	if !strings.Contains(out, "--target") {
		t.Error("predicted run should tell the user to run with --target")
	}
	// Must not present a bare "CONFIDENCE:" as if verified.
	if strings.Contains(out, "\nCONFIDENCE:") {
		t.Error("predicted run must not show a verified CONFIDENCE line")
	}
}

func TestWriteMarkdown_predictedFraming(t *testing.T) {
	var buf bytes.Buffer
	if err := reporter.WriteMarkdown(&buf, makePredictedResult()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Predicted Risks") {
		t.Error("predicted markdown should use 'Predicted Risks' heading")
	}
	if !strings.Contains(out, "predicted (no `--target`") {
		t.Error("predicted markdown should state mode is predicted")
	}
}

func TestExecSummary_JSONShape(t *testing.T) {
	var buf bytes.Buffer
	if err := reporter.WriteJSON(&buf, makeResult()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"executed": true`) {
		t.Error("JSON exec_summary should carry executed=true for an executed run")
	}
	if !strings.Contains(out, `"count": 1`) {
		t.Error("JSON exec_summary should carry a count")
	}
}

func TestSummarizeExecutions_SkippedNotCountedAsFailed(t *testing.T) {
	execs := []executor.ExecutionResult{
		{ScenarioID: "a", Passed: true},
		{ScenarioID: "b", Passed: false},                // real failure
		{ScenarioID: "c", Passed: false, Skipped: true}, // infra — not a failure
		{ScenarioID: "d", Passed: false, Skipped: true}, // infra — not a failure
	}
	s := reporter.SummarizeExecutions(execs, "http://localhost:8080")
	if s.Passed != 1 {
		t.Errorf("passed = %d, want 1", s.Passed)
	}
	if s.Failed != 1 {
		t.Errorf("failed = %d, want 1 (skipped must not count as failed)", s.Failed)
	}
	if s.Skipped != 2 {
		t.Errorf("skipped = %d, want 2", s.Skipped)
	}
	if s.Count != 4 {
		t.Errorf("count = %d, want 4", s.Count)
	}
}

func TestExitCodeFor(t *testing.T) {
	// No failures, high confidence -> 0
	result := reporter.RunResult{ConfidenceScore: 0.90}
	if code := reporter.ExitCodeFor(result, 0.80); code != 0 {
		t.Errorf("expected 0, got %d", code)
	}

	// Low confidence -> 1
	result.ConfidenceScore = 0.60
	if code := reporter.ExitCodeFor(result, 0.80); code != 1 {
		t.Errorf("expected 1, got %d", code)
	}

	// Critical failure -> 2
	result.Failures = []reporter.Failure{{Priority: "critical"}}
	if code := reporter.ExitCodeFor(result, 0.80); code != 2 {
		t.Errorf("expected 2, got %d", code)
	}
}

func TestSaveMarkdownReport(t *testing.T) {
	tmp := t.TempDir()
	result := makeResult()
	path, err := reporter.SaveMarkdownReport(tmp, result)
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
	// File should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("report file should exist")
	}
}
