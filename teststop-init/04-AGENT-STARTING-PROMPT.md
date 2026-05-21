# teststop — Initial Agent Prompt

> Copy this prompt and give it to Claude Code, GitHub Copilot, or any AI coding agent
> to begin building teststop from scratch.

---

## The Prompt

```
You are building "teststop" — an open-source, agent-native CLI tool.

Read the following carefully before writing any code.

---

## What teststop is

teststop is a CLI tool with one job:
Trigger AI to test any software system the way a real adversarial user would break it.

It is NOT a test runner.
It is NOT a test framework.
It is NOT a replacement for testers.
It is a TRIGGER — a thin CLI that gives AI the right mandate,
then gets out of the way.

---

## The Core Philosophy

Every test ever written was written by someone who knew how the system works.
Real users don't know. That's why production still surprises us.

We call it test coverage. What we actually have is assumption coverage.

teststop solves this by giving AI a mandate to think like a real adversarial user:
- Someone who never read the docs
- Someone who retries when something is slow
- Someone who opens the same form in two tabs
- Someone who pastes unexpected data
- Someone who abandons mid-flow and returns an hour later
- Someone who does what no spec ever imagined

AI already knows all of this. It just needed the right instruction.
teststop provides that instruction — called the Mandate.

---

## Design Principles (Non-negotiable)

1. ZERO CONFIGURATION — teststop run must work with no setup on any project
2. UNIVERSAL — must work on any language, any age, any system type
3. SELF-REDUCING — tests reduce over time as confidence builds. Not grow.
4. AGENT-NATIVE — output is machine-readable JSON for AI agent consumption
5. NO NEW LOOP — teststop must never become the thing that needs to be maintained
6. EXIT CONDITION — success = the user needs it less over time, not more

---

## Project Structure to Build

```
teststop/
├── cmd/
│   └── teststop/
│       └── main.go              # CLI entry point (Cobra)
├── internal/
│   ├── reader/
│   │   ├── scanner.go           # Scan project structure
│   │   ├── detector.go          # Detect language, type, entry points
│   │   └── analyzer.go          # Analyze data flows and surface area
│   ├── mandate/
│   │   ├── base.go              # The core adversarial user mandate (THE KEY FILE)
│   │   ├── composer.go          # Compose mandate with context
│   │   └── templates/
│   │       ├── base.md          # Base mandate text
│   │       └── context.md       # Context enrichment template
│   ├── ai/
│   │   ├── adapter.go           # AI adapter interface
│   │   ├── claude.go            # Anthropic Claude implementation
│   │   └── openai.go            # OpenAI fallback implementation
│   ├── memory/
│   │   ├── store.go             # Read/write .teststop/ memory files
│   │   ├── confidence.go        # Confidence scoring and updating
│   │   └── retire.go            # Test retirement logic
│   ├── executor/
│   │   └── runner.go            # Execute generated scenarios (v0.2)
│   └── reporter/
│       ├── json.go              # JSON output (agent-parseable)
│       ├── text.go              # Human-readable terminal output
│       └── markdown.go          # Markdown report generation
├── pkg/
│   └── scenario/
│       └── types.go             # Scenario data structures
├── .teststop/                   # Memory directory (created at runtime)
│   ├── memory.json
│   ├── retired.json
│   └── config.yaml
├── mandate/
│   └── base.md                  # THE MANDATE — core adversarial user instruction
├── go.mod
├── go.sum
├── README.md
└── MANDATE.md                   # Explains the mandate philosophy publicly
```

---

## Start Here: Build in This Order

### Step 1 — Project scaffold
- Initialize Go module: `go mod init github.com/shaifulshabuj/teststop`
- Set up Cobra CLI with these commands:
  - `teststop run` (main command)
  - `teststop status`
  - `teststop memory`
  - `teststop report`
  - `teststop mandate --show`

### Step 2 — The Mandate (most important file)
Build `mandate/base.md` — the core instruction that makes AI test like a real adversarial user.

This file is the intellectual heart of teststop. It must:
- Instruct AI to think as a real user who has never read documentation
- Cover real human behavior patterns (retry, abandon, concurrent, unexpected input)
- Cover chaos patterns (slow network, partial failure, state inconsistency)
- Request structured JSON scenario output
- Be language and system-type agnostic
- Be clear enough that a community member could improve it

Start with this base mandate structure:

```markdown
# teststop Mandate — Adversarial User Testing

You are testing [SYSTEM_NAME] as a real adversarial user.
You have never read the documentation. You do not know how it was built.
You only know what you want to accomplish.

## Your Testing Mindset

You are not a developer reviewing code.
You are a real human with real frustrations, making real mistakes.

You will:
- Try to accomplish your goal the most natural way, not the correct way
- Retry when things are slow or unclear
- Make mistakes and expect the system to handle them gracefully
- Use the system on a slow or unreliable connection
- Do multiple things at once
- Abandon tasks and come back later
- Paste data from other sources without cleaning it
- Do things in an order the developer never anticipated

## System Under Test

Project: [PROJECT_NAME]
Language: [DETECTED_LANGUAGE]
Type: [DETECTED_TYPE: web_app | api | cli | library | service]
Entry Points: [DETECTED_ENTRY_POINTS]
Key Flows: [DETECTED_FLOWS]

## Already Proven Stable (Do Not Re-test Aggressively)

[MEMORY_STABLE_AREAS]

## Focus Areas (New or Changed)

[MEMORY_VOLATILE_AREAS]

## Generate Test Scenarios

Generate [N] test scenarios as a real adversarial user would experience this system.

For each scenario, output valid JSON matching this schema:
{
  "scenario_id": "string",
  "title": "string",
  "user_perspective": "string — who is this user and what do they want",
  "preconditions": ["string"],
  "steps": ["string"],
  "chaos_factors": ["string — what makes this hard: slow network, bad input, etc"],
  "expected_behavior": "string — what should happen",
  "failure_modes": ["string — what could go wrong"],
  "priority": "critical | high | medium | low",
  "confidence_area": "string — which part of the system this covers",
  "is_edge_case": boolean
}

Output a JSON array of scenarios. Nothing else. No explanation. No preamble.
```

### Step 3 — Reader (Code Scanner)
Build `internal/reader/`:
- Detect language from file extensions and config files
- Detect project type (web app, API, CLI, library)
- Find entry points (main.go, app.py, index.js, routes/, etc.)
- Extract key flows (HTTP routes, exported functions, CLI commands)
- Output a `ProjectContext` struct for the mandate composer

```go
type ProjectContext struct {
    Name         string
    Language     string
    Type         string // web_app | api | cli | library | service
    EntryPoints  []string
    KeyFlows     []Flow
    Dependencies []string
    TestFiles    []string // existing tests, if any
    FileCount    int
    Complexity   string // simple | moderate | complex
}
```

### Step 4 — Memory Layer
Build `internal/memory/`:
- Read `.teststop/memory.json` on startup
- Write updated confidence after each run
- Identify stable areas (confidence >= 0.95) for reduced testing
- Identify volatile areas (recently changed or low confidence) for full testing
- Retire tests automatically when confidence threshold is met

```go
type Memory struct {
    SystemAreas       map[string]AreaConfidence `json:"system_areas"`
    OverallConfidence float64                   `json:"overall_confidence"`
    MaturityStage     string                    `json:"maturity_stage"` // new | growing | mature | legacy
    LastRun           time.Time                 `json:"last_run"`
    TotalRuns         int                       `json:"total_runs"`
}

type AreaConfidence struct {
    Confidence  float64   `json:"confidence"`
    LastTested  time.Time `json:"last_tested"`
    TestCount   int       `json:"test_count"`
    Status      string    `json:"status"` // stable | volatile | new
    Notes       string    `json:"notes"`
}
```

### Step 5 — AI Adapter
Build `internal/ai/`:
- Interface: `AIAdapter` with `GenerateScenarios(mandate string) ([]Scenario, error)`
- Claude implementation using Anthropic API
- OpenAI fallback
- Parse JSON response from AI into `[]Scenario`
- Handle API errors gracefully with clear messages

Environment variables:
```
ANTHROPIC_API_KEY=
OPENAI_API_KEY=      (fallback)
TESTSTOP_AI=claude   (claude | openai | local)
TESTSTOP_MODEL=claude-opus-4-5
```

### Step 6 — Mandate Composer
Build `internal/mandate/`:
- Load `mandate/base.md`
- Inject `ProjectContext` from Reader
- Inject memory state (stable areas, volatile areas)
- Inject scenario count based on project complexity and maturity
- Output composed mandate string ready for AI

### Step 7 — Reporter
Build `internal/reporter/`:

JSON output (default for agents):
```json
{
  "run_id": "string",
  "timestamp": "ISO8601",
  "project": "string",
  "overall_confidence": 0.82,
  "maturity_stage": "growing",
  "ready_for_deploy": true,
  "scenarios_generated": 34,
  "scenarios_passed": 31,
  "scenarios_failed": 3,
  "failures": [...],
  "retired_this_run": 2,
  "confidence_delta": 0.03
}
```

Exit codes:
- `0` = confidence threshold met
- `1` = below threshold, review needed
- `2` = critical failures, do not deploy
- `3` = teststop internal error

Text output (for humans):
```
teststop v0.1.0

Analyzing: my-app (Node.js API, 847 files)
Memory: 23 stable areas, 3 volatile areas

Generating scenarios...
  ✓ 34 scenarios generated

Running scenarios...
  ✓ 31 passed
  ✗  3 failed

Failures:
  [HIGH] concurrent-checkout-001 — Race condition on inventory lock
  [MED]  upload-partial-failure  — No recovery after 50% upload
  [MED]  session-timeout-retry   — Duplicate submission on retry

Confidence: 0.82 (+0.03 from last run)
Maturity:   Growing
Status:     Review required before deploy

Retired this run: 2 tests (auth-basic, auth-login — proven stable)

Run teststop report for full details.
```

### Step 8 — Main Command Wire-up
Wire everything together in `cmd/teststop/main.go`:

```
teststop run
  → reader.Scan(path)
  → memory.Load()
  → mandate.Compose(context, memory)
  → ai.GenerateScenarios(mandate)
  → memory.Update(results)
  → reporter.Output(results)
  → os.Exit(exitCode)
```

---

## Key Files to Get Right

### `mandate/base.md`
This is the most important file. It is the soul of teststop.
Get this wrong and everything else fails regardless of how good the code is.
Iterate on this constantly. It should improve with every release.

### `internal/memory/confidence.go`
This is what makes teststop self-reducing.
Get confidence scoring right: it must reward stability and penalize change.
The retirement threshold must be conservative (0.95, not 0.8).

### `pkg/scenario/types.go`
The scenario schema is the contract between teststop and AI agents.
Once defined in v0.1, changes to it are breaking changes.
Design it carefully.

---

## README Requirements

The README must explain:
1. What teststop is (one sentence)
2. The philosophy (assumption coverage vs reality coverage)
3. How to install
4. `teststop run` — that's it for basic usage
5. How the mandate works (link to MANDATE.md)
6. How memory works (tests reduce over time)
7. How to use with Claude Code / Copilot
8. The ecosystem (DocuFlow, Waymark, teststop)

---

## MANDATE.md (Public Philosophy File)

Create a `MANDATE.md` at the root that explains to any contributor:
- What the mandate is
- Why it is the core of teststop
- How to improve it
- The philosophy behind adversarial user testing
- How to contribute mandate improvements

---

## Do Not Build (MVP Scope)

- Do NOT build dynamic test execution (generating + running code) in v0.1
- Do NOT build Waymark integration yet
- Do NOT build DocuFlow integration yet
- Do NOT build a web UI
- Do NOT build CI/CD plugins
- Do NOT build Windows support yet

Keep it thin. Keep it working. Keep it honest.

---

## The One Question to Ask at Every Decision Point

> "Does this make the user's problem easier to solve — or our system easier to build?"

If the answer is the second one: it is accidental complexity. Cut it.

---

## Relevant Projects for Context

- **Waymark** (same author): MCP middleware for AI agent governance
  https://github.com/shaifulshabuj/waymark

- **DocuFlow** (same author): MCP server implementing the LLM Wiki pattern
  https://github.com/shaifulshabuj/docuflow-mcp

teststop completes the trilogy:
- DocuFlow gives AI context to act
- Waymark gives humans reason to trust
- teststop gives systems confidence to prove themselves

---

Start with Step 1. Build incrementally.
Ask before adding complexity.
The mandate is the priority. Everything else serves the mandate.
```

---

*This prompt is ready to paste directly into Claude Code, GitHub Copilot Chat,
or any AI coding agent's initial session.*

*Project: teststop*
*Author: Shaiful Shabuj*
