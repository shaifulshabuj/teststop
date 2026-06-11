package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// cannedScenarioJSON is a minimal valid scenario array returned by canned servers.
const cannedScenarioJSON = `[{"scenario_id":"ol-001","title":"Empty submit","user_perspective":"impatient user","preconditions":["form is loaded"],"steps":["click Submit"],"chaos_factors":["fast network"],"expected_behavior":"validation error shown","failure_modes":["silent data loss"],"priority":"high","confidence_area":"auth/form","is_edge_case":false}]`

// newCannedServer returns an httptest.Server that responds with body for POST /api/generate.
func newCannedServer(body string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
}

// newOllamaCLIWithBase creates an OllamaCLI pointed at the given base URL.
func newOllamaCLIWithBase(baseURL, model string) *OllamaCLI {
	return &OllamaCLI{
		baseURL: baseURL,
		model:   model,
		client:  http.DefaultClient,
	}
}

func TestOllamaCLI_Name(t *testing.T) {
	o := NewOllamaCLI()
	if o.Name() != "ollama" {
		t.Errorf("expected name 'ollama', got %q", o.Name())
	}
}

// TestOllamaCLI_Prompt_cleanJSON verifies the happy path: model returns plain JSON,
// Prompt returns it unchanged.
func TestOllamaCLI_Prompt_cleanJSON(t *testing.T) {
	resp := ollamaGenerateResponse{Response: cannedScenarioJSON, Done: true}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	out, err := o.Prompt("test mandate")
	if err != nil {
		t.Fatalf("Prompt error: %v", err)
	}
	if string(out) != cannedScenarioJSON {
		t.Errorf("unexpected output: %s", out)
	}
}

// TestOllamaCLI_Prompt_stripsThinkBlocks verifies that <think>...</think> blocks
// are stripped before returning. This handles qwen3 models on older ollama builds
// that ignore think:false.
func TestOllamaCLI_Prompt_stripsThinkBlocks(t *testing.T) {
	raw := "<think>\nlet me reason about this carefully...\n</think>\n" + cannedScenarioJSON
	resp := ollamaGenerateResponse{Response: raw, Done: true}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	out, err := o.Prompt("test mandate")
	if err != nil {
		t.Fatalf("Prompt error: %v", err)
	}
	if strings.Contains(string(out), "<think>") {
		t.Error("think block should be stripped")
	}
	if !strings.Contains(string(out), "ol-001") {
		t.Errorf("scenario content missing from output: %s", out)
	}
}

// TestOllamaCLI_Prompt_stripsMultipleThinkBlocks verifies nested/multiple think blocks.
func TestOllamaCLI_Prompt_stripsMultipleThinkBlocks(t *testing.T) {
	raw := "<think>first</think>some preamble<think>second\nmultiline</think>" + cannedScenarioJSON
	resp := ollamaGenerateResponse{Response: raw, Done: true}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	out, err := o.Prompt("mandate")
	if err != nil {
		t.Fatalf("Prompt error: %v", err)
	}
	if strings.Contains(string(out), "<think>") || strings.Contains(string(out), "</think>") {
		t.Error("all think blocks should be stripped")
	}
}

// TestOllamaCLI_Prompt_httpError verifies non-200 responses return an error.
func TestOllamaCLI_Prompt_httpError(t *testing.T) {
	srv := newCannedServer(`{"error":"model not found"}`, http.StatusNotFound)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "nonexistent:latest")
	_, err := o.Prompt("mandate")
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention HTTP status, got: %v", err)
	}
}

// TestOllamaCLI_Prompt_modelError verifies that a non-empty error field in the
// response body is surfaced as an error.
func TestOllamaCLI_Prompt_modelError(t *testing.T) {
	resp := ollamaGenerateResponse{Error: "context length exceeded"}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	_, err := o.Prompt("mandate")
	if err == nil {
		t.Fatal("expected error for model error field")
	}
	if !strings.Contains(err.Error(), "context length") {
		t.Errorf("error should contain model error message, got: %v", err)
	}
}

// TestOllamaCLI_Prompt_unreachable verifies that an unreachable server returns an error.
func TestOllamaCLI_Prompt_unreachable(t *testing.T) {
	o := newOllamaCLIWithBase("http://127.0.0.1:19999", "qwen3.6:latest")
	// Use a small timeout so the test doesn't block long.
	o.client = &http.Client{Timeout: 1}
	_, err := o.Prompt("mandate")
	if err == nil {
		t.Fatal("expected error when server is unreachable")
	}
}

// TestOllamaCLI_GenerateScenarios_clean verifies end-to-end parse of a clean response.
func TestOllamaCLI_GenerateScenarios_clean(t *testing.T) {
	resp := ollamaGenerateResponse{Response: cannedScenarioJSON, Done: true}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	scenarios, err := o.GenerateScenarios("mandate text")
	if err != nil {
		t.Fatalf("GenerateScenarios error: %v", err)
	}
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}
	if scenarios[0].ScenarioID != "ol-001" {
		t.Errorf("unexpected scenario_id: %s", scenarios[0].ScenarioID)
	}
}

// TestOllamaCLI_GenerateScenarios_fencedJSON verifies that markdown-fenced JSON is
// parsed correctly — local models frequently wrap output in ``` fences.
func TestOllamaCLI_GenerateScenarios_fencedJSON(t *testing.T) {
	fenced := "```json\n" + cannedScenarioJSON + "\n```"
	resp := ollamaGenerateResponse{Response: fenced, Done: true}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	scenarios, err := o.GenerateScenarios("mandate")
	if err != nil {
		t.Fatalf("fenced JSON should parse: %v", err)
	}
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}
}

// TestOllamaCLI_GenerateScenarios_thinkThenFencedJSON verifies the combined case:
// think block before fenced JSON (real qwen3 sloppy output).
func TestOllamaCLI_GenerateScenarios_thinkThenFencedJSON(t *testing.T) {
	raw := "<think>\nI should generate scenarios for this system.\n</think>\n```json\n" + cannedScenarioJSON + "\n```"
	resp := ollamaGenerateResponse{Response: raw, Done: true}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	scenarios, err := o.GenerateScenarios("mandate")
	if err != nil {
		t.Fatalf("think+fenced combo should parse: %v", err)
	}
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}
}

// TestOllamaCLI_GenerateScenarios_hollowRejected verifies that a hollow batch
// (local model returned empty objects) is rejected rather than silently accepted.
func TestOllamaCLI_GenerateScenarios_hollowRejected(t *testing.T) {
	hollow := `[{"some":"field"},{"other":"value"}]`
	resp := ollamaGenerateResponse{Response: hollow, Done: true}
	b, _ := json.Marshal(resp)
	srv := newCannedServer(string(b), http.StatusOK)
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	_, err := o.GenerateScenarios("mandate")
	if err == nil {
		t.Fatal("hollow batch should be rejected")
	}
	if !strings.Contains(err.Error(), "hollow") {
		t.Errorf("error should mention 'hollow', got: %v", err)
	}
}

// TestOllamaCLI_jsonOnlySuffix verifies the JSON-only suffix is appended to the
// mandate when calling GenerateScenarios (not when calling Prompt directly).
func TestOllamaCLI_jsonOnlySuffix(t *testing.T) {
	var receivedPrompt string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaGenerateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt
		resp := ollamaGenerateResponse{Response: cannedScenarioJSON, Done: true}
		b, _ := json.Marshal(resp)
		fmt.Fprint(w, string(b))
	}))
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "qwen3.6:latest")
	_, err := o.GenerateScenarios("base mandate")
	if err != nil {
		t.Fatalf("GenerateScenarios error: %v", err)
	}
	if !strings.Contains(receivedPrompt, jsonOnlySuffix) {
		t.Errorf("JSON-only suffix not appended: got prompt = %q", receivedPrompt)
	}
}

// TestOllamaCLI_requestShape verifies the HTTP request sent to ollama has the correct
// fields: model, stream=false, think=false, and num_ctx set.
func TestOllamaCLI_requestShape(t *testing.T) {
	var req ollamaGenerateRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&req)
		resp := ollamaGenerateResponse{Response: cannedScenarioJSON, Done: true}
		b, _ := json.Marshal(resp)
		fmt.Fprint(w, string(b))
	}))
	defer srv.Close()

	o := newOllamaCLIWithBase(srv.URL, "test-model:latest")
	_, err := o.Prompt("mandate")
	if err != nil {
		t.Fatalf("Prompt error: %v", err)
	}

	if req.Model != "test-model:latest" {
		t.Errorf("expected model 'test-model:latest', got %q", req.Model)
	}
	if req.Stream {
		t.Error("stream should be false")
	}
	if req.Think {
		t.Error("think should be false")
	}
	numCtx, ok := req.Options["num_ctx"]
	if !ok {
		t.Error("num_ctx option should be set")
	}
	// JSON numbers unmarshal as float64 into map[string]any
	if numCtxF, ok := numCtx.(float64); !ok || int(numCtxF) != ollamaNumCtx {
		t.Errorf("expected num_ctx %d, got %v", ollamaNumCtx, numCtx)
	}
}
