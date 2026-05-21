package memory

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := New()
	m.SystemAreas["auth"] = AreaConfidence{Confidence: 0.75, TestCount: 4, LastTested: time.Now().UTC().Truncate(time.Second)}
	m.OverallConfidence = 0.75
	m.MaturityStage = StageGrowing
	m.TotalRuns = 3

	if err := Save(dir, m); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.OverallConfidence != 0.75 {
		t.Fatalf("overall confidence = %v, want 0.75", loaded.OverallConfidence)
	}
	if loaded.SystemAreas["auth"].TestCount != 4 {
		t.Fatalf("auth test count = %d, want 4", loaded.SystemAreas["auth"].TestCount)
	}
}

func TestLoad_MissingIsEmpty(t *testing.T) {
	dir := t.TempDir()
	m, err := Load(dir)
	if err != nil {
		t.Fatalf("load missing: %v", err)
	}
	if len(m.SystemAreas) != 0 {
		t.Fatalf("expected empty memory, got %d areas", len(m.SystemAreas))
	}
}

func TestReset_RemovesMemoryFile(t *testing.T) {
	dir := t.TempDir()
	m := New()
	m.SystemAreas["auth"] = AreaConfidence{Confidence: 0.5, TestCount: 1}
	if err := Save(dir, m); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := Reset(dir); err != nil {
		t.Fatalf("reset: %v", err)
	}
	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("load after reset: %v", err)
	}
	if len(loaded.SystemAreas) != 0 {
		t.Fatalf("memory should be empty after reset, got %d areas", len(loaded.SystemAreas))
	}
}

func TestApply_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	m, _ := Load(dir)
	for i := 0; i < 15; i++ {
		m.Apply(RunOutcome{
			When:           time.Now(),
			AreaByScenario: map[string]string{"s1": "auth"},
			Results:        []scenario.Result{{ScenarioID: "s1", Passed: true}},
		})
	}
	retired := m.Retire(time.Now())
	if len(retired) != 1 {
		t.Fatalf("expected retirement after 15 consecutive passes, got %v (conf=%v)", retired, m.SystemAreas["auth"].Confidence)
	}
	if err := Save(dir, m); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := Load(filepath.Clean(dir)); err != nil {
		t.Fatalf("reload: %v", err)
	}
}
