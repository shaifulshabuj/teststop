package ai_test

import (
	"testing"

	"github.com/shaifulshabuj/teststop/internal/ai"
)

func TestParseScenarios_validJSON(t *testing.T) {
	raw := []byte(`[
  {
    "scenario_id": "test-001",
    "title": "Empty form submission",
    "user_perspective": "A new user who doesn't know required fields",
    "preconditions": ["User is on the registration form"],
    "steps": ["Click Submit without filling anything"],
    "chaos_factors": ["impatient user"],
    "expected_behavior": "Show validation errors for all required fields",
    "failure_modes": ["Form submits with empty data"],
    "priority": "high",
    "confidence_area": "auth/registration",
    "is_edge_case": false
  }
]`)

	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}
	if scenarios[0].ScenarioID != "test-001" {
		t.Errorf("expected id 'test-001', got %q", scenarios[0].ScenarioID)
	}
}

func TestParseScenarios_stripsMarkdownFences(t *testing.T) {
	raw := []byte("```json\n[{\"scenario_id\":\"s1\",\"title\":\"t\",\"user_perspective\":\"u\",\"preconditions\":[],\"steps\":[],\"chaos_factors\":[],\"expected_behavior\":\"e\",\"failure_modes\":[],\"priority\":\"low\",\"confidence_area\":\"x\",\"is_edge_case\":false}]\n```")

	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("should strip fences: %v", err)
	}
	if len(scenarios) != 1 {
		t.Errorf("expected 1 scenario, got %d", len(scenarios))
	}
}

func TestParseScenarios_stripsPlainFences(t *testing.T) {
	raw := []byte("```\n[{\"scenario_id\":\"s2\",\"title\":\"t\",\"user_perspective\":\"u\",\"preconditions\":[],\"steps\":[],\"chaos_factors\":[],\"expected_behavior\":\"e\",\"failure_modes\":[],\"priority\":\"low\",\"confidence_area\":\"x\",\"is_edge_case\":false}]\n```")

	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("should strip plain fences: %v", err)
	}
	if len(scenarios) != 1 {
		t.Errorf("expected 1 scenario, got %d", len(scenarios))
	}
}

func TestParseScenarios_invalidJSON(t *testing.T) {
	_, err := ai.ParseScenariosFromJSON([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseScenarios_emptyArray(t *testing.T) {
	scenarios, err := ai.ParseScenariosFromJSON([]byte("[]"))
	if err != nil {
		t.Fatalf("empty array should parse: %v", err)
	}
	if len(scenarios) != 0 {
		t.Errorf("expected 0 scenarios, got %d", len(scenarios))
	}
}

func TestParseScenarios_whitespace(t *testing.T) {
	raw := []byte("  \n  []  \n  ")
	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("whitespace-padded array should parse: %v", err)
	}
	if len(scenarios) != 0 {
		t.Errorf("expected 0 scenarios, got %d", len(scenarios))
	}
}

func TestParseScenarios_multipleScenarios(t *testing.T) {
	raw := []byte(`[
  {"scenario_id":"a1","title":"A","user_perspective":"u","preconditions":[],"steps":[],"chaos_factors":[],"expected_behavior":"e","failure_modes":[],"priority":"critical","confidence_area":"auth","is_edge_case":true},
  {"scenario_id":"b2","title":"B","user_perspective":"u","preconditions":[],"steps":[],"chaos_factors":[],"expected_behavior":"e","failure_modes":[],"priority":"medium","confidence_area":"api","is_edge_case":false}
]`)

	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(scenarios) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(scenarios))
	}
	if scenarios[0].Priority != "critical" {
		t.Errorf("expected critical priority, got %q", scenarios[0].Priority)
	}
	if !scenarios[0].IsEdgeCase {
		t.Error("expected first scenario to be edge case")
	}
}
