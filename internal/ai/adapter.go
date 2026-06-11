package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// invalidEscapeRE matches JSON-illegal escape sequences emitted by some local
// models (e.g. gemma4 produces \x41 hex notation). Valid JSON escape suffixes
// are: " \ / b f n r t u — everything else is invalid.
var invalidEscapeRE = regexp.MustCompile(`\\([^"\\\/bfnrtu])`)

// AIAdapter is the interface all AI backends implement.
type AIAdapter interface {
	// GenerateScenarios sends the mandate to the AI and returns parsed scenarios.
	GenerateScenarios(mandate string) ([]scenario.Scenario, error)
	// Prompt sends an arbitrary instruction to the AI and returns its raw stdout.
	// Used for AI-driven scenario execution (v0.2 executor).
	Prompt(input string) ([]byte, error)
	// Name returns the adapter name (e.g., "claude", "copilot").
	Name() string
}

// Detect auto-detects which AI backend is available.
// Respects TESTSTOP_CLI env var: "ollama", "claude", "copilot", "auto" (default).
//
// Auto-detection precedence: ollama (localhost:11434) → claude → copilot.
// ollama is preferred because local-model runs are free and unlimited; cloud
// CLIs (claude, copilot) share account quota with the whole agent team.
// To opt in to claude: TESTSTOP_CLI=claude.
func Detect() (AIAdapter, error) {
	cli := os.Getenv("TESTSTOP_CLI")
	if cli == "" {
		cli = "auto"
	}

	switch cli {
	case "ollama":
		if !IsOllamaAvailable() {
			return nil, fmt.Errorf("ai: TESTSTOP_CLI=ollama but ollama not reachable at %s", ollamaDefaultBaseURL)
		}
		return NewOllamaCLI(), nil
	case "claude":
		if _, err := exec.LookPath("claude"); err != nil {
			return nil, fmt.Errorf("ai: TESTSTOP_CLI=claude but claude not found on PATH")
		}
		return NewClaudeCLI(), nil
	case "copilot":
		if _, err := exec.LookPath("copilot"); err != nil {
			return nil, fmt.Errorf("ai: TESTSTOP_CLI=copilot but copilot not found on PATH")
		}
		return NewCopilotCLI(), nil
	default: // auto
		if IsOllamaAvailable() {
			return NewOllamaCLI(), nil
		}
		if _, err := exec.LookPath("claude"); err == nil {
			return NewClaudeCLI(), nil
		}
		if _, err := exec.LookPath("copilot"); err == nil {
			return NewCopilotCLI(), nil
		}
		return nil, fmt.Errorf("ai: no AI backend found (start ollama, install claude or copilot, or set TESTSTOP_CLI)")
	}
}

// ParseScenariosFromJSON parses a JSON array of scenarios from raw bytes.
// It is tolerant of leading/trailing whitespace, markdown code fences, and
// prose preamble (local models sometimes emit reasoning text before the array).
//
// After parsing, it validates that the batch is not entirely hollow: if every
// scenario has an empty scenario_id AND empty title, the input was almost
// certainly an AI CLI event stream or other non-scenario JSON, and an error is
// returned so the caller can surface the mismatch rather than silently counting
// placeholder objects.
func ParseScenariosFromJSON(data []byte) ([]scenario.Scenario, error) {
	// Strip markdown code fences if present.
	s := strings.TrimSpace(string(data))
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// Three-pass parse strategy to handle sloppy local-model output:
	// 1. Direct unmarshal (fast path — cloud CLIs, well-behaved models).
	// 2. Extract JSON array from prose-wrapped output (qwen3:4b-style preamble).
	// 3. Sanitize invalid escape sequences then retry (gemma4 emits \xNN hex).
	var scenarios []scenario.Scenario
	if err := json.Unmarshal([]byte(s), &scenarios); err != nil {
		// Pass 2: extract [...] from within prose.
		candidate := s
		if extracted, ok := extractJSONArray(s); ok {
			candidate = extracted
			if err2 := json.Unmarshal([]byte(candidate), &scenarios); err2 == nil {
				goto validated
			}
		}
		// Pass 3: sanitize invalid escape sequences and retry.
		sanitized := sanitizeJSONEscapes(candidate)
		if err3 := json.Unmarshal([]byte(sanitized), &scenarios); err3 != nil {
			return nil, fmt.Errorf("ai: failed to parse scenarios JSON: %w\nraw output: %s", err, truncate(data, 500))
		}
	}
validated:

	// Defense-in-depth: if every parsed scenario is hollow (missing both
	// scenario_id and title) the input was not a valid scenario array — most
	// likely the raw AI CLI event stream was fed here instead of the inner
	// result string.  Fail loudly rather than silently returning empty structs.
	if len(scenarios) > 0 {
		hollow := 0
		for _, sc := range scenarios {
			if sc.ScenarioID == "" && sc.Title == "" {
				hollow++
			}
		}
		if hollow == len(scenarios) {
			return nil, fmt.Errorf("ai: all %d parsed scenarios are hollow (empty scenario_id and title) — probable AI CLI output format mismatch; raw: %s", len(scenarios), truncate(data, 300))
		}
	}

	return scenarios, nil
}

// sanitizeJSONEscapes removes invalid JSON escape sequences emitted by some local
// models. It replaces \X (where X is not a valid JSON escape character) with just X,
// preserving the string content while making the JSON parseable.
// Common cases: gemma4 emits \x41 (hex notation), models emit \' (unnecessary).
func sanitizeJSONEscapes(s string) string {
	return invalidEscapeRE.ReplaceAllString(s, "$1")
}

// extractJSONArray scans s for the first '[' and finds its matching ']' using
// a simple bracket counter. Returns the extracted substring and true on success.
// This handles local models that emit prose before the JSON array.
func extractJSONArray(s string) (string, bool) {
	start := strings.Index(s, "[")
	if start < 0 {
		return "", false
	}
	depth := 0
	inStr := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inStr {
			escaped = true
			continue
		}
		if c == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		if c == '[' {
			depth++
		} else if c == ']' {
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

func truncate(b []byte, max int) string {
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
