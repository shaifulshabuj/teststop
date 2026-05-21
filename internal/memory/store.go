package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	RetirementThreshold = 0.95
	PassWeight          = 0.19
	FailPenalty         = 0.30
	VolatileThreshold   = 0.75
	StableThreshold     = 0.95

	MaturityNew     = "new"     // confidence < 0.40
	MaturityGrowing = "growing" // 0.40 <= confidence < 0.70
	MaturityMature  = "mature"  // 0.70 <= confidence < 0.90
	MaturityLegacy  = "legacy"  // confidence >= 0.90 (includes retired)
)

// Area holds memory for one system area.
type Area struct {
	Name          string    `json:"name"`
	Confidence    float64   `json:"confidence"`
	TestCount     int       `json:"test_count"`
	PassCount     int       `json:"pass_count"`
	FailCount     int       `json:"fail_count"`
	LastTestedAt  time.Time `json:"last_tested_at"`
	MaturityStage string    `json:"maturity_stage"`
	Retired       bool      `json:"retired"`
}

// Memory holds the full confidence state for a project.
type Memory struct {
	Areas     map[string]*Area `json:"areas"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Version   int              `json:"version"`
}

// NewMemory creates a fresh Memory with no areas.
func NewMemory() *Memory {
	now := time.Now()
	return &Memory{
		Areas:     make(map[string]*Area),
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

// memoryPath returns the path to memory.json for the given project.
func memoryPath(projectPath string) string {
	return filepath.Join(projectPath, ".teststop", "memory.json")
}

// Load reads memory from projectPath/.teststop/memory.json.
// Returns NewMemory() if the file does not exist.
func Load(projectPath string) (*Memory, error) {
	data, err := os.ReadFile(memoryPath(projectPath))
	if err != nil {
		if os.IsNotExist(err) {
			return NewMemory(), nil
		}
		return nil, fmt.Errorf("reading memory file: %w", err)
	}

	var m Memory
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing memory file: %w", err)
	}
	if m.Areas == nil {
		m.Areas = make(map[string]*Area)
	}
	return &m, nil
}

// Save writes memory to projectPath/.teststop/memory.json (pretty-printed JSON).
// Creates the .teststop directory if needed.
func (m *Memory) Save(projectPath string) error {
	dir := filepath.Join(projectPath, ".teststop")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating .teststop directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling memory: %w", err)
	}

	if err := os.WriteFile(memoryPath(projectPath), data, 0o644); err != nil {
		return fmt.Errorf("writing memory file: %w", err)
	}
	return nil
}
