package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// claudeAdapter implements Adapter against Anthropic's Messages API.
//
// We deliberately speak the raw HTTP API rather than depending on an
// SDK. The mandate is a single message, the response is a single block
// of text, and pulling in a versioned SDK would conflict with the
// project's "thin trigger" principle.
type claudeAdapter struct {
	apiKey  string
	model   string
	client  *http.Client
	baseURL string
}

func newClaude(cfg Config) (Adapter, error) {
	key, err := requireKey(cfg, "ANTHROPIC_API_KEY")
	if err != nil {
		return nil, err
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("TESTSTOP_MODEL")
	}
	if model == "" {
		model = DefaultClaudeModel
	}
	base := os.Getenv("ANTHROPIC_BASE_URL")
	if base == "" {
		base = "https://api.anthropic.com"
	}
	return &claudeAdapter{
		apiKey:  key,
		model:   model,
		client:  &http.Client{Timeout: cfg.Timeout},
		baseURL: strings.TrimRight(base, "/"),
	}, nil
}

func (c *claudeAdapter) Name() string {
	return "claude:" + c.model
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeResponse struct {
	Content []claudeContent `json:"content"`
	Error   *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *claudeAdapter) Generate(ctx context.Context, mandate string) ([]scenario.Scenario, error) {
	body, err := json.Marshal(claudeRequest{
		Model:     c.model,
		MaxTokens: 8000,
		// The mandate is intentionally given as the user turn. We do
		// not use a separate system message: the mandate is the system.
		Messages: []claudeMessage{{Role: "user", Content: mandate}},
	})
	if err != nil {
		return nil, fmt.Errorf("encode claude request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build claude request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call claude: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read claude response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("claude http %d: %s", resp.StatusCode, truncate(string(raw), 600))
	}

	var parsed claudeResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse claude response: %w (%s)", err, truncate(string(raw), 400))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("claude error: %s: %s", parsed.Error.Type, parsed.Error.Message)
	}

	text := concatClaudeText(parsed.Content)
	if strings.TrimSpace(text) == "" {
		return nil, ErrEmptyResponse
	}
	return parseScenarios(text)
}

func concatClaudeText(blocks []claudeContent) string {
	var b strings.Builder
	for _, c := range blocks {
		if c.Type == "text" {
			b.WriteString(c.Text)
		}
	}
	return b.String()
}
