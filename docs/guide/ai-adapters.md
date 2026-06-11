# AI Adapters

teststop generates scenarios by calling an AI backend. The default is **ollama** (free,
local, unlimited). Cloud CLIs (claude, copilot) are opt-in.

---

## How Adapters Work

The `AIAdapter` interface:

```go
type AIAdapter interface {
    GenerateScenarios(mandate string) ([]scenario.Scenario, error)
    Prompt(input string) ([]byte, error)
    Name() string
}
```

teststop passes the full composed mandate as a single prompt and parses the JSON array
response. All adapters share the same JSON parser (`ParseScenariosFromJSON`) which strips
markdown fences and rejects hollow batches.

---

## Auto-Detection

`TESTSTOP_CLI=auto` (the default) tries backends in this order:

| Priority | Backend | Detection |
|----------|---------|-----------|
| 1 | **ollama** | `localhost:11434` responds within 2s |
| 2 | claude | `claude` found on PATH |
| 3 | copilot | `copilot` found on PATH |

The `TESTSTOP_CLI` env var overrides auto-detection:

| `TESTSTOP_CLI` | Behavior |
|----------------|----------|
| `auto` _(default)_ | ollama → claude → copilot |
| `ollama` | Use ollama only; error if not reachable |
| `claude` | Use Claude CLI only; error if not found |
| `copilot` | Use GitHub Copilot CLI only; error if not found |

```bash
# Use local model (default when ollama is running)
teststop run

# Force Claude (opt-in — uses account quota)
TESTSTOP_CLI=claude teststop run

# Force a specific ollama model
TESTSTOP_MODEL=qwen3:4b teststop run
```

**Quality tradeoff:** Local models (ollama) produce valid, useful scenarios at the cost of
some specificity and edge-case creativity versus cloud models. For production use where quota
is not a concern, `TESTSTOP_CLI=claude` gives higher-quality output. For development
iteration and free CI runners, ollama is the right default.

---

## Ollama (default)

Calls the [ollama](https://ollama.com) HTTP API at `localhost:11434`. No subprocess,
no API key, no quota.

**Default model:** `qwen3.6:latest` (36B, Q4_K_M — high-quality, ~3–4 min per run)

**Model selection:**

```bash
# Use a smaller, faster model
TESTSTOP_MODEL=qwen3:4b teststop run

# Use gemma4
TESTSTOP_MODEL=gemma4:latest teststop run
```

| Model | Size | Speed | Scenario Quality |
|-------|------|-------|-----------------|
| `qwen3.6:latest` | 23 GB | ~3–4 min | Best local quality |
| `qwen3:4b` | 2.5 GB | ~30–60s | Good for quick iteration |
| `gemma4:latest` | 9.6 GB | ~1–2 min | Good general quality |
| `llama3.2:latest` | 2 GB | ~20–40s | Basic; less context-aware |

**Timeout:** 10 minutes (local inference is slower than cloud)

**Request shape:**

```json
{
  "model": "qwen3.6:latest",
  "prompt": "<mandate + JSON-only constraint>",
  "stream": false,
  "think": false,
  "options": {"num_ctx": 32768}
}
```

`think: false` disables the reasoning chain for qwen3 models. Even if the server ignores
this (older ollama builds), the adapter strips `<think>...</think>` blocks automatically
before parsing.

**Quick-start:**

```bash
# Install ollama and pull a model
brew install ollama
ollama pull qwen3:4b        # 2.5 GB, fast
ollama serve                 # starts the server

# teststop auto-detects it
teststop run
```

---

## Claude CLI

Uses the official [Claude Code](https://claude.ai/download) CLI. Requires an active
Claude subscription — runs consume account quota shared with all Claude Code agents.

**Command executed:**

```bash
claude -p "<mandate text>" --output-format json
# With model override:
claude -p "<mandate text>" --output-format json --model claude-sonnet-4-6
```

**Model selection:**

```bash
TESTSTOP_MODEL=claude-opus-4-8 TESTSTOP_CLI=claude teststop run
```

**Timeout:** 5 minutes

---

## GitHub Copilot CLI

Uses the [GitHub Copilot CLI](https://docs.github.com/en/copilot/github-copilot-in-the-cli).

**Command executed:**

```bash
copilot -p "<mandate text>" -s --no-ask-user
```

**Timeout:** 5 minutes

---

## JSON Parsing

All adapters use the same parser (`ParseScenariosFromJSON`):

1. Strip leading/trailing markdown fences (` ```json ` … ` ``` `)
2. Unmarshal into `[]scenario.Scenario`
3. Reject hollow batches (every object missing `scenario_id` and `title`) — catches
   event-stream format mismatches and empty-struct slippage

---

## Troubleshooting

**"no AI backend found"**
: Start ollama (`ollama serve`), or install `claude`/`copilot` and add to PATH.

**ollama: HTTP request failed**
: ollama is not running. Start it with `ollama serve`.

**"model not found"**
: Pull the model first: `ollama pull qwen3.6:latest`

**"all N parsed scenarios are hollow"**
: The model returned a non-scenario JSON structure. Check that the model is responding
  to the prompt (run `ollama run qwen3:4b "say hello"` to verify).

**"failed to parse scenarios JSON"**
: The AI returned non-JSON output. With cloud adapters, authenticate first:
  `claude auth` or `copilot auth`. With ollama, try a model with a larger context:
  `TESTSTOP_MODEL=qwen3.6:latest teststop run`.

**Timeout**
: For large projects, try `--depth light` to reduce mandate size. With ollama and
  qwen3.6:latest, a normal-depth run takes ~3–4 minutes.
