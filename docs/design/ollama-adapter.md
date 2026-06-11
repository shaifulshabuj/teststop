# Design: ollama Adapter (internal/ai/ollamacli.go)

**Date:** 2026-06-12  
**Issue:** #30  
**Branch:** feat/ollama-backend

---

## Problem

teststop's default backend (claude CLI) burns the same account quota as the entire
agent team. Local model runs are free and unlimited. ollama (already installed,
qwen3.6:latest preferred) becomes the default; claude/copilot become opt-in.

---

## Interface Fit

`OllamaCLI` implements `AIAdapter` identically to `ClaudeCLI` and `CopilotCLI`:

```go
type OllamaCLI struct {
    baseURL string  // default: http://localhost:11434
    model   string  // default: qwen3.6:latest
}

func (o *OllamaCLI) Name() string
func (o *OllamaCLI) Prompt(input string) ([]byte, error)
func (o *OllamaCLI) GenerateScenarios(mandate string) ([]scenario.Scenario, error)
```

---

## Invocation: HTTP API (not `ollama run`)

**Decision: use the ollama HTTP API at `localhost:11434/api/generate`.**

`ollama run <model>` is an interactive REPL — subprocess lifecycle, stdin/stdout
management, no clean structured output path. The HTTP API is the correct interface:
JSON in, JSON out, `stream: false` for a single blocking response.

```
POST http://localhost:11434/api/generate
{
  "model":  "qwen3.6:latest",
  "prompt": "<mandate + JSON-only suffix>",
  "stream": false,
  "think":  false,
  "options": {"num_ctx": 32768}
}
```

Response field: `response` (string).

`think: false` disables the `<think>...</think>` reasoning chain for qwen3 family
models (supported in ollama ≥0.3). Even when the server ignores the flag, the
`<think>` block stripper catches residual tokens.

`num_ctx: 32768` — ollama defaults to 2048 context; the mandate can be 4-6k tokens.
32768 is large enough for any realistic mandate without loading the full 262k window.

---

## Model Selection

`TESTSTOP_MODEL` env var, same as claude. Default: `qwen3.6:latest`.

Fallback models (documented, not auto-tried): `qwen3:4b`, `gemma4:latest`,
`llama3.2:latest`. User chooses by setting `TESTSTOP_MODEL`.

---

## Detection

`IsOllamaAvailable()` pings `GET http://localhost:11434/` with a 2-second timeout.
No PATH lookup needed — the API is what matters, not the binary location.

---

## Mandate → Local Model: Sloppy Output

Local models are less instruction-following than cloud models. Two defenses:

1. **JSON-only suffix appended by `OllamaCLI.GenerateScenarios`** (not in base mandate,
   so claude/copilot are unaffected):

   ```
   IMPORTANT: Your entire response MUST be a single valid JSON array.
   Do NOT include any explanatory text, preamble, or markdown prose outside the array.
   Begin your response with [ and end with ].
   ```

2. **Existing `ParseScenariosFromJSON`** already strips markdown fences and catches
   hollow batches. No new parser logic needed.

3. **`<think>` block stripper** in `OllamaCLI.Prompt`: strip everything between
   `<think>` and `</think>` (inclusive) before handing output to the parser.

---

## Auto-Detection Precedence (updated)

`TESTSTOP_CLI=auto` now resolves: **ollama → claude → copilot**

| Priority | When active |
|----------|------------|
| 1. ollama | `localhost:11434` responds within 2s |
| 2. claude | `claude` found on PATH |
| 3. copilot | `copilot` found on PATH |

**Quality tradeoff (documented):** Local models produce valid JSON but scenario depth,
specificity of failure modes, and edge-case creativity are measurably lower than claude.
For production use where token cost is not a constraint, `TESTSTOP_CLI=claude` is
recommended. For development, iteration, and CI on free runners, ollama is the default.

---

## Sandbox

`OllamaCLI` does NOT use `sandbox.Runner`. It calls `net/http` directly to
`localhost:11434`. The ollama server is a host process; the call is network-local,
not a spawned subprocess. No credential mounts needed.

---

## Context Window

qwen3.6:latest supports 262k tokens but ollama default `num_ctx` is 2048. We set
`num_ctx: 32768` (~24k words) in every request — enough for any mandate.

---

## Real-Run Results (vs waymark API, 2026-06-12)

| Model | Scenarios | Time | Parse failures | Outcome |
|-------|-----------|------|----------------|---------|
| `qwen3.6:latest` | 31 | 184s | 0 | ✅ High quality — IDOR, race conditions, token attacks |
| `gemma4:latest` | 52 | 499s | 0 (after escape fix) | ✅ Most thorough — good specificity, 3-pass parser needed |
| `qwen3:4b` | — | — | 100% | ❌ Not viable — outputs reasoning prose only, never reaches JSON |

**Unexpected resilience fixes** required by real model behavior (both now tested):

1. **Prose preamble/suffix** (`extractJSONArray`): qwen3:4b-style models emit a planning
   monologue before the array. Two-pass: direct unmarshal → extract `[…]` span.
2. **Invalid escape sequences** (`sanitizeJSONEscapes`): gemma4 emits `\xNN` hex
   notation. Three-pass: sanitize with `\([^"\\\/bfnrtu])` → `$1` then retry.

**Model floor:** ~10B parameters. Sub-10B models spend their output budget on
reasoning rather than JSON generation — the JSON-only suffix is insufficient.

---

## Files Changed

| File | Change |
|------|--------|
| `internal/ai/ollamacli.go` | New — OllamaCLI adapter (HTTP, no sandbox) |
| `internal/ai/ollamacli_test.go` | New — 11 unit tests with httptest canned responses |
| `internal/ai/adapter.go` | `Detect()` precedence, `IsOllamaAvailable()`, three-pass parser, prose extractor, escape sanitizer |
| `internal/ai/adapter_test.go` | +4 tests: proseBeforeArray, proseAfterArray, invalidEscapes, (existing hollowArray retained) |
| `docs/guide/ai-adapters.md` | Rewritten — ollama section, model table with real numbers |
| `docs/reference/configuration.md` | TESTSTOP_CLI/MODEL updated for ollama |
| `docs/design/ollama-adapter.md` | This file |
| `README.md` | Backend section added |
| `CHANGELOG.md` | v1.1.0 with real-run quality table |
