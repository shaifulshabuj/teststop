// Package ai is the thin shim between teststop and an LLM provider.
//
// teststop does not own a model; it owns the mandate. The adapter's job
// is to ship the mandate, take the response, parse it into scenarios,
// and stay out of the way. Provider-specific quirks live in their own
// files (claude.go, openai.go) behind a single interface.
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// Adapter is the contract every provider implements. New providers add
// a file and a constructor; nothing else in teststop needs to change.
type Adapter interface {
	// Name identifies the provider in logs and reports.
	Name() string
	// Generate sends the composed mandate to the model and returns the
	// scenarios parsed from its response. Implementations must not
	// retry forever — a single attempt with a clear error beats a
	// silent loop.
	Generate(ctx context.Context, mandate string) ([]scenario.Scenario, error)
}

// Config captures the small set of choices a caller can make. Empty
// fields fall back to environment variables, which fall back to
// conservative defaults.
type Config struct {
	Provider string        // claude | openai
	APIKey   string        // overrides env
	Model    string        // overrides env / default
	Timeout  time.Duration // per request
}

// Default model identifiers per provider. These are chosen for
// reasoning quality on long-form structured output, which is what the
// mandate produces. Override via TESTSTOP_MODEL for experimentation.
const (
	DefaultClaudeModel = "claude-opus-4-7"
	DefaultOpenAIModel = "gpt-4o"
	DefaultTimeout     = 90 * time.Second
)

// New chooses a provider from cfg and environment.
//
// Resolution order, highest priority first:
//
//	1. cfg.Provider explicitly set
//	2. TESTSTOP_AI=claude|openai
//	3. ANTHROPIC_API_KEY present → claude
//	4. OPENAI_API_KEY present    → openai
func New(cfg Config) (Adapter, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = strings.ToLower(strings.TrimSpace(os.Getenv("TESTSTOP_AI")))
	}
	if provider == "" {
		switch {
		case os.Getenv("ANTHROPIC_API_KEY") != "":
			provider = "claude"
		case os.Getenv("OPENAI_API_KEY") != "":
			provider = "openai"
		default:
			return nil, fmt.Errorf("no AI provider configured: set ANTHROPIC_API_KEY or OPENAI_API_KEY, or pass --provider")
		}
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}

	switch provider {
	case "claude", "anthropic":
		return newClaude(cfg)
	case "openai":
		return newOpenAI(cfg)
	default:
		return nil, fmt.Errorf("unknown AI provider %q", provider)
	}
}

// parseScenarios pulls a `[]Scenario` out of an LLM response. Models
// sometimes wrap JSON in markdown fences despite the mandate's rules,
// so we tolerate fences but reject anything else. If parsing fails we
// return the raw response in the error so the user can see what we got.
func parseScenarios(raw string) ([]scenario.Scenario, error) {
	cleaned := stripFences(strings.TrimSpace(raw))

	var scenarios []scenario.Scenario
	if err := json.Unmarshal([]byte(cleaned), &scenarios); err == nil {
		return scenarios, nil
	}

	// Some models occasionally emit a single object instead of an array
	// when asked to generate "one" scenario; accept that too.
	var one scenario.Scenario
	if err := json.Unmarshal([]byte(cleaned), &one); err == nil && one.ScenarioID != "" {
		return []scenario.Scenario{one}, nil
	}

	return nil, fmt.Errorf("ai response was not valid scenario JSON:\n%s", truncate(cleaned, 1200))
}

// stripFences removes leading/trailing ```json fences if the model
// ignored the mandate's "no markdown fences" rule.
func stripFences(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	// Drop first fence line.
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimRight(s, "` \n\t")
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// requireKey is a small helper used by the provider constructors.
func requireKey(cfg Config, envName string) (string, error) {
	if strings.TrimSpace(cfg.APIKey) != "" {
		return cfg.APIKey, nil
	}
	if v := os.Getenv(envName); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("%s not set", envName)
}

// ErrEmptyResponse signals a successful API call that returned no usable
// text. Surfaced separately so callers can decide whether to retry.
var ErrEmptyResponse = errors.New("ai: empty response from provider")
