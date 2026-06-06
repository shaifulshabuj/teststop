package executor

import (
	"context"
	"strings"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// StaticExecutor performs structural validation only — no live system is
// contacted. This preserves v0.1 semantics: a well-formed scenario is treated
// as a pass. It is the fallback when no --target is provided.
type StaticExecutor struct{}

// Execute validates that the scenario is well-formed.
func (e *StaticExecutor) Execute(_ context.Context, s scenario.Scenario) ExecutionResult {
	res := ExecutionResult{
		ScenarioID: s.ScenarioID,
		Area:       s.ConfidenceArea,
		Priority:   s.Priority,
		Mode:       ModeStatic,
	}

	var missing []string
	if len(s.Steps) == 0 {
		missing = append(missing, "steps")
	}
	if strings.TrimSpace(s.ExpectedBehavior) == "" {
		missing = append(missing, "expected_behavior")
	}
	if strings.TrimSpace(s.ConfidenceArea) == "" {
		missing = append(missing, "confidence_area")
	}

	if len(missing) > 0 {
		res.Passed = false
		res.ActualBehavior = "scenario is malformed"
		res.FailureReason = "missing required fields: " + strings.Join(missing, ", ")
		return res
	}

	res.Passed = true
	res.ActualBehavior = "structurally valid (not executed against a live target)"
	return res
}
