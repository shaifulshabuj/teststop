package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/internal/executor"
	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
	"github.com/shaifulshabuj/teststop/internal/reporter"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// baseResult returns a minimal RunResult suitable for output tests.
func baseResult() reporter.RunResult {
	return reporter.RunResult{
		ProjectName:     "test-proj",
		Language:        "Go",
		SystemType:      "cli",
		ConfidenceScore: 0.9,
		Scenarios: []scenario.Scenario{
			{ScenarioID: "s1", Title: "Auth bypass", Priority: "low", ConfidenceArea: "auth"},
		},
	}
}

func TestWriteOutput_JSON(t *testing.T) {
	var buf bytes.Buffer
	if err := writeOutput(&buf, baseResult(), "json", false, false); err != nil {
		t.Fatalf("writeOutput json: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output not valid JSON: %v\ngot: %s", err, buf.String())
	}
	if out["project_name"] != "test-proj" {
		t.Errorf("project_name: got %v", out["project_name"])
	}
}

func TestWriteOutput_Text(t *testing.T) {
	var buf bytes.Buffer
	if err := writeOutput(&buf, baseResult(), "text", false, false); err != nil {
		t.Fatalf("writeOutput text: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("text output should be non-empty")
	}
}

func TestWriteOutput_TextNoColor(t *testing.T) {
	var buf bytes.Buffer
	if err := writeOutput(&buf, baseResult(), "text", true, false); err != nil {
		t.Fatalf("writeOutput text+no-color: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("text no-color output should be non-empty")
	}
}

func TestWriteOutput_Markdown(t *testing.T) {
	var buf bytes.Buffer
	if err := writeOutput(&buf, baseResult(), "markdown", false, false); err != nil {
		t.Fatalf("writeOutput markdown: %v", err)
	}
	if !strings.Contains(buf.String(), "#") {
		t.Error("markdown output should contain a '#' header")
	}
}

func TestWriteOutput_Quiet(t *testing.T) {
	cases := []struct {
		exitCode int
		want     string
	}{
		{0, "OK"},
		{1, "REVIEW"},
		{2, "CRITICAL"},
		{3, "ERROR"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			r := baseResult()
			r.ExitCode = tc.exitCode
			var buf bytes.Buffer
			if err := writeOutput(&buf, r, "text", false, true); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("quiet exit %d: want %q in %q", tc.exitCode, tc.want, buf.String())
			}
		})
	}
}

func TestBuildRunResult(t *testing.T) {
	ctx := reader.ProjectContext{
		Name:     "myproject",
		Path:     "/tmp/myproject",
		Language: "Go",
		Type:     "cli",
	}
	scenarios := []scenario.Scenario{
		{ScenarioID: "s1", Title: "Test auth", ConfidenceArea: "auth", Priority: "high"},
		{ScenarioID: "s2", Title: "Test input", ConfidenceArea: "input", Priority: "low"},
	}
	executions := []executor.ExecutionResult{
		{ScenarioID: "s1", Area: "auth", Passed: false, Priority: "high", FailureReason: "auth bypass possible"},
		{ScenarioID: "s2", Area: "input", Passed: true},
	}
	mem := memory.NewMemory()
	start := time.Now()

	result := buildRunResult(ctx, scenarios, executions, mem, "claude", "normal", "", start)

	if result.ProjectName != "myproject" {
		t.Errorf("ProjectName: want %q, got %q", "myproject", result.ProjectName)
	}
	if result.Language != "Go" {
		t.Errorf("Language: want %q, got %q", "Go", result.Language)
	}
	if len(result.Scenarios) != 2 {
		t.Errorf("Scenarios: want 2, got %d", len(result.Scenarios))
	}
	if len(result.Failures) != 1 {
		t.Errorf("Failures: want 1, got %d", len(result.Failures))
	}
	if result.Failures[0].Title != "Test auth" {
		t.Errorf("Failures[0].Title: want %q, got %q", "Test auth", result.Failures[0].Title)
	}
	if result.Failures[0].Description != "auth bypass possible" {
		t.Errorf("Failures[0].Description: want %q, got %q", "auth bypass possible", result.Failures[0].Description)
	}
	if result.AdapterName != "claude" {
		t.Errorf("AdapterName: want %q, got %q", "claude", result.AdapterName)
	}
	if result.Depth != "normal" {
		t.Errorf("Depth: want %q, got %q", "normal", result.Depth)
	}
	if result.Duration <= 0 {
		t.Error("Duration should be > 0")
	}
}

func TestBuildRunResult_SkippedNotFailure(t *testing.T) {
	ctx := reader.ProjectContext{Name: "p", Path: "/p", Language: "Go", Type: "cli"}
	scenarios := []scenario.Scenario{
		{ScenarioID: "s1", ConfidenceArea: "auth"},
	}
	// Skipped execution: must not appear in failures.
	executions := []executor.ExecutionResult{
		{ScenarioID: "s1", Skipped: true, Area: "auth"},
	}
	mem := memory.NewMemory()

	result := buildRunResult(ctx, scenarios, executions, mem, "claude", "normal", "", time.Now())

	if len(result.Failures) != 0 {
		t.Errorf("skipped execution must not create a failure entry, got %d", len(result.Failures))
	}
}

func TestBuildRunResult_ConfidenceAverage(t *testing.T) {
	ctx := reader.ProjectContext{Name: "p", Path: "/p", Language: "Go", Type: "cli"}
	mem := memory.NewMemory()
	mem.UpdateArea("auth", true)   // bumps confidence > 0
	mem.UpdateArea("input", false) // drops confidence

	result := buildRunResult(ctx, nil, nil, mem, "fake", "light", "", time.Now())

	// Average should be between 0 and 1 (and non-zero since one area improved).
	if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
		t.Errorf("ConfidenceScore out of range [0,1]: %f", result.ConfidenceScore)
	}
}

// TestApplyRunEnv_OutputAndTarget covers the OUTPUT and TARGET env paths in
// applyRunEnv that the precedence tests don't reach.
func TestApplyRunEnv_OutputAndTarget(t *testing.T) {
	t.Run("output env overrides default", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_OUTPUT", "json")
		runOutput = "text"
		if err := applyRunEnv(func(string) bool { return false }); err != nil {
			t.Fatal(err)
		}
		if runOutput != "json" {
			t.Errorf("TESTSTOP_RUN_OUTPUT: want %q, got %q", "json", runOutput)
		}
	})

	t.Run("target env sets value", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_TARGET", "http://localhost:9999")
		runTarget = ""
		if err := applyRunEnv(func(string) bool { return false }); err != nil {
			t.Fatal(err)
		}
		if runTarget != "http://localhost:9999" {
			t.Errorf("TESTSTOP_RUN_TARGET: want %q, got %q", "http://localhost:9999", runTarget)
		}
	})
}

// TestApplyRunEnv_BoolAndDuration exercises paths in applyRunEnv that the
// precedence tests don't reach: bool and duration env vars and their error paths.
func TestApplyRunEnv_BoolAndDuration(t *testing.T) {
	t.Run("no-color env true", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_NO_COLOR", "true")
		runNoColor = false
		if err := applyRunEnv(func(string) bool { return false }); err != nil {
			t.Fatal(err)
		}
		if !runNoColor {
			t.Error("TESTSTOP_RUN_NO_COLOR=true should set runNoColor=true")
		}
	})

	t.Run("quiet env true", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_QUIET", "true")
		runQuiet = false
		if err := applyRunEnv(func(string) bool { return false }); err != nil {
			t.Fatal(err)
		}
		if !runQuiet {
			t.Error("TESTSTOP_RUN_QUIET=true should set runQuiet=true")
		}
	})

	t.Run("exec-timeout env valid", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_EXEC_TIMEOUT", "30s")
		runExecTimeout = 10 * time.Second
		if err := applyRunEnv(func(string) bool { return false }); err != nil {
			t.Fatal(err)
		}
		if runExecTimeout != 30*time.Second {
			t.Errorf("exec-timeout: want 30s, got %v", runExecTimeout)
		}
	})

	t.Run("exec-timeout env malformed", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_EXEC_TIMEOUT", "not-a-duration")
		if err := applyRunEnv(func(string) bool { return false }); err == nil {
			t.Error("expected error for malformed TESTSTOP_RUN_EXEC_TIMEOUT")
		}
	})

	t.Run("no-color env malformed", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_NO_COLOR", "notabool")
		if err := applyRunEnv(func(string) bool { return false }); err == nil {
			t.Error("expected error for malformed TESTSTOP_RUN_NO_COLOR")
		}
	})

	t.Run("quiet env malformed", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_QUIET", "notabool")
		if err := applyRunEnv(func(string) bool { return false }); err == nil {
			t.Error("expected error for malformed TESTSTOP_RUN_QUIET")
		}
	})

	t.Run("max-retries env valid", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_MAX_RETRIES", "5")
		runMaxRetries = 2
		if err := applyRunEnv(func(string) bool { return false }); err != nil {
			t.Fatal(err)
		}
		if runMaxRetries != 5 {
			t.Errorf("max-retries: want 5, got %d", runMaxRetries)
		}
	})

	t.Run("max-retries env malformed", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_MAX_RETRIES", "bad")
		if err := applyRunEnv(func(string) bool { return false }); err == nil {
			t.Error("expected error for malformed TESTSTOP_RUN_MAX_RETRIES")
		}
	})

	t.Run("concurrency env malformed", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_CONCURRENCY", "bad")
		if err := applyRunEnv(func(string) bool { return false }); err == nil {
			t.Error("expected error for malformed TESTSTOP_RUN_CONCURRENCY")
		}
	})
}
