package memory

import "time"

// GetOrCreate returns the Area for areaName, creating it if it does not exist.
func (m *Memory) GetOrCreate(areaName string) *Area {
	if area, ok := m.Areas[areaName]; ok {
		return area
	}
	area := &Area{
		Name:          areaName,
		Confidence:    0.0,
		MaturityStage: MaturityNew,
	}
	m.Areas[areaName] = area
	return area
}

// UpdateArea records the result of testing an area.
// pass=true increases confidence using exponential approach; pass=false decreases it.
func (m *Memory) UpdateArea(areaName string, pass bool) {
	area := m.GetOrCreate(areaName)
	area.TestCount++
	area.LastTestedAt = time.Now()

	if pass {
		area.PassCount++
		// Exponential approach: new = old + PassWeight * (1.0 - old)
		area.Confidence += PassWeight * (1.0 - area.Confidence)
		if area.Confidence > 1.0 {
			area.Confidence = 1.0
		}
	} else {
		area.FailCount++
		area.Confidence -= FailPenalty
		if area.Confidence < 0.0 {
			area.Confidence = 0.0
		}
	}

	area.MaturityStage = maturityFor(area.Confidence)
	m.UpdatedAt = time.Now()
}

// GetStableAreas returns all non-retired areas with confidence >= StableThreshold.
func (m *Memory) GetStableAreas() []*Area {
	var out []*Area
	for _, area := range m.Areas {
		if !area.Retired && area.Confidence >= StableThreshold {
			out = append(out, area)
		}
	}
	return out
}

// GetVolatileAreas returns all non-retired areas with confidence < VolatileThreshold.
func (m *Memory) GetVolatileAreas() []*Area {
	var out []*Area
	for _, area := range m.Areas {
		if !area.Retired && area.Confidence < VolatileThreshold {
			out = append(out, area)
		}
	}
	return out
}

// maturityFor returns the maturity stage string for the given confidence value.
func maturityFor(c float64) string {
	switch {
	case c < 0.40:
		return MaturityNew
	case c < 0.70:
		return MaturityGrowing
	case c < 0.90:
		return MaturityMature
	default:
		return MaturityLegacy
	}
}
