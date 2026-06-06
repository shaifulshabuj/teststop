package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shaifulshabuj/teststop/internal/ai"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// AIExecutor executes prose scenarios by handing them back to the AI CLI, which
// actually performs the steps against Target and reports a structured verdict.
// Used when a live --target is set but the scenario has no structured exec block.
type AIExecutor struct {
	Adapter ai.AIAdapter
	Target  string
}

// verdict is the small JSON contract the AI returns from an execution prompt.
type verdict struct {
	Passed         bool   `json:"passed"`
	ActualBehavior string `json:"actual_behavior"`
	FailureReason  string `json:"failure_reason"`
}

// Execute prompts the AI to run the scenario and parses its verdict.
func (e *AIExecutor) Execute(_ context.Context, s scenario.Scenario) ExecutionResult {
	start := time.Now()
	res := ExecutionResult{
		ScenarioID: s.ScenarioID,
		Area:       s.ConfidenceArea,
		Priority:   s.Priority,
		Mode:       ModeAI,
	}

	out, err := e.Adapter.Prompt(e.buildPrompt(s))
	res.Duration = time.Since(start)
	if err != nil {
		res.Passed = false
		res.ActualBehavior = "AI execution failed to run"
		res.FailureReason = err.Error()
		return res
	}

	v, err := parseVerdict(out)
	if err != nil {
		res.Passed = false
		res.ActualBehavior = "AI returned an unparseable verdict"
		res.FailureReason = err.Error()
		return res
	}

	res.Passed = v.Passed
	res.ActualBehavior = v.ActualBehavior
	if !v.Passed {
		res.FailureReason = v.FailureReason
	}
	return res
}

// buildPrompt composes the execution instruction sent to the AI.
func (e *AIExecutor) buildPrompt(s scenario.Scenario) string {
	var b strings.Builder
	b.WriteString("You are executing a single adversarial test scenario against a RUNNING system.\n")
	b.WriteString("Target base URL: ")
	b.WriteString(e.Target)
	b.WriteString("\n\nActually perform the steps below against the target, then judge the outcome.\n\n")
	b.WriteString("Scenario: ")
	b.WriteString(s.Title)
	b.WriteString("\nUser perspective: ")
	b.WriteString(s.UserPerspective)
	if len(s.Preconditions) > 0 {
		b.WriteString("\nPreconditions: ")
		b.WriteString(strings.Join(s.Preconditions, "; "))
	}
	b.WriteString("\nSteps:\n")
	for i, step := range s.Steps {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
	}
	if len(s.ChaosFactors) > 0 {
		b.WriteString("Chaos factors to apply: ")
		b.WriteString(strings.Join(s.ChaosFactors, "; "))
		b.WriteString("\n")
	}
	b.WriteString("Expected behavior: ")
	b.WriteString(s.ExpectedBehavior)
	b.WriteString("\n\nReply with ONLY a JSON object, no prose, in this exact shape:\n")
	b.WriteString(`{"passed": true|false, "actual_behavior": "what actually happened", "failure_reason": "why it failed, or empty"}`)
	b.WriteString("\nThe system passes only if it behaves safely and as expected under the steps and chaos factors.\n")
	return b.String()
}

// parseVerdict tolerantly parses the AI's verdict JSON, stripping code fences
// and any surrounding prose (extracting the outermost {...} object).
func parseVerdict(data []byte) (verdict, error) {
	s := strings.TrimSpace(string(data))
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// Extract the outermost JSON object if the AI wrapped it in prose.
	if start := strings.Index(s, "{"); start >= 0 {
		if end := strings.LastIndex(s, "}"); end > start {
			s = s[start : end+1]
		}
	}

	var v verdict
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return verdict{}, fmt.Errorf("parse verdict JSON: %w", err)
	}
	return v, nil
}
