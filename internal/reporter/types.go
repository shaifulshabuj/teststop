package reporter

import (
	"time"

	"github.com/shaifulshabuj/teststop/internal/executor"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// RunResult is the complete output of a teststop run.
type RunResult struct {
	ProjectName     string                     `json:"project_name"`
	ProjectPath     string                     `json:"project_path"`
	Language        string                     `json:"language"`
	SystemType      string                     `json:"system_type"`
	Timestamp       time.Time                  `json:"timestamp"`
	Duration        time.Duration              `json:"duration_ms"` // duration in ms for JSON
	Scenarios       []scenario.Scenario        `json:"scenarios"`
	Executions      []executor.ExecutionResult `json:"executions,omitempty"`
	ExecSummary     ExecSummary                `json:"exec_summary"`
	Failures        []Failure                  `json:"failures"`
	StableAreas     []string                   `json:"stable_areas"`
	VolatileAreas   []string                   `json:"volatile_areas"`
	RetiredAreas    []string                   `json:"retired_areas"`
	ExitCode        int                        `json:"exit_code"`
	ConfidenceScore float64                    `json:"confidence_score"`
	AdapterName     string                     `json:"adapter_name"`
	Depth           string                     `json:"depth"`
}

// ExecSummary aggregates execution outcomes for quick scanning.
type ExecSummary struct {
	Executed int `json:"executed"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	// Target is the live system URL executed against, or "" if static-only.
	Target string `json:"target,omitempty"`
}

// SummarizeExecutions tallies execution results into an ExecSummary.
func SummarizeExecutions(execs []executor.ExecutionResult, target string) ExecSummary {
	s := ExecSummary{Executed: len(execs), Target: target}
	for _, e := range execs {
		if e.Passed {
			s.Passed++
		} else {
			s.Failed++
		}
	}
	return s
}

// Failure records a scenario that failed to meet expectations.
type Failure struct {
	ScenarioID  string `json:"scenario_id"`
	Title       string `json:"title"`
	Area        string `json:"area"`
	Priority    string `json:"priority"`
	Description string `json:"description"`
}

// ExitCodeFor returns the appropriate exit code for the run result.
// 0 = all good (confidence >= threshold)
// 1 = review needed (some failures, but not critical)
// 2 = critical failures (critical-priority failures found)
// 3 = teststop internal error (set by caller before calling reporter)
func ExitCodeFor(result RunResult, threshold float64) int {
	// Check for critical failures first
	for _, f := range result.Failures {
		if f.Priority == "critical" {
			return 2
		}
	}
	// Check confidence threshold
	if result.ConfidenceScore < threshold {
		return 1
	}
	return 0
}
