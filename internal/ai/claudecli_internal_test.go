package ai

import (
	"strings"
	"testing"
)

func TestExtractClaudeResult_happyPath(t *testing.T) {
	env := []byte(`{"type":"result","result":"[{\"scenario_id\":\"s1\",\"title\":\"t\",\"user_perspective\":\"u\",\"preconditions\":[],\"steps\":[],\"chaos_factors\":[],\"expected_behavior\":\"e\",\"failure_modes\":[],\"priority\":\"low\",\"confidence_area\":\"x\",\"is_edge_case\":false}]","session_id":"sess-123","total_cost_usd":"0.01"}`)

	out, err := extractClaudeResult(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `[{"scenario_id":"s1","title":"t","user_perspective":"u","preconditions":[],"steps":[],"chaos_factors":[],"expected_behavior":"e","failure_modes":[],"priority":"low","confidence_area":"x","is_edge_case":false}]`
	if string(out) != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestExtractClaudeResult_fallsBackToRawOnInvalidJSON(t *testing.T) {
	raw := []byte(`not a json envelope`)

	out, err := extractClaudeResult(raw)
	if err != nil {
		t.Fatalf("should fallback to raw on invalid envelope: %v", err)
	}
	if string(out) != string(raw) {
		t.Errorf("got %q, want raw fallback", out)
	}
}

func TestExtractClaudeResult_isError(t *testing.T) {
	env := []byte(`{"type":"result","result":"something","is_error":true}`)

	_, err := extractClaudeResult(env)
	if err == nil {
		t.Fatal("expected error for is_error=true")
	}
	if !strings.Contains(err.Error(), "is_error=true") {
		t.Errorf("error message should mention is_error: %v", err)
	}
}

func TestExtractClaudeResult_rateLimit(t *testing.T) {
	env := []byte(`{"type":"result","result":"slow down","rate_limit_event":{"status":"rate_limit_exceeded"}}`)

	_, err := extractClaudeResult(env)
	if err == nil {
		t.Fatal("expected error for rate_limit_event")
	}
	if !strings.Contains(err.Error(), "rate-limit event") {
		t.Errorf("error message should mention rate-limit: %v", err)
	}
}

func TestExtractClaudeResult_rateLimitOk(t *testing.T) {
	env := []byte(`{"type":"result","result":"[{\"scenario_id\":\"s1\"}]","rate_limit_event":{"status":"ok"}}`)

	out, err := extractClaudeResult(env)
	if err != nil {
		t.Fatalf("unexpected error for rate_limit ok: %v", err)
	}
	if !strings.Contains(string(out), `"scenario_id"`) {
		t.Errorf("expected inner result, got %q", out)
	}
}

func TestExtractClaudeResult_emptyRateLimitAllowed(t *testing.T) {
	env := []byte(`{"type":"result","result":"all good","rate_limit_event":{}}`)

	out, err := extractClaudeResult(env)
	if err != nil {
		t.Fatalf("unexpected error for empty rate-limit event: %v", err)
	}
	if string(out) != "all good" {
		t.Errorf("got %q, want 'all good'", out)
	}
}

func TestExtractClaudeResult_verdictJSON(t *testing.T) {
	env := []byte(`{"type":"result","result":"{\"passed\": true, \"actual_behavior\": \"ok\", \"failure_reason\": \"\"}","session_id":"sess-456"}`)

	out, err := extractClaudeResult(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), `"passed": true`) {
		t.Errorf("expected inner verdict JSON, got %q", out)
	}
}
