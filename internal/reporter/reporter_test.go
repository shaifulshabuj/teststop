package reporter_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

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
