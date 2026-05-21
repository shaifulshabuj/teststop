package memory

import (
	"math"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// Confidence-scoring constants. These are deliberately conservative
// because the cost of retiring a test that still finds bugs is much
// higher than the cost of running a test we have proven stable.
const (
	// PassWeight is how much a single passing scenario nudges an
	// area's confidence toward 1.0. With 0.19 it takes roughly 15-16
	// consecutive passes from cold to reach the retirement threshold.
	PassWeight = 0.19

	// FailPenalty is the multiplicative damage a failing scenario
	// does. A single failure cuts confidence in half. Reality > history.
	FailPenalty = 0.5

	// RetireThreshold is the confidence at which an area is considered
	// proven stable and tested less aggressively next run.
	RetireThreshold = 0.95

	// VolatileThreshold is the confidence below which an area is
	// considered untrusted and gets extra scenario budget.
	VolatileThreshold = 0.60

	// MinTestsForRetirement guards against retiring an area after a
	// single lucky pass. Confidence alone is not enough; the area must
	// have been exercised repeatedly.
	MinTestsForRetirement = 8
)

// RunOutcome captures everything memory needs from a single run to
// update confidence: which areas were tested and which scenarios passed
// or failed.
type RunOutcome struct {
	When    time.Time
	Results []scenario.Result
	// AreaByScenario maps scenario_id → confidence_area, so memory
	// does not need to re-parse scenarios it never saw.
	AreaByScenario map[string]string
}

// Apply folds a run's outcome into memory, updating per-area confidence,
// overall confidence, and maturity stage.
//
// The update is intentionally simple: passes add a constant nudge,
// failures multiply by FailPenalty, and unseen areas decay slightly so
// confidence ages out of an area we have stopped exercising.
func (m *Memory) Apply(outcome RunOutcome) {
	if m.SystemAreas == nil {
		m.SystemAreas = map[string]AreaConfidence{}
	}

	touched := map[string]struct{}{}
	for _, r := range outcome.Results {
		area := outcome.AreaByScenario[r.ScenarioID]
		if area == "" {
			// A scenario without an area is a memory miss; we count it
			// against overall confidence below but cannot bucket it.
			continue
		}
		touched[area] = struct{}{}

		a := m.SystemAreas[area]
		a.TestCount++
		a.LastTested = outcome.When
		if r.Passed {
			a.PassCount++
			a.Confidence = nudgeUp(a.Confidence, PassWeight)
		} else {
			a.FailCount++
			a.Confidence = nudgeDown(a.Confidence)
		}
		a.Status = classifyArea(a)
		m.SystemAreas[area] = a
	}

	// Light decay on areas we did not exercise this run. Confidence is
	// a recency-weighted belief: if we are not looking, we cannot keep
	// claiming certainty forever.
	for k, a := range m.SystemAreas {
		if _, was := touched[k]; was {
			continue
		}
		a.Confidence = decay(a.Confidence)
		if a.Confidence < VolatileThreshold && a.Status == StatusStable {
			a.Status = StatusVolatile
		}
		m.SystemAreas[k] = a
	}

	m.TotalRuns++
	m.LastRun = outcome.When
	m.OverallConfidence = computeOverall(m.SystemAreas)
	m.MaturityStage = classifyStage(m)
}

// nudgeUp moves confidence toward 1.0 with diminishing returns. The
// closer to 1.0 we are, the harder it is to gain more.
func nudgeUp(c, weight float64) float64 {
	if c < 0 {
		c = 0
	}
	gap := 1.0 - c
	c += gap * weight
	if c > 1.0 {
		c = 1.0
	}
	return c
}

// nudgeDown applies the failure penalty. We never go to zero — a single
// fluke does not erase years of evidence — but we drop sharply.
func nudgeDown(c float64) float64 {
	if c < 0 {
		c = 0
	}
	c *= FailPenalty
	if c < 0.05 {
		c = 0.05
	}
	return c
}

// decay is the per-run drift applied to untouched areas.
func decay(c float64) float64 {
	if c <= 0 {
		return 0
	}
	c -= 0.01
	if c < 0 {
		return 0
	}
	return c
}

func classifyArea(a AreaConfidence) string {
	switch {
	case a.TestCount < 2:
		return StatusNew
	case a.Confidence >= RetireThreshold && a.TestCount >= MinTestsForRetirement:
		return StatusStable
	case a.Confidence < VolatileThreshold:
		return StatusVolatile
	default:
		return StatusNew
	}
}

func computeOverall(areas map[string]AreaConfidence) float64 {
	if len(areas) == 0 {
		return 0
	}
	total := 0.0
	for _, a := range areas {
		total += a.Confidence
	}
	avg := total / float64(len(areas))
	// Round to two decimals so the report stays readable and stable.
	return math.Round(avg*100) / 100
}

func classifyStage(m *Memory) string {
	switch {
	case m.TotalRuns < 3:
		return StageNew
	case m.OverallConfidence >= 0.90 && m.TotalRuns >= 10:
		return StageMature
	case m.OverallConfidence >= 0.95 && m.TotalRuns >= 25:
		return StageLegacy
	default:
		return StageGrowing
	}
}

// StableAreas returns areas at or above the retirement threshold,
// suitable for the mandate's "already proven" section.
func (m *Memory) StableAreas() []string {
	out := []string{}
	for k, a := range m.SystemAreas {
		if a.Confidence >= RetireThreshold && a.TestCount >= MinTestsForRetirement {
			out = append(out, k)
		}
	}
	return out
}

// VolatileAreas returns areas below the volatile threshold or that have
// failed recently, suitable for the mandate's "focus areas" section.
func (m *Memory) VolatileAreas() []string {
	out := []string{}
	for k, a := range m.SystemAreas {
		if a.Confidence < VolatileThreshold || a.Status == StatusVolatile || a.Status == StatusNew {
			out = append(out, k)
		}
	}
	return out
}
