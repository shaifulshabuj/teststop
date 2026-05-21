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

// openAIAdapter implements Adapter against the OpenAI Chat Completions
// API. As with claude.go we speak HTTP directly to avoid an SDK.
type openAIAdapter struct {
	apiKey  string
	model   string
	client  *http.Client
	baseURL string
}

func newOpenAI(cfg Config) (Adapter, error) {
	key, err := requireKey(cfg, "OPENAI_API_KEY")
	if err != nil {
		return nil, err
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("TESTSTOP_MODEL")
	}
	if model == "" {
		model = DefaultOpenAIModel
	}
	base := os.Getenv("OPENAI_BASE_URL")
	if base == "" {
		base = "https://api.openai.com"
	}
	return &openAIAdapter{
		apiKey:  key,
		model:   model,
		client:  &http.Client{Timeout: cfg.Timeout},
		baseURL: strings.TrimRight(base, "/"),
	}, nil
}

func (o *openAIAdapter) Name() string {
	return "openai:" + o.model
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float32         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (o *openAIAdapter) Generate(ctx context.Context, mandate string) ([]scenario.Scenario, error) {
	body, err := json.Marshal(openAIRequest{
		Model:       o.model,
		Temperature: 0.7,
		MaxTokens:   8000,
		Messages: []openAIMessage{
			// Keep the mandate intact as a single user message; using
			// a separate system message is tempting but the mandate is
			// already authoritative — splitting weakens it.
			{Role: "user", Content: mandate},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("encode openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		o.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build openai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call openai: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read openai response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("openai http %d: %s", resp.StatusCode, truncate(string(raw), 600))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse openai response: %w (%s)", err, truncate(string(raw), 400))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai error: %s: %s", parsed.Error.Type, parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	text := parsed.Choices[0].Message.Content
	if strings.TrimSpace(text) == "" {
		return nil, ErrEmptyResponse
	}
	return parseScenarios(text)
}
