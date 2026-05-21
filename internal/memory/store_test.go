package memory_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shaifulshabuj/teststop/internal/memory"
)

func TestNewMemory(t *testing.T) {
	m := memory.NewMemory()
	if m.Areas == nil {
		t.Error("Areas map should be initialized")
	}
	if m.Version != 1 {
		t.Errorf("expected Version=1, got %d", m.Version)
	}
}

func TestLoad_nonExistentFile(t *testing.T) {
	tmp := t.TempDir()
	m, err := memory.Load(tmp)
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}
	if m == nil {
		t.Error("should return empty Memory, not nil")
	}
}

func TestSaveLoad_roundtrip(t *testing.T) {
	tmp := t.TempDir()
	m := memory.NewMemory()
	m.UpdateArea("auth", true)
	m.UpdateArea("auth", true)

	if err := m.Save(tmp); err != nil {
		t.Fatal(err)
	}

	m2, err := memory.Load(tmp)
	if err != nil {
		t.Fatal(err)
	}

	area := m2.Areas["auth"]
	if area == nil {
		t.Fatal("expected auth area after roundtrip")
	}
	if area.PassCount != 2 {
		t.Errorf("expected PassCount=2, got %d", area.PassCount)
	}
}

func TestUpdateArea_15PassesReachesRetirement(t *testing.T) {
	m := memory.NewMemory()
	for i := 0; i < 15; i++ {
		m.UpdateArea("api", true)
	}
	area := m.Areas["api"]
	if area.Confidence < memory.RetirementThreshold {
		t.Errorf("15 passes should reach %.2f threshold, got %.4f", memory.RetirementThreshold, area.Confidence)
	}
}

func TestUpdateArea_failDecreases(t *testing.T) {
	m := memory.NewMemory()
	m.UpdateArea("api", true)
	m.UpdateArea("api", true)
	beforeFail := m.Areas["api"].Confidence

	m.UpdateArea("api", false)
	afterFail := m.Areas["api"].Confidence

	if afterFail >= beforeFail {
		t.Errorf("confidence should decrease after fail: was %.4f, now %.4f", beforeFail, afterFail)
	}
	if afterFail < 0.0 {
		t.Errorf("confidence should not go below 0: got %.4f", afterFail)
	}
}

func TestUpdateArea_confidenceFloor(t *testing.T) {
	m := memory.NewMemory()
	// Multiple failures on a new area should not go below 0.
	m.UpdateArea("fragile", false)
	m.UpdateArea("fragile", false)
	m.UpdateArea("fragile", false)
	if m.Areas["fragile"].Confidence < 0.0 {
		t.Errorf("confidence must not go below 0: got %.4f", m.Areas["fragile"].Confidence)
	}
}

func TestGetVolatileAreas(t *testing.T) {
	m := memory.NewMemory()
	m.UpdateArea("new-area", true) // confidence = 0.19, < 0.75 → volatile

	volatile := m.GetVolatileAreas()
	if len(volatile) != 1 {
		t.Errorf("expected 1 volatile area, got %d", len(volatile))
	}
}

func TestGetStableAreas(t *testing.T) {
	m := memory.NewMemory()
	for i := 0; i < 15; i++ {
		m.UpdateArea("proven-area", true)
	}

	stable := m.GetStableAreas()
	if len(stable) == 0 {
		t.Error("expected at least 1 stable area after 15 passes")
	}
}

func TestGetVolatileAreas_excludesRetired(t *testing.T) {
	tmp := t.TempDir()
	m := memory.NewMemory()
	// Make area eligible for retirement.
	for i := 0; i < 15; i++ {
		m.UpdateArea("old-area", true)
	}
	if _, err := m.RetireEligibleAreas(tmp); err != nil {
		t.Fatal(err)
	}
	// Add a genuinely volatile area.
	m.UpdateArea("volatile-area", true)

	volatile := m.GetVolatileAreas()
	for _, a := range volatile {
		if a.Name == "old-area" {
			t.Error("retired area should not appear in volatile list")
		}
	}
}

func TestRetireEligibleAreas(t *testing.T) {
	tmp := t.TempDir()
	m := memory.NewMemory()

	// Make area eligible: confidence >= 0.95 AND testCount >= 15.
	for i := 0; i < 15; i++ {
		m.UpdateArea("proven", true)
	}

	retired, err := m.RetireEligibleAreas(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(retired) == 0 {
		t.Error("expected 'proven' to be retired")
	}

	// Area should still be in Areas map but marked retired.
	if !m.Areas["proven"].Retired {
		t.Error("area should be marked as Retired in memory")
	}

	// retired.json should exist.
	retiredPath := filepath.Join(tmp, ".teststop", "retired.json")
	if _, err := os.Stat(retiredPath); os.IsNotExist(err) {
		t.Error("retired.json should be created")
	}
}

func TestRetireEligibleAreas_requiresMinTestCount(t *testing.T) {
	tmp := t.TempDir()
	m := memory.NewMemory()

	// Force confidence >= 0.95 manually via enough passes but verify testCount gate.
	// After 5 additive passes (not exponential — just check the gate):
	// With exponential formula, need 15 passes for >= 0.95.
	// Use 14 passes — confidence will be just under 0.95.
	for i := 0; i < 14; i++ {
		m.UpdateArea("almost", true)
	}

	retired, err := m.RetireEligibleAreas(tmp)
	if err != nil {
		t.Fatal(err)
	}
	// Should NOT be retired — either confidence < 0.95 or testCount < 15.
	if len(retired) != 0 {
		t.Errorf("area with only 14 passes should not be retired, got: %v", retired)
	}
}

func TestRetireEligibleAreas_idempotent(t *testing.T) {
	tmp := t.TempDir()
	m := memory.NewMemory()
	for i := 0; i < 15; i++ {
		m.UpdateArea("stable", true)
	}

	first, err := m.RetireEligibleAreas(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 {
		t.Fatal("expected retirement on first call")
	}

	// Second call should retire nothing (area already marked retired).
	second, err := m.RetireEligibleAreas(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Errorf("second retirement call should retire nothing, got: %v", second)
	}
}
