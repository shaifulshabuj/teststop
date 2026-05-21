package memory

import (
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

func TestApply_PassNudgesUp(t *testing.T) {
	m := New()
	outcome := RunOutcome{
		When:           time.Now(),
		AreaByScenario: map[string]string{"s1": "auth"},
		Results: []scenario.Result{
			{ScenarioID: "s1", Passed: true},
		},
	}
	m.Apply(outcome)
	if got := m.SystemAreas["auth"].Confidence; got <= 0 {
		t.Fatalf("confidence should rise from 0 after a pass, got %v", got)
	}
	if got := m.SystemAreas["auth"].TestCount; got != 1 {
		t.Fatalf("test count = %d, want 1", got)
	}
}

func TestApply_FailHalvesConfidence(t *testing.T) {
	m := New()
	m.SystemAreas["auth"] = AreaConfidence{Confidence: 0.8, TestCount: 10}
	outcome := RunOutcome{
		When:           time.Now(),
		AreaByScenario: map[string]string{"s1": "auth"},
		Results: []scenario.Result{
			{ScenarioID: "s1", Passed: false},
		},
	}
	m.Apply(outcome)
	if got := m.SystemAreas["auth"].Confidence; got > 0.41 {
		t.Fatalf("confidence after fail = %v, want ~0.40", got)
	}
}

func TestApply_DecayOnUntouched(t *testing.T) {
	m := New()
	m.SystemAreas["billing"] = AreaConfidence{Confidence: 0.5, TestCount: 5, Status: StatusNew}
	outcome := RunOutcome{
		When:           time.Now(),
		AreaByScenario: map[string]string{"s1": "auth"},
		Results: []scenario.Result{
			{ScenarioID: "s1", Passed: true},
		},
	}
	m.Apply(outcome)
	if got := m.SystemAreas["billing"].Confidence; got >= 0.5 {
		t.Fatalf("untouched area should decay, got %v", got)
	}
}

func TestRetire_NeedsMinimumTests(t *testing.T) {
	m := New()
	m.SystemAreas["auth"] = AreaConfidence{Confidence: 0.99, TestCount: 3}
	if got := m.Retire(time.Now()); len(got) != 0 {
		t.Fatalf("should not retire with only 3 tests, got %v", got)
	}
	m.SystemAreas["auth"] = AreaConfidence{Confidence: 0.99, TestCount: MinTestsForRetirement}
	if got := m.Retire(time.Now()); len(got) != 1 || got[0] != "auth" {
		t.Fatalf("should retire after enough tests, got %v", got)
	}
}

func TestRetire_ReviveOnDrop(t *testing.T) {
	m := New()
	m.SystemAreas["auth"] = AreaConfidence{Confidence: 0.99, TestCount: 20, Status: StatusStable}
	m.Retire(time.Now())
	// Simulate a failure dragging confidence below threshold.
	a := m.SystemAreas["auth"]
	a.Confidence = 0.3
	m.SystemAreas["auth"] = a
	revived := m.Revive()
	if len(revived) != 1 || revived[0] != "auth" {
		t.Fatalf("expected revive([auth]), got %v", revived)
	}
	if got := m.SystemAreas["auth"].Status; got == StatusRetired {
		t.Fatalf("status should not be retired after revive, got %v", got)
	}
}

func TestNudgeUp_DiminishingReturns(t *testing.T) {
	c := 0.0
	for i := 0; i < 50; i++ {
		c = nudgeUp(c, PassWeight)
	}
	if c >= 1.0 {
		t.Fatalf("nudgeUp should not exceed 1.0, got %v", c)
	}
	if c < 0.9 {
		t.Fatalf("50 passes should comfortably exceed 0.9, got %v", c)
	}
}
