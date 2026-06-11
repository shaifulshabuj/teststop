package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shaifulshabuj/teststop/internal/sandbox"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// ClaudeCLI implements AIAdapter by shelling out to the `claude` CLI.
type ClaudeCLI struct {
	runner *sandbox.Runner
}

// NewClaudeCLI creates a ClaudeCLI with auto sandbox detection.
func NewClaudeCLI() *ClaudeCLI {
	return &ClaudeCLI{runner: sandbox.New(sandbox.ModeFromEnv())}
}

func (c *ClaudeCLI) Name() string { return "claude" }

// claudeEnvelope is the JSON wrapper emitted by `claude --output-format json`.
// We parse this to detect errors (rate limit, auth, refusal) before extracting
// the inner result for downstream parsing.
//
// claude CLI 2.1.x+ streams a JSON array of event objects; earlier versions
// emit a single JSON object. Both carry the same is_error/result fields on the
// terminal "result" event, so we handle both with the same struct.
type claudeEnvelope struct {
	Type           string          `json:"type"` // event type: "system", "assistant", "result", etc.
	IsError        bool            `json:"is_error"`
	Result         string          `json:"result"`
	RateLimitEvent json.RawMessage `json:"rate_limit_event"`
}

// Prompt sends an arbitrary instruction to the claude CLI and returns raw stdout.
// It uses --output-format json so structured errors (rate limit, auth, refusal)
// are detectable; the inner .result is returned so callers see plain text output.
func (c *ClaudeCLI) Prompt(input string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{"-p", input, "--output-format", "json"}

	if model := os.Getenv("TESTSTOP_MODEL"); model != "" {
		args = append(args, "--model", model)
	}

	result := c.runner.Run(ctx, sandbox.RunConfig{}, "claude", args...)
	if result.Err != nil {
		// Try to pull a structured reason from whatever stdout we got before the
		// non-zero exit, to give callers a richer error message.
		if env, parseErr := parseClaudeEnvelope(result.Stdout); parseErr == nil && env.IsError {
			return nil, fmt.Errorf("claude: structured error: %s\nstderr: %s", env.Result, result.Stderr)
		}
		return nil, fmt.Errorf("claude: %w\nstderr: %s", result.Err, result.Stderr)
	}

	env, err := parseClaudeEnvelope(result.Stdout)
	if err != nil {
		// Envelope parse failed — fall back to raw stdout so legacy behaviour is
		// preserved (e.g. if an older claude CLI doesn't support --output-format json).
		return result.Stdout, nil
	}
	if env.IsError {
		detail := env.Result
		// Include rate_limit_event if present and non-null.
		if len(env.RateLimitEvent) > 0 && string(env.RateLimitEvent) != "null" {
			detail = fmt.Sprintf("%s (rate_limit_event: %s)", detail, env.RateLimitEvent)
		}
		return nil, fmt.Errorf("claude: %s", strings.TrimSpace(detail))
	}

	return []byte(env.Result), nil
}

// GenerateScenarios sends the mandate to the claude CLI and parses the returned JSON.
func (c *ClaudeCLI) GenerateScenarios(mandate string) ([]scenario.Scenario, error) {
	out, err := c.Prompt(mandate)
	if err != nil {
		return nil, err
	}
	return ParseScenariosFromJSON(out)
}

// parseClaudeEnvelope parses the JSON envelope from `claude --output-format json`.
//
// claude CLI 2.1.x+ emits a JSON array of streaming event objects; older
// versions emit a single JSON object. Both formats are handled:
//
//   - Single object: `{"is_error":false,"result":"..."}` — parsed directly.
//   - Array of events: `[{"type":"system",...},{"type":"result","is_error":false,"result":"..."}]`
//     — the last element whose "type" is "result" is returned.
func parseClaudeEnvelope(data []byte) (claudeEnvelope, error) {
	s := strings.TrimSpace(string(data))

	// Fast path: single JSON object (legacy format or non-streaming mode).
	if len(s) > 0 && s[0] == '{' {
		var env claudeEnvelope
		if err := json.Unmarshal([]byte(s), &env); err != nil {
			return claudeEnvelope{}, err
		}
		return env, nil
	}

	// Streaming array format: claude CLI 2.1.x+ wraps all events in a JSON
	// array. Find the last element whose "type" is "result" — that event carries
	// is_error and the inner result string.
	var events []claudeEnvelope
	if err := json.Unmarshal([]byte(s), &events); err != nil {
		return claudeEnvelope{}, fmt.Errorf("claude output is neither a JSON object nor a JSON array: %s", truncate([]byte(s), 200))
	}
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Type == "result" {
			return events[i], nil
		}
	}
	return claudeEnvelope{}, fmt.Errorf("claude streaming output: no 'result' event found in %d events", len(events))
}
