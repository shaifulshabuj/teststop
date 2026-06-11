# AI Adapters

teststop works by shelling out to an AI CLI already on your PATH. No API SDK, no API keys, no vendor lock-in.

---

## How Adapters Work

The `AIAdapter` interface has two methods:

```go
type AIAdapter interface {
    GenerateScenarios(mandate string) ([]scenario.Scenario, error)
    Name() string
}
```

teststop passes the full composed mandate as a single prompt argument and parses the JSON response. Both adapters have a 5-minute timeout.

---

## Auto-Detection

teststop auto-detects which CLI is available. The `TESTSTOP_CLI` environment variable controls this:

| `TESTSTOP_CLI` | Behavior |
|----------------|----------|
| `auto` _(default)_ | Try `claude` first, then `copilot` |
| `claude` | Use Claude CLI only; error if not found |
| `copilot` | Use GitHub Copilot CLI only; error if not found |

```bash
# Force Claude
TESTSTOP_CLI=claude teststop run

# Force Copilot
TESTSTOP_CLI=copilot teststop run
```

---

## Claude CLI

Uses the official [Claude Code](https://claude.ai/download) CLI.

**Command executed:**

```bash
claude -p "<mandate text>"
# With optional model override:
claude -p "<mandate text>" --model claude-sonnet-4-6
```

**Model selection:**

```bash
# Use a specific Claude model
TESTSTOP_MODEL=claude-opus-4-7 teststop run
```

If `TESTSTOP_MODEL` is not set, the Claude CLI uses its default model.

**Timeout:** 5 minutes

---

## GitHub Copilot CLI

Uses the [GitHub Copilot CLI](https://docs.github.com/en/copilot/github-copilot-in-the-cli).

**Command executed:**

```bash
copilot -p "<mandate text>" -s --no-ask-user
```

The `-s` flag enables structured output mode. `--no-ask-user` prevents interactive prompts.

**Timeout:** 5 minutes

---

## JSON Parsing

The adapter expects the AI to return a JSON array of scenario objects. Both adapters use the same parser:

1. Strip any leading/trailing markdown fences (` ```json ` … ` ``` `)
2. Unmarshal the JSON into `[]scenario.Scenario`

If parsing fails, teststop returns an error with the raw output for debugging.

---

## Checking Which Adapter Is Active

```bash
teststop status
```

The status output shows which AI adapter was used in the last run.

Or peek at the last report:

```bash
teststop report
```

---

## Future Adapters

The `AIAdapter` interface is designed to be extensible.

- **Ollama** (local models) — planned in [issue #30](https://github.com/shaifulshabuj/teststop/issues/30). Not yet implemented; `TESTSTOP_CLI=ollama` is not a valid value today.
- **OpenAI CLI** — community contribution welcome.

To add a new adapter, implement the `AIAdapter` interface in `internal/ai/` and register it in `Detect()`.

---

## Troubleshooting

**"no AI CLI found on PATH"**
: Install `claude` or `copilot` and make sure it's on your `PATH`. Verify with `which claude`.

**"AI CLI returned empty output"**
: The AI CLI may have failed silently. Run the CLI manually:
  ```bash
  claude -p "Return the word hello"
  ```

**"failed to parse scenarios JSON"**
: The AI returned non-JSON output. This can happen if the AI CLI prompts for auth.
  Authenticate your CLI first: `claude auth` or `copilot auth`.

**Timeout**
: The AI CLI has a 5-minute timeout. For very large projects, try `--depth light` to reduce the mandate size.
