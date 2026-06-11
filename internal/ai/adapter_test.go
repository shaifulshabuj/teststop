package ai_test

import (
	"strings"
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

func TestParseScenarios_legacyHasNilExec(t *testing.T) {
	// v0.1 JSON has no exec field; it must still parse with Exec == nil.
	raw := []byte(`[{"scenario_id":"s1","title":"t","user_perspective":"u","preconditions":[],"steps":["a"],"chaos_factors":[],"expected_behavior":"e","failure_modes":[],"priority":"low","confidence_area":"x","is_edge_case":false}]`)
	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("legacy JSON should parse: %v", err)
	}
	if scenarios[0].Exec != nil {
		t.Errorf("expected nil Exec for legacy scenario, got %+v", scenarios[0].Exec)
	}
}

func TestParseScenarios_execRoundTrips(t *testing.T) {
	raw := []byte(`[{"scenario_id":"s1","title":"t","user_perspective":"u","preconditions":[],"steps":["a"],"chaos_factors":[],"expected_behavior":"e","failure_modes":[],"priority":"critical","confidence_area":"auth","is_edge_case":true,"exec":{"mode":"http","method":"POST","path":"/api/login","headers":{"Content-Type":"application/json"},"body":"{}","expected_status":400}}]`)
	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("exec JSON should parse: %v", err)
	}
	e := scenarios[0].Exec
	if e == nil {
		t.Fatal("expected non-nil Exec")
	}
	if e.Mode != "http" || e.Method != "POST" || e.Path != "/api/login" || e.ExpectedStatus != 400 {
		t.Errorf("exec fields not parsed correctly: %+v", e)
	}
	if e.Headers["Content-Type"] != "application/json" {
		t.Errorf("exec headers not parsed: %+v", e.Headers)
	}
}

func TestParseScenarios_execConcurrency(t *testing.T) {
	raw := []byte(`[{"scenario_id":"s1","title":"t","user_perspective":"u","preconditions":[],"steps":["a"],"chaos_factors":[],"expected_behavior":"e","failure_modes":[],"priority":"critical","confidence_area":"actions","is_edge_case":true,"exec":{"mode":"http","method":"POST","path":"/approve","expected_status":200,"concurrency":10}}]`)
	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("exec concurrency JSON should parse: %v", err)
	}
	if e := scenarios[0].Exec; e == nil || e.Concurrency != 10 {
		t.Errorf("expected concurrency 10, got %+v", e)
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

// TestParseScenarios_proseBeforeArray verifies that a local model emitting reasoning
// prose before the JSON array is still parsed correctly (qwen3:4b failure mode).
func TestParseScenarios_proseBeforeArray(t *testing.T) {
	raw := []byte(`We are generating exactly 30 scenarios for the waymark API system.
The system type is "api", so we focus on the API level.

[{"scenario_id":"prose-001","title":"Double submit","user_perspective":"impatient user","preconditions":[],"steps":["submit twice"],"chaos_factors":[],"expected_behavior":"idempotent","failure_modes":["duplicate entry"],"priority":"high","confidence_area":"api/submit","is_edge_case":false}]`)

	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("prose preamble should be tolerated: %v", err)
	}
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}
	if scenarios[0].ScenarioID != "prose-001" {
		t.Errorf("unexpected scenario_id: %s", scenarios[0].ScenarioID)
	}
}

// TestParseScenarios_proseAfterArray verifies that trailing prose after ']' is ignored.
func TestParseScenarios_proseAfterArray(t *testing.T) {
	raw := []byte(`[{"scenario_id":"trail-001","title":"T","user_perspective":"u","preconditions":[],"steps":[],"chaos_factors":[],"expected_behavior":"e","failure_modes":[],"priority":"low","confidence_area":"x","is_edge_case":false}]

This completes the scenario generation.`)

	scenarios, err := ai.ParseScenariosFromJSON(raw)
	if err != nil {
		t.Fatalf("trailing prose should be tolerated: %v", err)
	}
	if len(scenarios) != 1 {
		t.Errorf("expected 1 scenario, got %d", len(scenarios))
	}
}

// TestParseScenarios_hollowArray exercises the defense-in-depth validation that
// rejects a parsed batch where every object is hollow (empty scenario_id and
// title). This is the failure mode when the raw AI CLI event stream is
// accidentally fed to ParseScenariosFromJSON instead of the inner result string.
func TestParseScenarios_hollowArray_rejected(t *testing.T) {
	// Mimics the JSONL streaming events from claude CLI 2.1.x+ being parsed as
	// scenarios: each event object has no fields that match Scenario json tags,
	// so all structs come out zero-valued.
	hollow := `[
		{"type":"system","session_id":"abc","subtype":"init"},
		{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"}},
		{"type":"assistant","message":{"role":"assistant"},"session_id":"abc"},
		{"type":"result","subtype":"success","is_error":false,"result":"...","session_id":"abc"}
	]`
	_, err := ai.ParseScenariosFromJSON([]byte(hollow))
	if err == nil {
		t.Fatal("expected error for hollow scenario array, got nil")
	}
	if !strings.Contains(err.Error(), "hollow") {
		t.Errorf("error should mention 'hollow', got: %v", err)
	}
}
