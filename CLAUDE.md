# Claude Code — teststop Project

> Read this fully before writing any code. This is the complete project context.

## What teststop is

teststop is a CLI tool with ONE job:
**Trigger AI to test any software system the way a real adversarial user would break it.**

It is NOT a test runner. It is NOT a test framework. It is a **TRIGGER** — a thin CLI that gives AI the right mandate, then gets out of the way.

## The 6 Non-Negotiables (Design Principles)

1. **ZERO CONFIGURATION** — `teststop run` must work with no setup on any project
2. **UNIVERSAL** — must work on any language, any age, any system type
3. **SELF-REDUCING** — tests reduce over time as confidence builds. Not grow.
4. **AGENT-NATIVE** — output is machine-readable JSON for AI agent consumption
5. **NO NEW LOOP** — teststop must never become the thing that needs to be maintained
6. **EXIT CONDITION** — success = the user needs it less over time, not more

## The One Question to Ask at Every Decision Point

> "Does this make the user's problem easier to solve — or our system easier to build?"

If the answer is the second one: it is accidental complexity. **Cut it.**

---

## Architecture

```
teststop/
├── cmd/teststop/main.go         # CLI entry point (calls Execute())
├── internal/
│   ├── cli/                     # Cobra command handlers
│   │   ├── root.go              # Root command
│   │   ├── run.go               # teststop run — THE main command
│   │   ├── status.go            # teststop status
│   │   ├── memory.go            # teststop memory [--reset]
│   │   ├── report.go            # teststop report [--format md]
│   │   └── mandate.go           # teststop mandate --show
│   ├── reader/                  # Scan + understand any codebase (static)
│   │   ├── scanner.go           # Walk file tree, collect files
│   │   ├── detector.go          # Detect language, type, entry points
│   │   ├── analyzer.go          # Extract flows, routes, dependencies
│   │   └── types.go             # ProjectContext, Flow structs
│   ├── mandate/                 # Compose the AI instruction
│   │   ├── composer.go          # Inject context + memory into base mandate
│   │   └── templates/           # context.md enrichment template
│   ├── ai/                      # AI adapter layer — shells out to CLI, no SDK
│   │   ├── adapter.go           # AIAdapter interface + ParseScenariosFromJSON + Detect()
│   │   ├── claudecli.go         # `claude -p "mandate"` (Claude Code CLI)
│   │   └── copilotcli.go        # `copilot -p "mandate" -s --no-ask-user`
│   ├── memory/                  # Confidence persistence
│   │   ├── store.go             # Read/write .teststop/memory.json
│   │   ├── confidence.go        # Confidence scoring (PassWeight=0.19)
│   │   └── retire.go            # Test retirement (threshold=0.95)
│   └── reporter/                # Output formatting
│       ├── json.go              # JSON output (default, agent-parseable)
│       ├── text.go              # ANSI human-readable terminal output
│       ├── markdown.go          # Markdown report file
│       └── types.go             # RunResult, Failure structs
├── pkg/scenario/
│   └── types.go                 # Scenario struct — STABLE CONTRACT after v0.1
├── mandate/
│   ├── base.md                  # THE MANDATE — the soul of teststop ⭐
│   └── embed.go                 # //go:embed base.md
└── .teststop/                   # Runtime memory (created at first run)
    ├── memory.json              # Confidence state (commit this)
    ├── retired.json             # Retired test areas (commit this)
    ├── runs/                    # Run history (gitignored)
    └── config.yaml              # Optional project config
```

## The Priority: mandate/base.md

**This is the most important file in the entire project.**

`mandate/base.md` is the instruction that makes AI test like a real adversarial user. The quality of this file determines the quality of every test scenario generated. Get this wrong and everything else fails regardless of how good the code is. **Iterate on this constantly.**

## teststop run Pipeline

```
teststop run
  → reader.Scan(path)              # Detect language, type, flows
  → memory.Load()                  # Load .teststop/memory.json
  → mandate.Compose(context, mem)  # Build the AI instruction
  → ai.GenerateScenarios(mandate)  # Shell out to claude/copilot CLI
  → memory.Update(results)         # Update confidence scores
  → memory.RetireEligible()        # Retire areas >= 0.95 confidence
  → reporter.Output(results)       # JSON or text or markdown
  → os.Exit(exitCode)              # 0=ok, 1=review, 2=critical, 3=error
```

## Build Commands

```bash
go build ./...           # Build everything (must pass)
go test ./...            # Run all tests (must pass)
go run ./cmd/teststop    # Run CLI locally
go mod tidy              # Sync dependencies
go vet ./...             # Vet all packages
```

**Always run `go test ./...` before committing.**

## Environment Variables

```bash
TESTSTOP_CLI=auto        # auto | claude | copilot | ollama  (default: auto-detect)
TESTSTOP_MODEL=          # optional — passed as --model to claude CLI
```

**No API keys. No SDK.** teststop shells out to the AI CLI already on the user's PATH.
- `claude` CLI (Claude Code) — auto-detected first
- `copilot` CLI (GitHub Copilot) — auto-detected second
- `ollama` — opt-in via `TESTSTOP_CLI=ollama` (v0.2)

## Exit Codes

| Code | Meaning | Agent Action |
|------|---------|--------------|
| 0 | Confidence threshold met | Safe to deploy |
| 1 | Below threshold | Review required |
| 2 | Critical failures found | Do NOT deploy |
| 3 | teststop internal error | Debug teststop |

## Key Constants (internal/memory/confidence.go)

```go
RetirementThreshold = 0.95   // retire area when confidence >= this
PassWeight          = 0.19   // ~15 passes to reach 0.95 from 0.0
FailPenalty         = 0.30   // significant drop on failure
VolatileThreshold   = 0.75   // below this = volatile, full testing
StableThreshold     = 0.95   // above this = stable, reduced testing
```

**The math:** PassWeight must be ≥ 0.181 for 15 passes to reach 0.95. Use 0.19.

## CLI Commands Reference

```bash
teststop run                    # Run on current directory
teststop run --path ./src       # Run on specific path
teststop run --depth aggressive # light | normal | aggressive
teststop run --output json      # json | text | markdown  
teststop run --threshold 85     # Confidence threshold (0-100)
teststop run --no-color         # Disable ANSI (for agents)
teststop run --quiet            # Minimal output

teststop status                 # Show confidence state
teststop memory                 # Show accumulated memory
teststop memory --reset         # Clear memory (with confirmation)
teststop report                 # Show last run report
teststop report --format md     # Markdown output
teststop mandate --show         # Print exact mandate sent to AI
```

## Anti-Patterns (DO NOT BUILD in v0.1)

- ❌ Dynamic test execution (executor is v0.2)
- ❌ Waymark integration (v1.0)
- ❌ DocuFlow integration (v1.0)
- ❌ Web UI of any kind
- ❌ CI/CD plugins (v1.0)
- ❌ Windows support (macOS + Linux first)
- ❌ Any feature that requires its own configuration

## Go Patterns Used

- **Cobra** for CLI framework (`github.com/spf13/cobra`)
- **`//go:embed`** for mandate file (ship as single binary)
- **`os/exec`** for AI calls — shell out to `claude` or `copilot` CLI, no SDK
- **Interfaces** for AI adapter (`AIAdapter` — claudecli and copilotcli implement it)
- **JSON** for all memory files (human-readable, version-controllable)
- **No CGO** (`CGO_ENABLED=0`) — cross-platform single binary

## Scenario Schema (pkg/scenario/types.go) — STABLE after v0.1

```json
{
  "scenario_id": "string",
  "title": "string",
  "user_perspective": "who is this user and what do they want",
  "preconditions": ["string"],
  "steps": ["string"],
  "chaos_factors": ["slow network, bad input, etc."],
  "expected_behavior": "string",
  "failure_modes": ["string"],
  "priority": "critical | high | medium | low",
  "confidence_area": "which system area this covers",
  "is_edge_case": true
}
```

**Changes to this schema after v0.1 are breaking changes.**

## Development Container (Isolated Agent Environment)

Run Claude Code / Copilot CLI agent inside an isolated linux/arm64 container using Apple Container. The agent can only access the repo (mounted as `/workspace`) — it cannot touch the rest of your host filesystem.

### Prerequisites (macOS arm64 only)

```bash
brew install container          # Install Apple Container CLI
container system start          # Start (downloads ~100MB Linux kernel once)
```

### Launch Container

```bash
./scripts/dev-container.sh              # Interactive bash — then run `claude` inside
./scripts/dev-container.sh -- claude    # Start Claude Code agent directly
```

### What is mounted (read-only from host)

| Host path | Container path | Purpose |
|---|---|---|
| `this repo` | `/workspace` | Your code (read-write) |
| `~/.claude` | `/root/.claude` | Claude Code credentials |
| `~/.config/gh` | `/root/.config/gh` | gh CLI token |
| `~/.gitconfig` | `/root/.gitconfig` | Git identity |

**The agent cannot access:** `~/.ssh`, `~/.zshrc`, `~/.aws`, other projects, system files.

### Dockerfile.dev

`Dockerfile.dev` in the repo root. Installs: Go 1.24, gh CLI, Node.js 22, `claude` CLI, `copilot` CLI.
Build is automatic on first `./scripts/dev-container.sh` run.

---

## PRD & Philosophy References

All source-of-truth documents are in `teststop-init/`:
- `01-PHILOSOPHY.md` — why teststop exists
- `02-PROJECT-GOALS.md` — measurable goals (G1-G10)
- `03-PRD.md` — full product requirements
- `04-AGENT-STARTING-PROMPT.md` — original agent starting prompt

## Ecosystem Context

```
DocuFlow  → gives AI the context to act with purpose
Waymark   → gives humans the reason to trust and step back
teststop  → gives systems the confidence to prove themselves
```

teststop is the third tool in this trilogy.
- DocuFlow: https://github.com/shaifulshabuj/docuflow-mcp
- Waymark: https://github.com/shaifulshabuj/waymark
