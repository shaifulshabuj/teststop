package ai

import (
	"testing"
)

// TestParseClaudeEnvelope_success verifies that a well-formed success envelope
// is parsed and that Result is extracted correctly.
func TestParseClaudeEnvelope_success(t *testing.T) {
	raw := []byte(`{"type":"result","subtype":"success","is_error":false,"result":"[{\"scenario_id\":\"s1\"}]","rate_limit_event":null}`)
	env, err := parseClaudeEnvelope(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if env.IsError {
		t.Error("is_error should be false")
	}
	if env.Result == "" {
		t.Error("result should not be empty")
	}
}

// TestParseClaudeEnvelope_isError verifies that an error envelope sets IsError.
func TestParseClaudeEnvelope_isError(t *testing.T) {
	raw := []byte(`{"type":"result","subtype":"error_during_run","is_error":true,"result":"rate limit exceeded","rate_limit_event":{"status":429}}`)
	env, err := parseClaudeEnvelope(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !env.IsError {
		t.Error("is_error should be true")
	}
	if env.Result == "" {
		t.Error("error result should be non-empty")
	}
	if string(env.RateLimitEvent) == "null" || len(env.RateLimitEvent) == 0 {
		t.Error("rate_limit_event should be populated")
	}
}

// TestParseClaudeEnvelope_invalid verifies that non-envelope JSON returns an error.
func TestParseClaudeEnvelope_invalid(t *testing.T) {
	_, err := parseClaudeEnvelope([]byte("not json"))
	if err == nil {
		t.Error("expected parse error for non-JSON input")
	}
}

// TestParseClaudeEnvelope_rateLimitNull verifies null rate_limit_event is handled.
func TestParseClaudeEnvelope_rateLimitNull(t *testing.T) {
	raw := []byte(`{"type":"result","is_error":false,"result":"ok","rate_limit_event":null}`)
	env, err := parseClaudeEnvelope(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// null rate_limit_event should be represented as "null" raw JSON
	if string(env.RateLimitEvent) != "null" && len(env.RateLimitEvent) != 0 {
		t.Errorf("unexpected rate_limit_event: %s", env.RateLimitEvent)
	}
}
