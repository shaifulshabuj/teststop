// Package reporter turns a finished teststop run into output a human, an
// agent, or a CI gate can act on. Three formats are supported by design:
// JSON for machines, text for terminals, Markdown for PR comments.
package reporter

import (
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// Format is the output style requested by the caller.
type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// Failure is a single scenario that did not pass, surfaced in reports.
type Failure struct {
	ScenarioID     string            `json:"scenario_id"`
	Title          string            `json:"title"`
	Priority       scenario.Priority `json:"priority"`
	ConfidenceArea string            `json:"confidence_area"`
	Notes          string            `json:"notes,omitempty"`
}

// Run is the full payload a reporter renders. It is the contract between
// the wire-up in `teststop run` and every output format.
type Run struct {
	RunID              string             `json:"run_id"`
	Timestamp          time.Time          `json:"timestamp"`
	Project            string             `json:"project"`
	Language           string             `json:"language"`
	ProjectType        string             `json:"project_type"`
	OverallConfidence  float64            `json:"overall_confidence"`
	PreviousConfidence float64            `json:"previous_confidence"`
	ConfidenceDelta    float64            `json:"confidence_delta"`
	MaturityStage      string             `json:"maturity_stage"`
	ReadyForDeploy     bool               `json:"ready_for_deploy"`
	Threshold          float64            `json:"threshold"`
	ScenariosGenerated int                `json:"scenarios_generated"`
	ScenariosPassed    int                `json:"scenarios_passed"`
	ScenariosFailed    int                `json:"scenarios_failed"`
	ScenariosUnknown   int                `json:"scenarios_unknown"`
	Scenarios          []scenario.Scenario `json:"scenarios,omitempty"`
	Failures           []Failure          `json:"failures"`
	RetiredThisRun     []string           `json:"retired_this_run"`
	StableAreas        []string           `json:"stable_areas"`
	VolatileAreas      []string           `json:"volatile_areas"`
	Notes              []string           `json:"notes,omitempty"`
}

// ExitCode picks the conventional teststop exit code from a run.
//
//   0 = confidence threshold met
//   1 = below threshold, review needed
//   2 = critical failures, do not deploy
//
// Internal errors (exit 3) are reported by the caller, not here.
func (r Run) ExitCode() int {
	for _, f := range r.Failures {
		if f.Priority == scenario.PriorityCritical {
			return 2
		}
	}
	if r.OverallConfidence < r.Threshold {
		return 1
	}
	return 0
}
