package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RetiredArea records an area that has been retired.
type RetiredArea struct {
	Name       string    `json:"name"`
	RetiredAt  time.Time `json:"retired_at"`
	TestCount  int       `json:"test_count"`
	PassCount  int       `json:"pass_count"`
	Confidence float64   `json:"confidence"`
}

// RetirementRecord is the full retired.json structure.
type RetirementRecord struct {
	Areas     []RetiredArea `json:"areas"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// retiredPath returns the path to retired.json for the given project.
func retiredPath(projectPath string) string {
	return filepath.Join(projectPath, ".teststop", "retired.json")
}

// loadRetirementRecord reads the existing retired.json or returns an empty record.
func loadRetirementRecord(projectPath string) (*RetirementRecord, error) {
	data, err := os.ReadFile(retiredPath(projectPath))
	if err != nil {
		if os.IsNotExist(err) {
			return &RetirementRecord{}, nil
		}
		return nil, fmt.Errorf("reading retired.json: %w", err)
	}
	var rec RetirementRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("parsing retired.json: %w", err)
	}
	return &rec, nil
}

// RetireEligibleAreas moves areas with Confidence >= RetirementThreshold AND
// TestCount >= 15 from m.Areas to projectPath/.teststop/retired.json.
// The areas remain in m.Areas but are marked Retired=true so they are not reset.
// Returns the names of areas that were retired this call.
func (m *Memory) RetireEligibleAreas(projectPath string) ([]string, error) {
	var toRetire []*Area
	for _, area := range m.Areas {
		if !area.Retired && area.Confidence >= RetirementThreshold && area.TestCount >= 15 {
			toRetire = append(toRetire, area)
		}
	}
	if len(toRetire) == 0 {
		return nil, nil
	}

	rec, err := loadRetirementRecord(projectPath)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var names []string
	for _, area := range toRetire {
		area.Retired = true
		rec.Areas = append(rec.Areas, RetiredArea{
			Name:       area.Name,
			RetiredAt:  now,
			TestCount:  area.TestCount,
			PassCount:  area.PassCount,
			Confidence: area.Confidence,
		})
		names = append(names, area.Name)
	}
	rec.UpdatedAt = now

	// Ensure directory exists before writing.
	dir := filepath.Join(projectPath, ".teststop")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating .teststop directory: %w", err)
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling retired.json: %w", err)
	}
	if err := os.WriteFile(retiredPath(projectPath), data, 0o644); err != nil {
		return nil, fmt.Errorf("writing retired.json: %w", err)
	}

	m.UpdatedAt = now
	return names, nil
}
