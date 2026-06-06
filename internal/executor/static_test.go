package executor

import (
	"context"
	"testing"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

func TestStaticExecutor_WellFormedPasses(t *testing.T) {
	s := scenario.Scenario{
		ScenarioID:       "s1",
		ConfidenceArea:   "auth",
		Steps:            []string{"do a", "do b"},
		ExpectedBehavior: "it works",
	}
	res := (&StaticExecutor{}).Execute(context.Background(), s)
	if !res.Passed {
		t.Fatalf("expected pass, got: %s", res.FailureReason)
	}
	if res.Mode != ModeStatic {
		t.Errorf("mode = %q, want %q", res.Mode, ModeStatic)
	}
}

func TestStaticExecutor_MalformedFails(t *testing.T) {
	s := scenario.Scenario{ScenarioID: "s1"} // no steps, no expected, no area
	res := (&StaticExecutor{}).Execute(context.Background(), s)
	if res.Passed {
		t.Fatal("expected fail for malformed scenario, got pass")
	}
	if res.FailureReason == "" {
		t.Error("expected a failure reason listing missing fields")
	}
}
