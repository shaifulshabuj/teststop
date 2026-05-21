package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

func sampleRun() Run {
	return Run{
		RunID:              "abc",
		Timestamp:          time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC),
		Project:            "demo",
		Language:           "Go",
		ProjectType:        "cli",
		OverallConfidence:  0.72,
		PreviousConfidence: 0.69,
		ConfidenceDelta:    0.03,
		MaturityStage:      "growing",
		Threshold:          0.85,
		ScenariosGenerated: 3,
		ScenariosPassed:    2,
		ScenariosFailed:    1,
		Failures: []Failure{
			{ScenarioID: "race-1", Title: "race", Priority: scenario.PriorityHigh, ConfidenceArea: "checkout"},
		},
		StableAreas:    []string{"auth"},
		VolatileAreas:  []string{"checkout"},
		RetiredThisRun: []string{"login"},
	}
}

func TestWriteJSON_RoundTrips(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, sampleRun()); err != nil {
		t.Fatalf("write json: %v", err)
	}
	var back Run
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Project != "demo" {
		t.Fatalf("project = %s", back.Project)
	}
	if len(back.Failures) != 1 {
		t.Fatalf("failures = %d", len(back.Failures))
	}
}

func TestWriteText_ContainsKeyFields(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, sampleRun()); err != nil {
		t.Fatalf("write text: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"demo", "Go", "cli", "Confidence:", "0.72", "race-1", "Retired this run"} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q\n%s", want, out)
		}
	}
}

func TestWriteMarkdown_GroupsByPriority(t *testing.T) {
	var buf bytes.Buffer
	r := sampleRun()
	r.Failures = append(r.Failures, Failure{
		ScenarioID: "boom", Title: "boom", Priority: scenario.PriorityCritical, ConfidenceArea: "billing",
	})
	if err := WriteMarkdown(&buf, r); err != nil {
		t.Fatalf("write md: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "### CRITICAL") {
		t.Errorf("expected CRITICAL section\n%s", out)
	}
	if !strings.Contains(out, "### HIGH") {
		t.Errorf("expected HIGH section\n%s", out)
	}
}

func TestExitCode(t *testing.T) {
	r := sampleRun()
	r.OverallConfidence = 0.9
	r.Threshold = 0.85
	r.Failures = nil
	if got := r.ExitCode(); got != 0 {
		t.Errorf("above threshold no failures = %d, want 0", got)
	}
	r.OverallConfidence = 0.5
	if got := r.ExitCode(); got != 1 {
		t.Errorf("below threshold = %d, want 1", got)
	}
	r.Failures = []Failure{{Priority: scenario.PriorityCritical}}
	if got := r.ExitCode(); got != 2 {
		t.Errorf("critical failure = %d, want 2", got)
	}
}
