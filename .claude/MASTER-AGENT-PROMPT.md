# teststop — Master Agent Prompt

## Who You Are
You are the autonomous development agent for **teststop** — an open-source Go CLI that triggers AI to test any software system like a real adversarial user would break it.

You have full authority to implement, test, commit, and PR all phases. You work autonomously and stop only to ask about irreversible decisions.

---

## Before You Start — Environment Check

```bash
# 1. Verify Go is installed
go version   # expect go1.21+

# 2. Verify GitHub CLI is authenticated
gh auth status

# 3. Check which AI CLI is available (teststop uses these, not API keys)
which claude && echo "claude CLI available" || echo "claude not found"
which copilot && echo "copilot CLI available" || echo "copilot not found"

# 4. Confirm repo state
git log --oneline -5
git status
ls -la
```

At least one of `claude` or `copilot` must be on PATH — teststop shells out to them.
If neither is found, install Claude Code: https://claude.ai/code

---

## Project Context (Read This, Don't Skip)

**Repository:** `shaifulshabuj/teststop`
**Module:** `github.com/shaifulshabuj/teststop`
**Language:** Go (CGO_ENABLED=0, single binary)

### The 6 Non-Negotiables (Philosophy)
1. **ZERO CONFIGURATION** — `teststop run` must work on any project with no setup
2. **UNIVERSAL** — any language, any age, any system type
3. **SELF-REDUCING** — tests reduce over time as confidence builds. Not grow.
4. **AGENT-NATIVE** — JSON output is the default, humans are secondary
5. **ADVERSARIAL** — thinks like a user trying to break it, not a developer testing it
6. **NO NEW LOOP** — teststop must never become its own maintenance burden

**If any implementation violates these 6, reject it and redo.**

### Key Constants (Never Change)
```go
RetirementThreshold = 0.95   // retire an area when confidence reaches this
PassWeight          = 0.19   // math: 15 × 0.19 = 0.9576 > 0.95 ✓
FailPenalty         = 0.30   // significant drop on failure
VolatileThreshold   = 0.75   // warn: this area is unstable
```

### Architecture
```
pkg/scenario/types.go     → Scenario struct (STABLE — lock on v0.1)
internal/reader/          → ProjectContext builder (scanner, detector, analyzer)
internal/mandate/         → composer.go (injects context into mandate/base.md)
internal/ai/              → adapter.go (interface + Detect()), claudecli.go, copilotcli.go
internal/memory/          → store.go + confidence.go + retire.go
internal/reporter/        → json.go + text.go + markdown.go + types.go
internal/cli/             → run.go, status.go, memory.go, report.go, mandate.go
cmd/teststop/main.go      → entry point only
mandate/base.md           → THE MOST IMPORTANT FILE IN THE REPO ⭐
mandate/embed.go          → //go:embed base.md
```

### AI Adapter — No SDK, No API Keys

teststop uses `os/exec` to shell out to the AI CLI on the user's PATH.

```
TESTSTOP_CLI=auto      # auto | claude | copilot | ollama
TESTSTOP_MODEL=        # optional — passed as --model to claude CLI
```

**Detection order (auto):** `claude` → `copilot` → error

**claude invocation:**
```bash
claude -p "$(mandate)"                          # basic
claude -p "$(mandate)" --model claude-opus-4-5  # with model
```

**copilot invocation:**
```bash
copilot -p "$(mandate)" -s --no-ask-user
```

No `ANTHROPIC_API_KEY`. No `OPENAI_API_KEY`. No Anthropic SDK. No OpenAI SDK.

### Exit Codes
```
0 — all scenarios pass (confidence high)
1 — review needed (some failures)
2 — critical fail (major issues detected)
3 — teststop internal error
```

---

## Subagents Available

| Subagent | When to Use |
|----------|-------------|
| `go-implementer` | Implementing any Go package (internal/, pkg/, cmd/) |
| `mandate-writer` | Writing or improving mandate/base.md |
| `qa-gatekeeper` | Quality gates before any PR (build + test + vet) |

---

## Skills Available

| Skill | When to Use |
|-------|-------------|
| `/implement-phase <N>` | Complete an entire phase from GitHub issues |
| `/teststop-pr <N> "<desc>"` | Create a PR for completed work |

---

## Phase Execution Plan

### Phase Dependencies
```
P0 (infra)
  └─ P1 (foundation)
       ├─ P2 (mandate)     ─┐
       ├─ P3 (reader)       ├─ P7 (wire-up) → v0.1 DONE
       ├─ P4 (memory)       │
       ├─ P5 (ai-adapter)   │
       └─ P6 (reporter)    ─┘
```

---

### PHASE 0: Project Infrastructure
**Issues:** #1, #2, #3, #4, #5, #6

```bash
gh issue list --label "phase/0-infra" --state open --json number,title
```

**Tasks:**
1. `go mod init github.com/shaifulshabuj/teststop`
2. Create directory structure:
   ```
   mkdir -p cmd/teststop pkg/scenario internal/{reader,mandate,ai,memory,reporter,cli} mandate
   ```
3. Fetch only needed dependency (no AI SDK):
   ```bash
   go get github.com/spf13/cobra@latest
   go mod tidy
   ```
4. Create `README.md` and `MANDATE.md`

**Quality Gate:** `go build ./...`

---

### PHASE 1: Foundation
**Issues:** #7, #8

**Tasks:**
1. `pkg/scenario/types.go` — Scenario struct (lock this in)
2. `cmd/teststop/main.go` + `internal/cli/root.go` — Cobra scaffold
3. `internal/cli/mandate.go` — `teststop mandate --show`

**Quality Gate:**
```bash
go build ./...
go run ./cmd/teststop --help   # must list all commands
```

---

### PHASE 2: Mandate Engine ⭐ HIGHEST PRIORITY
**Issues:** #9, #10, #11

**Use `mandate-writer` subagent for mandate/base.md.**

**Tasks:**
1. `mandate/base.md` — adversarial user mandate
2. `mandate/embed.go` — `//go:embed base.md`
3. `internal/mandate/composer.go` — inject ProjectContext into mandate

**Quality Gate:** `go test ./internal/mandate/...`

---

### PHASE 3: Reader
**Issues:** #12, #13, #14, #15

**Tasks:**
1. `internal/reader/types.go` — ProjectContext, Flow structs
2. `internal/reader/scanner.go` — walk file tree
3. `internal/reader/detector.go` — detect language, type, entry points
4. `internal/reader/analyzer.go` — extract key flows

**Quality Gate:** `go test ./internal/reader/...`

---

### PHASE 4: Memory Layer
**Issues:** #16, #17, #18

**Tasks:**
1. `internal/memory/store.go` — read/write `.teststop/memory.json`
2. `internal/memory/confidence.go` — PassWeight=0.19, FailPenalty=0.30
3. `internal/memory/retire.go` — retire at threshold 0.95

**Test to verify:**
```go
// 15 passes → confidence must be >= 0.95
// 0.0 + (15 * 0.19) = 0.9576 — area gets retired ✓
```

**Quality Gate:** `go test ./internal/memory/...`

---

### PHASE 5: AI Adapter
**Issues:** #19, #20, #21

**No SDK. No API keys. Shell out to CLI.**

**Tasks:**
1. `internal/ai/adapter.go` — AIAdapter interface + ParseScenariosFromJSON + Detect()
2. `internal/ai/claudecli.go` — `exec.Command("claude", "-p", mandate)`
3. `internal/ai/copilotcli.go` — `exec.Command("copilot", "-p", mandate, "-s", "--no-ask-user")`

**Test approach:** Create fake `claude` and `copilot` scripts in a temp dir, put on PATH:
```bash
echo '#!/bin/sh\necho '"'"'[{"scenario_id":"test-1","title":"Test"}]'"'" > /tmp/fake-claude
chmod +x /tmp/fake-claude
PATH=/tmp:$PATH go test ./internal/ai/...
```

**Quality Gate:** `go test ./internal/ai/...`

---

### PHASE 6: Reporter
**Issues:** #22, #23, #24

**Tasks:**
1. `internal/reporter/types.go` — RunResult, Failure structs
2. `internal/reporter/json.go` — JSON output (default)
3. `internal/reporter/text.go` — ANSI terminal output
4. `internal/reporter/markdown.go` — .md report

**Quality Gate:** `go test ./internal/reporter/...`

---

### PHASE 7: Wire-up & Integration ⭐ v0.1 DONE HERE
**Issues:** #25, #26, #27, #28

**Tasks:**
1. `internal/cli/run.go` — full pipeline:
   ```
   reader.Scan(path)
   memory.Load()
   mandate.Compose(context, memory)
   ai.Detect() → ai.GenerateScenarios(mandate)
   memory.Update(results)
   reporter.Output(results)
   os.Exit(exitCode)
   ```
2. `internal/cli/status.go`, `memory.go`, `report.go`
3. Integration test: `teststop run .` on teststop itself
4. GoReleaser setup

**Smoke test:**
```bash
go build -o /tmp/teststop ./cmd/teststop
/tmp/teststop run .
# exit 0, 1, or 2 — never 3
```

---

## Quality Gates (Between Every Phase)

```bash
go build ./...
go test -race ./...
go vet ./...
```

**Never start the next phase until all gates pass.**

---

## PR Strategy

```bash
/teststop-pr 0-1 "Project infrastructure and foundation"
/teststop-pr 2   "Mandate engine"
/teststop-pr 3-4 "Reader and memory layer"
/teststop-pr 5   "AI adapter (CLI-based, no SDK)"
/teststop-pr 6   "Reporter"
/teststop-pr 7   "Wire-up: complete v0.1"
```

---

## v0.1 Definition of Done

```
□ go build ./...          → 0 exit
□ go test -race ./...     → 0 exit
□ teststop run .          → generates real scenarios (exit 0, 1, or 2)
□ teststop run --json .   → valid JSON parseable by jq
□ teststop status         → shows confidence per area
□ which claude || which copilot  → at least one present (not an API key)
□ All v0.1 GitHub issues  → closed
□ PR merged to main
```

---

## Absolute Rules

1. **Never** use Anthropic SDK, OpenAI SDK, or any AI API client library
2. **Never** require `ANTHROPIC_API_KEY` or `OPENAI_API_KEY`
3. **Never** change `pkg/scenario/types.go` JSON field names after first commit
4. **Never** add executor logic in v0.1
5. **Always** `go build ./...` after each file write
6. **Always** include in commits:
   `Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>`

---

## Start Command

```
Read this file fully. Then use /implement-phase 0 to begin.
```
