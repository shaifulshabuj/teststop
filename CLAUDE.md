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
│   ├── sandbox/                 # Container isolation layer (Apple Container)
│   │   ├── detector.go          # Detect() — is `container` system running?
│   │   ├── runner.go            # Run cmd in container OR directly (auto-fallback)
│   │   └── types.go             # Mode (auto|required|none), RunConfig, Result
│   ├── ai/                      # AI adapter layer — shells out to CLI, no SDK
│   │   ├── adapter.go           # AIAdapter interface + ParseScenariosFromJSON + Detect()
│   │   ├── claudecli.go         # `claude -p "mandate"` via sandbox.Runner
│   │   └── copilotcli.go        # `copilot -p "mandate" -s --no-ask-user` via sandbox.Runner
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
├── Dockerfile.agent             # Minimal runtime image — AI CLIs only (used by sandbox)
├── Dockerfile.dev               # Full dev environment (Go + CLIs + gh)
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
  → sandbox.Detect()               # Is Apple Container available?
  → ai.GenerateScenarios(mandate)  # Run claude/copilot in sandbox (or direct)
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
TESTSTOP_SANDBOX=auto    # auto | required | none  (default: auto)
                         #   auto     = use container if available, else run direct
                         #   required = error if Apple Container not running
                         #   none     = always run AI CLI directly (no container)
```

**No API keys. No SDK.** teststop shells out to the AI CLI already on the user's PATH.

## Config File (internal/config/)

`.teststop/config.yaml` is an **optional** per-project config for `teststop run`.
It is loaded by `internal/config` and applied in `internal/cli/run.go`. Every key
maps one-to-one onto an existing `run` flag — config introduces **no new settings**
(zero-config stays the default; a missing file is never an error).

**Precedence (lowest → highest):** `config.yaml` < `TESTSTOP_RUN_*` env var < explicit CLI flag.

```yaml
# .teststop/config.yaml — all keys optional
depth: normal          # --depth        / TESTSTOP_RUN_DEPTH
output: text           # --output       / TESTSTOP_RUN_OUTPUT
threshold: 80          # --threshold    / TESTSTOP_RUN_THRESHOLD
no_color: false        # --no-color     / TESTSTOP_RUN_NO_COLOR
quiet: false           # --quiet        / TESTSTOP_RUN_QUIET
target: ""             # --target       / TESTSTOP_RUN_TARGET
concurrency: 4         # --concurrency  / TESTSTOP_RUN_CONCURRENCY
exec_timeout: 10s      # --exec-timeout / TESTSTOP_RUN_EXEC_TIMEOUT
max_retries: 2         # --max-retries  / TESTSTOP_RUN_MAX_RETRIES
```

Malformed YAML or an unknown key fails loudly. See `.teststop/config.example.yaml`.

## Runtime Sandbox (internal/sandbox/)

teststop runs the AI CLI inside an Apple Container VM when available. The AI executes inside an isolated linux/arm64 environment — it cannot access the user's host filesystem beyond the mounted project path.

```
teststop run (user's machine)
  └─ sandbox.Runner.Run(mandate)
       ├─ [container available] → container run --rm teststop-agent:latest claude -p "..."
       │       Isolated VM: AI reads mandate, outputs JSON → captured by teststop
       └─ [no container] → exec.Command("claude", "-p", mandate)  ← direct fallback
```

**sandbox.Runner logic (`internal/sandbox/runner.go`):**
```go
func (r *Runner) Run(ctx context.Context, cfg RunConfig, cmd string, args ...string) Result {
    if r.shouldUseContainer() {
        return r.runInContainer(ctx, cfg, cmd, args...) // container run --rm ...
    }
    return r.runDirect(ctx, cmd, args...)               // exec.Command(cmd, args...)
}
```

**Sandbox modes (`TESTSTOP_SANDBOX`):**
- `auto` — detect if `container system status` shows running; use if yes, direct if no
- `required` — error if container not available (strict isolation enforcement)
- `none` — always run directly (useful for CI, Docker-in-Docker, non-macOS)

**Runtime image:** `Dockerfile.agent` — minimal Ubuntu 24.04 + claude + copilot CLI only.
No Go, no gh, no dev tools. Ephemeral per-run (`--rm`). Published as `ghcr.io/shaifulshabuj/teststop-agent:latest`.

**Credential mounts (auto, read-only):**
- `~/.claude` → `/root/.claude:ro` (Claude auth)
- `~/.config/gh` → `/root/.config/gh:ro` (Copilot auth)

### Permission Map — Two Isolation Layers

**Layer 1 — Dev container** (`Dockerfile.dev`, launched by `scripts/dev-container.sh`):
The coding agent (Claude/Copilot) that *builds teststop* runs here.

| Path | Permission | Purpose |
|------|-----------|---------|
| `/workspace` | ✅ full R/W | The teststop repo — agent edits code, runs builds/tests |
| Container OS | ✅ full root | Can run any command inside the VM |
| `/root/.claude` | 🔒 read-only | Claude Code credentials — cannot be modified |
| `/root/.config/gh` | 🔒 read-only | gh CLI credentials — cannot be modified |
| Everything else | ❌ invisible | No ~/.ssh, no other projects, no host system files |

**Layer 2 — Runtime container** (`Dockerfile.agent`, spawned by `sandbox.Runner`):
The AI that *runs inside teststop* when generating/executing test scenarios.

| Context | Path | Permission |
|---------|------|-----------|
| v0.1 (generate) | no filesystem mount | AI receives mandate as CLI arg, outputs JSON to stdout |
| v0.2 (execute) | user's project | 🔒 read-only — AI reads code but cannot modify it |
| v0.2 (execute) | localhost network | ✅ allowed — AI calls the running app under test |
| Host filesystem | everything else | ❌ invisible |

**Fallback — no container** (`TESTSTOP_SANDBOX=none` or container not installed):
Direct `exec.Command("claude", "-p", mandate)` — inherits full host user permissions.
This is the only case where the AI touches the real host. Always prefer sandbox when available.

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

<!-- BEGIN DOCUFLOW -->
# DocuFlow — AI Documentation Assistant

DocuFlow preserves decision context for AI agents. Intent in, value out.

## Core tools (use these first)

- **query_wiki({ project_path, question })** — Ask the wiki. Returns an answer with citations.
- **ingest_source({ project_path, source_filename })** — Fold a markdown source into the wiki.
- **wiki_search({ project_path, query })** — BM25 search across all pages.
- **read_module({ path })** — Read and extract facts from a single source file.

## CLI — Core Commands

```
docuflow query "<question>"         # ask the wiki from the shell
docuflow ingest <source.md>         # add a source doc to the wiki
docuflow status                     # wiki health and counts
docuflow rewiki                     # re-ingest with current rules
docuflow init                       # initialise .docuflow/ in this project
```

## Workflows

### Answer a question
```
query_wiki({ project_path: "/Volumes/SATECHI_WD_BLACK_2/dev/teststop", question: "How does authentication work?" })
```

### Add new context
```
# drop a markdown file in .docuflow/sources/
ingest_source({ project_path: "/Volumes/SATECHI_WD_BLACK_2/dev/teststop", source_filename: "auth-design.md" })
```

## Advanced tools

Use when the core tools don't cover the workflow. Each has more parameters and side effects.

- **list_modules** — Walk a directory tree and extract facts in bulk
- **list_wiki** — Inventory pages by category, with staleness flags
- **write_spec / read_specs** — Persistent agent-written specs
- **save_answer_as_page** — Promote a synthesised answer into the wiki
- **synthesize_answer** — Combine multiple pages into a markdown synthesis
- **update_index** — Rebuild `.docuflow/index.md`
- **lint_wiki** — Health checks: orphans, broken refs, stale content
- **get_schema_guidance** — Recommend what pages should exist
- **preview_generation** — Show what a tool will do before running
- **generate_dependency_graph** — Build the import/shared-table graph

## Storage Layout

```
.docuflow/
├── specs/           Spec files written by write_spec
├── wiki/            LLM-generated wiki pages
│   ├── entities/    Named things (services, APIs, databases)
│   ├── concepts/    Design patterns, principles, integrations
│   ├── timelines/   Chronological pages
│   └── syntheses/   Cross-cutting synthesis pages
├── sources/         Raw input files for ingest_source
├── schema.md        Wiki configuration (edit to customise)
├── index.md         Auto-maintained catalog
└── log.md           Operation log
```
<!-- END DOCUFLOW -->