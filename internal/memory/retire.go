package memory

import "time"

// Retire moves areas that have crossed the retirement threshold into
// the Retired list and tags their AreaConfidence with StatusRetired.
// It returns the slice of areas retired in this pass (may be empty).
//
// Retirement is sticky-but-not-permanent: a retired area still has its
// confidence carried in SystemAreas, and a future failure on that area
// drops the confidence back below the threshold via Apply, which will
// then re-mark it StatusVolatile. The Retired log is purely audit.
func (m *Memory) Retire(now time.Time) []string {
	retired := []string{}
	known := map[string]struct{}{}
	for _, r := range m.Retired {
		known[r.Area] = struct{}{}
	}

	for k, a := range m.SystemAreas {
		if a.Status == StatusRetired {
			continue
		}
		if a.Confidence < RetireThreshold || a.TestCount < MinTestsForRetirement {
			continue
		}
		a.Status = StatusRetired
		m.SystemAreas[k] = a
		if _, already := known[k]; already {
			continue
		}
		m.Retired = append(m.Retired, RetiredArea{
			Area:       k,
			RetiredAt:  now,
			Confidence: a.Confidence,
			TestCount:  a.TestCount,
		})
		retired = append(retired, k)
	}
	return retired
}

// Revive removes the StatusRetired tag from any area whose confidence
// has dropped below the retirement threshold (e.g. after a failure).
// It is the symmetric operation to Retire and should be called after
// Apply when confidence may have shifted.
func (m *Memory) Revive() []string {
	revived := []string{}
	for k, a := range m.SystemAreas {
		if a.Status != StatusRetired {
			continue
		}
		if a.Confidence >= RetireThreshold {
			continue
		}
		a.Status = StatusVolatile
		m.SystemAreas[k] = a
		revived = append(revived, k)
	}
	return revived
}
