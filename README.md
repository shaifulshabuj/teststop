# teststop

> Trigger AI to test any software system the way a real adversarial user would break it.

teststop is an agent-native CLI. It is not a test runner, not a framework, and not a replacement for testers. It is a thin trigger that gives AI the right mandate — and then gets out of the way.

> *"Tests should reduce over time. Not grow forever."*

---

## The philosophy

Every test ever written was written by someone who already knew how the system works. That is **assumption coverage** — and it is why production still surprises us.

Real users do not know how the system was built. They retry, they paste, they abandon, they open two tabs, they ignore the warning the first time. AI already knows all of this. It just needed the right instruction.

teststop provides that instruction. We call it the [**Mandate**](./MANDATE.md). Run `teststop mandate --show` to read it any time.

---

## Status

**Pre-release — v0.1.0-dev.** End-to-end pipeline is wired: `teststop run` scans the project, composes the mandate, calls Anthropic Claude (or OpenAI), parses the scenarios, updates memory, and emits a report. Scenarios are generated, not yet executed — that lands in v0.2. See *Roadmap* below.

---

## Install

While the binary release is pending, build from source:

```bash
git clone https://github.com/shaifulshabuj/teststop
cd teststop
go mod tidy
go build -o teststop ./cmd/teststop
./teststop --help
```

Requires Go 1.22+.

---

## Use it

```bash
# read the mandate the AI will receive
teststop mandate --show

# run a full adversarial-user cycle on the current project
teststop run                       # text output, default
teststop run --output json         # agent-parseable
teststop run --depth aggressive    # wider scenario net
teststop run --dry-run             # print the composed mandate without calling an AI

# inspect what teststop has learned about this project
teststop status
teststop memory
teststop memory --reset            # clear all learning

# render the last run in a different format
teststop run --report-dir .teststop && teststop report --format markdown
```

### Configuration

teststop is zero-config by default — drop in, run, get scenarios. Two env vars wire the AI:

| variable               | meaning                                                       |
|------------------------|---------------------------------------------------------------|
| `ANTHROPIC_API_KEY`    | Anthropic key. Auto-selects Claude when present.              |
| `OPENAI_API_KEY`       | OpenAI key. Used when no Anthropic key is set, or via flag.   |
| `TESTSTOP_AI`          | `claude` or `openai` to override auto-detection.              |
| `TESTSTOP_MODEL`       | Override the provider's default model.                        |
| `ANTHROPIC_BASE_URL`   | Override the Anthropic API base (proxy/testing).              |
| `OPENAI_BASE_URL`      | Override the OpenAI API base (proxy/testing).                 |

`teststop run` exits with:

| code | meaning                                    |
|------|--------------------------------------------|
| 0    | confidence threshold met — safe to proceed |
| 1    | below threshold — review required          |
| 2    | critical failures — do not deploy          |
| 3    | teststop internal error                    |

---

## The mandate

The mandate (`mandate/base.md`) is the entire instruction the AI receives. It is plain Markdown, embedded into the binary, and printable on demand. It is open by design so anyone can audit, fork, and improve it.

Read [`MANDATE.md`](./MANDATE.md) for the philosophy and the contribution guidelines.

---

## How memory works

teststop writes a `.teststop/` directory at the project root. It is committed to version control. Each run updates per-area confidence: stable areas get tested less; volatile areas get tested harder. Tests are retired automatically when confidence crosses a conservative threshold (default 0.95).

> The success condition is that you need teststop **less** over time — not more.

---

## Using teststop with Claude Code, Copilot, and other agents

teststop is agent-native by design. The default output mode is structured JSON, with deterministic exit codes for deploy gates. Any AI coding agent can invoke it, parse the result, and make a deploy / no-deploy decision without a human in the loop.

Suggested wiring:

```bash
teststop run --output json --quiet --no-color
```

---

## Roadmap (v0.1 MVP)

- [x] **Step 1** — Cobra scaffold and CLI surface
- [x] **Step 2** — The mandate (`mandate/base.md`) and the scenario contract
- [x] **Step 3** — Reader: detect language, type, entry points, key flows
- [x] **Step 4** — Memory: read/write `.teststop/`, confidence scoring, retirement
- [x] **Step 5** — AI Adapter: Anthropic Claude (primary), OpenAI (fallback)
- [x] **Step 6** — Composer: assemble mandate with project context and memory
- [x] **Step 7** — Reporter: JSON, text, and Markdown output
- [x] **Step 8** — Wire `teststop run` end-to-end

Next (v0.2): dynamic scenario execution, per-scenario pass/fail verdicts feeding memory.
Explicit non-goals for v0.1: dynamic test execution, Waymark/DocuFlow integration, local models, Windows support.

---

## The ecosystem

teststop is the third tool in a small set:

| tool                                                              | role                                                                 |
|-------------------------------------------------------------------|----------------------------------------------------------------------|
| [DocuFlow](https://github.com/shaifulshabuj/docuflow-mcp)         | Captures the *why* — the system's living theory, for AI agents.      |
| [Waymark](https://github.com/shaifulshabuj/waymark)               | Makes agent actions transparent — governance and audit for AI work.  |
| **teststop**                                                      | Validates reality — self-reducing, adversarial-user confidence.      |

> *Intent in. Value out. Nothing in between.*

---

## License

To be decided before the first tagged release.

---

*Author: Shaiful Shabuj*
