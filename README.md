# teststop

> Trigger AI to test your software the way a real adversarial user would break it.

[![Documentation](https://img.shields.io/badge/docs-shaifulshabuj.github.io%2Fteststop-6a0dad?style=flat-square)](https://shaifulshabuj.github.io/teststop/)
[![Go Reference](https://pkg.go.dev/badge/github.com/shaifulshabuj/teststop.svg)](https://pkg.go.dev/github.com/shaifulshabuj/teststop)

**[Full Documentation →](https://shaifulshabuj.github.io/teststop/)**

---

## The Problem

Every test ever written was written by someone who *knew* how the system works.

Real users don't know. That's why production still surprises us.

We call it test coverage. What we actually have is **assumption coverage**.

teststop solves this by giving AI a mandate to think like a real adversarial user — someone who never read the docs, retries when things are slow, opens the same form in two tabs, pastes unexpected data, and does what no spec ever imagined.

AI already knows all of this. It just needed the right instruction. teststop provides that instruction.

---

## Install

```bash
go install github.com/shaifulshabuj/teststop/cmd/teststop@latest
```

Or build from source:

```bash
git clone https://github.com/shaifulshabuj/teststop
cd teststop
go build -o teststop ./cmd/teststop
```

**Requirements:** [ollama](https://ollama.com) (default, free) **or** `claude`/`copilot` CLI on your PATH. No API keys needed for local runs.

---

## Usage

```bash
# Run on the current directory
teststop run

# Run on a specific path
teststop run --path ./src

# Control testing depth
teststop run --depth aggressive   # light | normal | aggressive

# Execute scenarios against a RUNNING system
teststop run --target http://localhost:8080

# Machine-readable output for AI agents
teststop run --output json --no-color --quiet

# Show what teststop knows about your project
teststop status

# Show accumulated testing memory
teststop memory

# Show the exact mandate sent to AI
teststop mandate --show

# Show version and build info
teststop version          # or: teststop --version / teststop -v
```

### Execute Against a Live System

By default teststop **generates** adversarial scenarios and validates them
structurally. Point it at a running system with `--target` and it will also
**execute** those scenarios and feed the real pass/fail outcomes into
confidence memory. teststop tests whatever is already running at that URL —
local, staging, or a production-like instance — it never starts or manages the
app itself (that stays your job).

Execution is **hybrid**, chosen per scenario:

| Condition | Executor | Behavior |
|-----------|----------|----------|
| Scenario has a structured `exec` block + `--target` set | **HTTP** | Deterministic `net/http` request, judged on status code |
| `--target` set, prose-only scenario | **AI-driven** | The AI actually performs the steps against the target and returns a verdict |
| No `--target` | **Static** | Structural validation only (default) |

```bash
# Local app — run the AI tester directly (a sandboxed container can't reach host localhost)
TESTSTOP_SANDBOX=none teststop run --target http://localhost:8080

# Staging / prod-like URL — fully sandboxed AI tester works
TESTSTOP_SANDBOX=required teststop run --target https://staging.example.com

# Tune execution
teststop run --target http://localhost:8080 \
  --concurrency 8 --exec-timeout 15s --max-retries 3
```

For **race conditions** (double-submit, claim-the-last-item), a scenario's `exec`
block can set `concurrency: N` — teststop fires N identical requests at once and
asserts exactly one wins. Runs **without** `--target` are clearly labelled
**predicted** (a risk surface), not verified failures.

A failed `critical` scenario sets exit code `2`. See
[Execution](https://shaifulshabuj.github.io/teststop/guide/execution/) for details.

### Exit Codes (for AI agents)

| Code | Meaning | Action |
|------|---------|--------|
| `0` | Confidence threshold met | Safe to deploy |
| `1` | Below threshold | Review required |
| `2` | Critical failures found | Do NOT deploy |
| `3` | teststop internal error | Debug teststop |

---

## How the Mandate Works

The **mandate** is the instruction that makes AI test like a real adversarial user rather than a developer.

teststop composes a mandate from:
1. **Base mandate** — universal adversarial user thinking patterns
2. **Project context** — what teststop learned by scanning your code
3. **Memory** — what areas are already proven stable (test less) vs. volatile (test more)

Run `teststop mandate --show` to see the exact text sent to the AI.

The mandate is in `mandate/base.md` — readable, editable, community-improvable.

See [MANDATE.md](./MANDATE.md) for the full philosophy.

---

## How Memory Works (Tests Reduce Over Time)

teststop maintains a confidence score per system area in `.teststop/memory.json`.

- **Passes increase confidence** (+0.19 per pass)
- **Failures decrease confidence** (-0.30 per failure)
- **Areas at 0.95+ confidence are retired** — teststop stops testing them aggressively
- **Changed areas lose confidence** — teststop focuses back on them

After ~15 clean passes, an area is considered proven stable. teststop moves on.

**Commit `.teststop/memory.json` to version control** — it's the accumulated proof that your system works.

---

## AI Backend

teststop supports three AI backends. The default is **ollama** — free, local, and unlimited.

| Backend | How to use | Quality | Cost |
|---------|-----------|---------|------|
| **ollama** (default) | `ollama serve` + `ollama pull qwen3.6:latest` | Very good | Free |
| claude | `TESTSTOP_CLI=claude teststop run` | Best | Account quota |
| copilot | `TESTSTOP_CLI=copilot teststop run` | Good | Subscription |

```bash
# Quick-start with local model
brew install ollama
ollama pull qwen3:4b   # 2.5 GB, fast; or qwen3.6:latest for best quality
ollama serve
teststop run           # auto-detects ollama
```

Auto-detection order: **ollama → claude → copilot**. Set `TESTSTOP_CLI=claude` to always
use Claude regardless of ollama availability.

See [AI Adapters](https://shaifulshabuj.github.io/teststop/guide/ai-adapters/) for model
comparison, troubleshooting, and the quality tradeoff between local and cloud backends.

---

## Using with Claude Code / Copilot

Add to your agent's workflow:

```bash
# After modifying code, run teststop
teststop run --output json --quiet
# Exit 0: safe to proceed. Exit 1: review. Exit 2: stop.
```

teststop is designed to be invoked by AI coding agents, not just humans.

---

## Ecosystem

```
DocuFlow  → gives AI the context to act with purpose
Waymark   → gives humans the reason to trust and step back
teststop  → gives systems the confidence to prove themselves
```

- [DocuFlow](https://github.com/shaifulshabuj/docuflow-mcp) — MCP server implementing the LLM Wiki pattern
- [Waymark](https://github.com/shaifulshabuj/waymark) — MCP middleware for AI agent governance
- **teststop** — adversarial user testing trigger (this repo)

---

## Contributing

The mandate (`mandate/base.md`) is the most valuable thing to improve. If you've seen a failure pattern that teststop didn't catch, improve the mandate.

See [MANDATE.md](./MANDATE.md) for contribution guidelines.

---

*teststop v0.3.1*
