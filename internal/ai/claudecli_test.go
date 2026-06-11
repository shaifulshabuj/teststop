package ai

import (
	"strings"
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

// TestParseClaudeEnvelope_streamingArray verifies the claude CLI 2.1.x+ format:
// a JSON array of event objects where the last "result" event carries the payload.
func TestParseClaudeEnvelope_streamingArray(t *testing.T) {
	// Mirrors the actual output format observed from claude 2.1.172.
	raw := []byte(`[
		{"type":"system","subtype":"init","session_id":"s1","tools":[]},
		{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"},"session_id":"s1"},
		{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"[...]"}]},"session_id":"s1"},
		{"type":"result","subtype":"success","is_error":false,"result":"[{\"scenario_id\":\"sc-stream-001\"}]","session_id":"s1"}
	]`)
	env, err := parseClaudeEnvelope(raw)
	if err != nil {
		t.Fatalf("streaming array format should parse: %v", err)
	}
	if env.IsError {
		t.Error("is_error should be false")
	}
	if !strings.Contains(env.Result, "sc-stream-001") {
		t.Errorf("result should contain scenario id, got: %q", env.Result)
	}
}

// TestParseClaudeEnvelope_streamingArrayError verifies that an error event
// in streaming format is correctly detected.
func TestParseClaudeEnvelope_streamingArrayError(t *testing.T) {
	raw := []byte(`[
		{"type":"system","session_id":"s1"},
		{"type":"result","subtype":"error_during_run","is_error":true,"result":"rate limit exceeded","session_id":"s1"}
	]`)
	env, err := parseClaudeEnvelope(raw)
	if err != nil {
		t.Fatalf("error envelope in array format should parse: %v", err)
	}
	if !env.IsError {
		t.Error("is_error should be true for error event")
	}
	if !strings.Contains(env.Result, "rate limit") {
		t.Errorf("result should contain error message, got: %q", env.Result)
	}
}

// TestParseClaudeEnvelope_arrayNoResultEvent verifies an error is returned when
// there is no "result" type event in the array.
func TestParseClaudeEnvelope_arrayNoResultEvent(t *testing.T) {
	raw := []byte(`[{"type":"system","session_id":"s1"},{"type":"assistant","session_id":"s1"}]`)
	_, err := parseClaudeEnvelope(raw)
	if err == nil {
		t.Error("expected error when no 'result' event in array")
	}
}
