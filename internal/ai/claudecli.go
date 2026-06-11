package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

// claudeEnvelope is the JSON envelope returned by `claude --output-format json`.
type claudeEnvelope struct {
	Type             string          `json:"type"`
	Result           string          `json:"result"`
	IsError          bool            `json:"is_error"`
	RateLimitEvent   *rateLimitEvent `json:"rate_limit_event,omitempty"`
	SessionID        string          `json:"session_id,omitempty"`
	NumTurns         int             `json:"num_turns,omitempty"`
	TotalCostUSD     string          `json:"total_cost_usd,omitempty"`
	DurationMs       int             `json:"duration_ms,omitempty"`
}

type rateLimitEvent struct {
	Status string `json:"status"`
}

// extractClaudeResult parses the JSON envelope from `--output-format json` and
// returns the inner `.result` string.  If the envelope signals an error or an
// active rate-limit, it returns an error so that callers (generation and
// execution) treat it as an infrastructure failure and surface it as skipped
// rather than a parse failure.
func extractClaudeResult(data []byte) ([]byte, error) {
	var env claudeEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		// Not valid envelope JSON — fall back to returning raw data so
		// existing tolerant parsing still works in unexpected edge cases.
		return data, nil
	}

	if env.IsError {
		return nil, fmt.Errorf("claude: is_error=true (envelope error)\nresult: %s", env.Result)
	}

	if env.RateLimitEvent != nil && env.RateLimitEvent.Status != "" && env.RateLimitEvent.Status != "ok" {
		return nil, fmt.Errorf("claude: rate-limit event (%s)\nresult: %s", env.RateLimitEvent.Status, env.Result)
	}

	return []byte(env.Result), nil
}

// Prompt sends an arbitrary instruction to the claude CLI and returns the
// content extracted from the `--output-format json` envelope.
func (c *ClaudeCLI) Prompt(input string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{"-p", input, "--output-format", "json"}

	// Optional model override.
	if model := os.Getenv("TESTSTOP_MODEL"); model != "" {
		args = append(args, "--model", model)
	}

	result := c.runner.Run(ctx, sandbox.RunConfig{}, "claude", args...)
	if result.Err != nil {
		return nil, fmt.Errorf("claude: %w\nstderr: %s", result.Err, result.Stderr)
	}

	return extractClaudeResult(result.Stdout)
}

// GenerateScenarios sends the mandate to the claude CLI and parses the returned JSON.
func (c *ClaudeCLI) GenerateScenarios(mandate string) ([]scenario.Scenario, error) {
	out, err := c.Prompt(mandate)
	if err != nil {
		return nil, err
	}
	return ParseScenariosFromJSON(out)
}
