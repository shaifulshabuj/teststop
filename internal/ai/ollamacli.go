package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

const (
	ollamaDefaultBaseURL = "http://localhost:11434"
	ollamaDefaultModel   = "qwen3.6:latest"
	ollamaNumCtx         = 32768

	// jsonOnlySuffix is appended to the mandate for local models, which are less
	// instruction-following than cloud models and need explicit output constraints.
	jsonOnlySuffix = "\n\nIMPORTANT: Your entire response MUST be a single valid JSON array.\nDo NOT include any explanatory text, preamble, or markdown prose outside the array.\nBegin your response with [ and end with ]."
)

// thinkBlockRE strips <think>...</think> reasoning chains emitted by qwen3 family
// models even when think:false is set in the request (older ollama versions ignore
// the flag).
var thinkBlockRE = regexp.MustCompile(`(?s)<think>.*?</think>`)

// OllamaCLI implements AIAdapter by calling the ollama HTTP API at localhost:11434.
// Unlike ClaudeCLI/CopilotCLI it does not use sandbox.Runner — it calls net/http
// directly, so it works regardless of sandbox mode.
type OllamaCLI struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaCLI creates an OllamaCLI. The base URL and model are taken from
// ollamaDefaultBaseURL / TESTSTOP_MODEL (defaulting to qwen3.6:latest).
func NewOllamaCLI() *OllamaCLI {
	model := os.Getenv("TESTSTOP_MODEL")
	if model == "" {
		model = ollamaDefaultModel
	}
	return &OllamaCLI{
		baseURL: ollamaDefaultBaseURL,
		model:   model,
		client:  &http.Client{Timeout: 10 * time.Minute},
	}
}

func (o *OllamaCLI) Name() string { return "ollama" }

// ollamaGenerateRequest is the JSON body for POST /api/generate.
type ollamaGenerateRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Think   bool           `json:"think"`
	Options map[string]any `json:"options"`
}

// ollamaGenerateResponse is the JSON body returned by POST /api/generate when
// stream is false.
type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// Prompt sends input to the ollama generate API and returns the raw response text.
// It sets stream=false and think=false, and strips any residual <think> blocks.
func (o *OllamaCLI) Prompt(input string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	reqBody := ollamaGenerateRequest{
		Model:   o.model,
		Prompt:  input,
		Stream:  false,
		Think:   false,
		Options: map[string]any{"num_ctx": ollamaNumCtx},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		o.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ollama: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, truncate(rawBody, 300))
	}

	var genResp ollamaGenerateResponse
	if err := json.Unmarshal(rawBody, &genResp); err != nil {
		return nil, fmt.Errorf("ollama: failed to parse response JSON: %w\nraw: %s", err, truncate(rawBody, 300))
	}

	if genResp.Error != "" {
		return nil, fmt.Errorf("ollama: model error: %s", genResp.Error)
	}

	// Strip <think>...</think> blocks (qwen3 thinking chain — should be absent when
	// think:false but older ollama versions ignore the flag).
	text := thinkBlockRE.ReplaceAllString(genResp.Response, "")
	text = strings.TrimSpace(text)

	return []byte(text), nil
}

// GenerateScenarios appends the JSON-only output constraint to the mandate
// (local models need explicit instruction; cloud adapters do not) and parses the
// returned JSON using the shared ParseScenariosFromJSON which handles fences and
// hollow-batch detection.
func (o *OllamaCLI) GenerateScenarios(mandate string) ([]scenario.Scenario, error) {
	out, err := o.Prompt(mandate + jsonOnlySuffix)
	if err != nil {
		return nil, err
	}
	return ParseScenariosFromJSON(out)
}

// IsOllamaAvailable returns true if the ollama HTTP API is reachable.
// Uses a short timeout so auto-detection is not slow.
func IsOllamaAvailable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(ollamaDefaultBaseURL + "/")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}
