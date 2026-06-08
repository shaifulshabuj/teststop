package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// fakeAdapter is a test double for ai.AIAdapter.
type fakeAdapter struct {
	out      []byte
	err      error
	lastSeen string
}

func (f *fakeAdapter) GenerateScenarios(string) ([]scenario.Scenario, error) { return nil, nil }
func (f *fakeAdapter) Name() string                                          { return "fake" }
func (f *fakeAdapter) Prompt(input string) ([]byte, error) {
	f.lastSeen = input
	return f.out, f.err
}

func aiScenario() scenario.Scenario {
	return scenario.Scenario{
		ScenarioID:       "s1",
		ConfidenceArea:   "checkout",
		Priority:         scenario.PriorityCritical,
		Title:            "double submit",
		Steps:            []string{"click pay twice"},
		ExpectedBehavior: "charged once",
	}
}

func TestAIExecutor_ParsesVerdictPass(t *testing.T) {
	fa := &fakeAdapter{out: []byte(`{"passed": true, "actual_behavior": "charged once", "failure_reason": ""}`)}
	ex := &AIExecutor{Adapter: fa, Target: "http://localhost:9999"}
	res := ex.Execute(context.Background(), aiScenario())

	if !res.Passed {
		t.Fatalf("expected pass, got: %s", res.FailureReason)
	}
	if res.Mode != ModeAI {
		t.Errorf("mode = %q, want %q", res.Mode, ModeAI)
	}
	if res.ActualBehavior != "charged once" {
		t.Errorf("actual = %q", res.ActualBehavior)
	}
}

func TestAIExecutor_ParsesVerdictFailWithFences(t *testing.T) {
	fa := &fakeAdapter{out: []byte("Here is the result:\n```json\n{\"passed\": false, \"actual_behavior\": \"charged twice\", \"failure_reason\": \"no idempotency\"}\n```\n")}
	ex := &AIExecutor{Adapter: fa, Target: "http://localhost:9999"}
	res := ex.Execute(context.Background(), aiScenario())

	if res.Passed {
		t.Fatal("expected fail, got pass")
	}
	if res.FailureReason != "no idempotency" {
		t.Errorf("failure reason = %q", res.FailureReason)
	}
}

func TestAIExecutor_AdapterErrorIsSkipped(t *testing.T) {
	// An AI CLI error (e.g. exit 1 on rate-limit exhaustion) is infrastructure,
	// not a verdict — it must be marked skipped, not failed.
	fa := &fakeAdapter{err: errors.New("claude: exit status 1\nstderr: ")}
	ex := &AIExecutor{Adapter: fa, Target: "http://localhost:9999"}
	res := ex.Execute(context.Background(), aiScenario())

	if !res.Skipped {
		t.Fatal("adapter error should be marked Skipped")
	}
	if res.Passed {
		t.Error("skipped result should not be passed")
	}
	if res.FailureReason == "" {
		t.Error("expected failure reason carried through for debugging")
	}
}

func TestAIExecutor_UnparseableVerdictIsSkipped(t *testing.T) {
	fa := &fakeAdapter{out: []byte("the system seems fine to me")}
	ex := &AIExecutor{Adapter: fa, Target: "http://localhost:9999"}
	res := ex.Execute(context.Background(), aiScenario())

	if !res.Skipped {
		t.Fatal("unparseable verdict should be marked Skipped, not failed")
	}
	if res.Passed {
		t.Error("skipped result should not be passed")
	}
}

func TestAIExecutor_PromptIncludesTarget(t *testing.T) {
	fa := &fakeAdapter{out: []byte(`{"passed": true}`)}
	ex := &AIExecutor{Adapter: fa, Target: "http://localhost:1234"}
	_ = ex.Execute(context.Background(), aiScenario())

	if fa.lastSeen == "" {
		t.Fatal("adapter was not prompted")
	}
	if !contains(fa.lastSeen, "http://localhost:1234") {
		t.Error("prompt did not include target URL")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
