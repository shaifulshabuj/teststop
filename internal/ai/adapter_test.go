package ai

import (
	"strings"
	"testing"
)

func TestParseScenarios_PlainArray(t *testing.T) {
	raw := `[
	  {"scenario_id":"s1","title":"T","user_perspective":"u","preconditions":[],"steps":["a"],"chaos_factors":["slow"],"expected_behavior":"ok","failure_modes":["x"],"priority":"high","confidence_area":"auth","is_edge_case":false}
	]`
	s, err := parseScenarios(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(s) != 1 || s[0].ScenarioID != "s1" {
		t.Fatalf("unexpected scenarios: %+v", s)
	}
}

func TestParseScenarios_StripsCodeFences(t *testing.T) {
	raw := "```json\n[" +
		`{"scenario_id":"s1","title":"T","user_perspective":"u","preconditions":[],"steps":["a"],"chaos_factors":["c"],"expected_behavior":"ok","failure_modes":["f"],"priority":"low","confidence_area":"x","is_edge_case":false}` +
		"]\n```"
	s, err := parseScenarios(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(s) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(s))
	}
}

func TestParseScenarios_RejectsGarbage(t *testing.T) {
	_, err := parseScenarios("Here are the scenarios you asked for!")
	if err == nil {
		t.Fatal("expected error on non-JSON input")
	}
	if !strings.Contains(err.Error(), "not valid scenario JSON") {
		t.Fatalf("error should mention invalid JSON, got %v", err)
	}
}

func TestParseScenarios_AcceptsSingleObject(t *testing.T) {
	raw := `{"scenario_id":"s1","title":"T","user_perspective":"u","preconditions":[],"steps":["a"],"chaos_factors":["c"],"expected_behavior":"ok","failure_modes":["f"],"priority":"low","confidence_area":"x","is_edge_case":false}`
	s, err := parseScenarios(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(s) != 1 || s[0].ScenarioID != "s1" {
		t.Fatalf("unexpected scenarios: %+v", s)
	}
}
