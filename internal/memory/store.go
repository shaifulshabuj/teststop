// Package memory persists what teststop has learned about a project so
// each run can spend its scenario budget on what is new or volatile and
// leave alone what is already proven.
//
// Memory lives at .teststop/memory.json at the project root and is meant
// to be committed to version control: the confidence in a system is part
// of the system. The directory is created lazily on the first write.
package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Dir is the on-disk directory name teststop writes to at the project root.
const Dir = ".teststop"

// MemoryFile is the JSON file inside Dir that holds accumulated confidence.
const MemoryFile = "memory.json"

// RetiredFile records areas teststop has stopped testing aggressively.
// It is separate from MemoryFile so a manual reset of one does not lose
// the other.
const RetiredFile = "retired.json"

// Maturity stage labels. Promoted from raw confidence in classify.go.
const (
	StageNew     = "new"
	StageGrowing = "growing"
	StageMature  = "mature"
	StageLegacy  = "legacy"
)

// Status labels per area.
const (
	StatusNew      = "new"
	StatusVolatile = "volatile"
	StatusStable   = "stable"
	StatusRetired  = "retired"
)

// AreaConfidence is the confidence record for a single subsystem of the
// project under test. The key in Memory.SystemAreas is a stable string
// the AI is asked to reuse across runs (the scenario's confidence_area).
type AreaConfidence struct {
	Confidence float64   `json:"confidence"`
	LastTested time.Time `json:"last_tested"`
	TestCount  int       `json:"test_count"`
	PassCount  int       `json:"pass_count"`
	FailCount  int       `json:"fail_count"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes,omitempty"`
}

// Memory is the full persisted state for one project.
type Memory struct {
	SystemAreas       map[string]AreaConfidence `json:"system_areas"`
	OverallConfidence float64                   `json:"overall_confidence"`
	MaturityStage     string                    `json:"maturity_stage"`
	LastRun           time.Time                 `json:"last_run"`
	TotalRuns         int                       `json:"total_runs"`
	Retired           []RetiredArea             `json:"retired,omitempty"`
}

// RetiredArea records an area that has crossed the retirement threshold
// and is no longer tested aggressively. We keep them so they can be
// resurrected by an explicit reset or a confidence drop.
type RetiredArea struct {
	Area       string    `json:"area"`
	RetiredAt  time.Time `json:"retired_at"`
	Confidence float64   `json:"confidence"`
	TestCount  int       `json:"test_count"`
}

// New returns an empty Memory suitable for a first run.
func New() *Memory {
	return &Memory{
		SystemAreas:   map[string]AreaConfidence{},
		MaturityStage: StageNew,
	}
}

// Load reads memory for the project at root. A missing file is not an
// error: the first run on a project has no prior state.
func Load(root string) (*Memory, error) {
	path := filepath.Join(root, Dir, MemoryFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return New(), nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	m := New()
	if err := json.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.SystemAreas == nil {
		m.SystemAreas = map[string]AreaConfidence{}
	}
	return m, nil
}

// Save writes the memory atomically: temp file then rename, so a crash
// mid-write never leaves a corrupted memory.json behind.
func Save(root string, m *Memory) error {
	dir := filepath.Join(root, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encode memory: %w", err)
	}
	target := filepath.Join(dir, MemoryFile)
	tmp, err := os.CreateTemp(dir, "memory-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, target); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s: %w", target, err)
	}
	return nil
}

// Reset deletes the memory file. The .teststop directory itself is kept
// so a committed-but-empty directory survives.
func Reset(root string) error {
	for _, name := range []string{MemoryFile, RetiredFile} {
		path := filepath.Join(root, Dir, name)
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
	}
	return nil
}
