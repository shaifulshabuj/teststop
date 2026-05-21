# teststop — Product Requirements Document (PRD)

**Version:** 0.1 (Initial)
**Status:** Pre-build — Idea to Implementation
**Author:** Shaiful Shabuj

---

## Overview

teststop is an agent-native CLI tool that triggers AI to test any software system from the perspective of a real adversarial user. It requires zero configuration, works on any codebase of any age, and is designed to reduce its own footprint over time as the system matures.

---

## Problem Statement

### The Status Quo

Modern software testing is broken in a specific, consistent way:

1. Tests are written by people who understand the system
2. Real users do not understand the system — and break it in ways testers never imagined
3. The test suite grows indefinitely, becoming a maintenance burden of its own
4. Coverage metrics go up. Production surprises continue.

This is the **Assumption Coverage Problem**: we test what we think will happen, not what actually happens.

### The Gap teststop Fills

AI already knows:
- How humans actually behave with software
- What failure patterns exist across system types
- What edge cases emerge in production
- How to think adversarially about any system

AI just needed the right mandate to apply this knowledge. teststop provides that mandate.

---

## User Stories

### Primary Users

**The AI Coding Agent (Primary)**
> As an AI coding agent (Claude Code, Copilot, Codex, OpenCode),
> I want to invoke teststop after generating or modifying code,
> so that I can validate my changes against real-world usage patterns
> without requiring human intervention.

**The Developer (Secondary)**
> As a developer,
> I want to run `teststop run` on any project,
> so that I get a realistic assessment of how real users would break my system
> without having to write test scenarios myself.

**The Tester / QA Engineer (Strategic)**
> As a tester,
> I want teststop to handle the mechanical generation and execution of reality-based tests,
> so that I can focus on defining what confidence means for this system
> rather than writing test cases manually.

---

## Functional Requirements

### FR1 — Core CLI Interface

```bash
# Minimum viable command set

teststop run                    # Run full testing cycle on current directory
teststop run --path ./src       # Run on specific path
teststop run --depth aggressive # Control testing depth (light | normal | aggressive)
teststop run --output json      # Output format (json | text | markdown)
teststop run --threshold 85     # Confidence threshold (0-100)

teststop status                 # Show current confidence state of the project
teststop memory                 # Show what teststop has learned about this project
teststop memory --reset         # Clear accumulated memory (start fresh)

teststop report                 # Generate human-readable report of last run
teststop report --format md     # Output as markdown
```

### FR2 — The Mandate Engine

The mandate is the core of teststop. It is the instruction set given to the AI that makes it test like a real adversarial user rather than a developer.

The mandate must:
- Instruct AI to approach the system as a user who has never read documentation
- Instruct AI to generate scenarios based on real human behavior patterns
- Instruct AI to think adversarially — how would a real user break this?
- Be language-agnostic and system-type-agnostic
- Produce structured, executable test scenarios as output
- Be transparent, readable, and improvable by the community

The mandate is **not** hardcoded. It is a composable, versioned instruction set that improves over time.

### FR3 — Code Reading and Understanding

teststop must be able to read and understand any codebase:

- Scan project structure and identify system type automatically
- Identify entry points (API routes, CLI commands, UI flows, functions)
- Identify data flows (inputs, outputs, transformations)
- Identify dependencies and integration points
- Identify existing tests (if any) and assess their nature
- Work without running the code (static analysis is sufficient for scenario generation)

### FR4 — Scenario Generation

From code understanding, teststop generates test scenarios that reflect real user behavior:

Each scenario must include:
```json
{
  "scenario_id": "unique-id",
  "title": "Human-readable title",
  "user_perspective": "Who is this user and what do they want?",
  "steps": ["Step 1", "Step 2", "Step 3"],
  "chaos_factors": ["slow network", "invalid input", "concurrent action"],
  "expected_behavior": "What should happen",
  "failure_modes": ["What could go wrong"],
  "priority": "critical | high | medium | low",
  "confidence_area": "What part of the system this covers"
}
```

Scenarios must cover:
- Happy path (baseline)
- Abandonment and retry patterns
- Concurrent user actions
- Edge case inputs (special characters, empty values, extreme sizes)
- Network and dependency failure conditions
- State inconsistency scenarios
- Cross-session and cross-device scenarios (where applicable)

### FR5 — Memory Layer

teststop maintains a memory of what has been proven over time.

Storage location: `.teststop/` directory in project root (committed to version control)

```
.teststop/
├── memory.json          # Confidence state per system area
├── retired.json         # Tests retired and why
├── runs/                # History of test runs
│   └── 2025-05-19.json
└── config.yaml          # Optional project-level configuration
```

Memory schema:
```json
{
  "system_areas": {
    "auth_flow": {
      "confidence": 0.94,
      "last_tested": "2025-05-19",
      "test_count": 847,
      "status": "stable",
      "notes": "Proven stable. JWT expiry edge case resolved 2024-11."
    },
    "checkout_flow": {
      "confidence": 0.61,
      "last_tested": "2025-05-19",
      "test_count": 23,
      "status": "volatile",
      "notes": "New payment provider added 2025-04. Still accumulating confidence."
    }
  },
  "overall_confidence": 0.78,
  "maturity_stage": "growing",
  "last_run": "2025-05-19T10:30:00Z"
}
```

Memory rules:
- Confidence increases when an area passes tests repeatedly
- Confidence decreases when new changes touch an area
- Tests are automatically retired when confidence exceeds threshold (default: 0.95)
- Memory persists across runs and is committed to version control
- Memory is human-readable (JSON/YAML, not binary)

### FR6 — Structured Output (Agent-Parseable)

All teststop output must be parseable by AI agents:

```json
{
  "run_id": "2025-05-19-001",
  "timestamp": "2025-05-19T10:30:00Z",
  "project": "my-app",
  "overall_confidence": 0.82,
  "maturity_stage": "growing",
  "ready_for_deploy": true,
  "scenarios_generated": 34,
  "scenarios_executed": 34,
  "scenarios_passed": 31,
  "scenarios_failed": 3,
  "failures": [
    {
      "scenario_id": "concurrent-checkout-001",
      "title": "Concurrent checkout from 2 sessions",
      "failure": "Race condition on inventory lock",
      "severity": "high",
      "recommendation": "Add optimistic locking to inventory update"
    }
  ],
  "retired_this_run": 2,
  "new_scenarios_added": 1,
  "confidence_delta": +0.03
}
```

### FR7 — Agent Integration Interface

teststop must expose an interface that AI coding agents can invoke:

```bash
# Agent-friendly invocation
teststop run --output json --no-color --quiet

# Exit codes (agent decision-making)
# 0 = confidence threshold met, safe to proceed
# 1 = confidence below threshold, review required
# 2 = critical failures found, do not deploy
# 3 = teststop error (not a test failure)
```

### FR8 — Waymark Integration (Optional)

When Waymark is present, teststop hooks into governance:

```yaml
# .teststop/config.yaml
governance:
  waymark: true
  policy: ./governance/testing-policy.yaml
  strict: false
```

Waymark validates:
- That test quality meets policy standards
- That agent-invoked testing follows defined governance rules
- That test retirement decisions are auditable

### FR9 — DocuFlow Integration (Optional)

When DocuFlow is present, teststop reads the system's WHY:

```yaml
# .teststop/config.yaml
context:
  docuflow: true
  wiki_path: ./.docuflow/
```

DocuFlow context improves:
- Scenario relevance (tests match actual system purpose)
- Edge case identification (business rules inform failure modes)
- Confidence calibration (critical business flows get higher testing depth)

---

## Non-Functional Requirements

### NFR1 — Performance
- `teststop run` must complete within 5 minutes for a medium-sized project
- Memory reads/writes must not block the main execution path
- AI API calls must be parallelized where possible

### NFR2 — Reliability
- teststop must not fail silently — all errors must be reported clearly
- A failed teststop run must not affect the project being tested
- teststop must be idempotent — running it twice produces consistent results

### NFR3 — Portability
- Must run on macOS, Linux, Windows
- Must not require Docker or any containerization
- Single binary distribution preferred
- No runtime dependencies beyond the binary itself

### NFR4 — Privacy
- teststop does not send source code to external services by default
- When using cloud AI (Claude API), code is sent only to generate test scenarios
- A local-model mode must be supported (Ollama, LM Studio, etc.)
- No telemetry without explicit opt-in

### NFR5 — Transparency
- The mandate (core AI instruction) must be readable by the user
- `teststop mandate --show` prints the exact instructions given to AI
- Memory state is always human-readable (never opaque binary)

---

## Technical Architecture

### High-Level Components

```
teststop CLI
│
├── Reader          — Scans and understands the codebase
│   ├── Structure   — Project type, entry points, dependencies
│   ├── Flows       — Data flows, user flows, API surface
│   └── Existing    — Any existing tests (context only)
│
├── Mandate Engine  — Composes the instruction for the AI
│   ├── Base        — Core adversarial user mandate
│   ├── Context     — System-specific additions (from Reader)
│   └── Memory      — What's already been proven (from Memory)
│
├── AI Adapter      — Sends mandate + context to AI, receives scenarios
│   ├── Claude      — Anthropic Claude API
│   ├── OpenAI      — GPT-4/Codex
│   └── Local       — Ollama / LM Studio
│
├── Executor        — Runs generated scenarios
│   ├── Static      — Scenario validation without running code
│   └── Dynamic     — Scenario execution (requires runnable project)
│
├── Memory          — Persists and evolves confidence state
│   ├── Read        — Load prior confidence from .teststop/
│   ├── Update      — Adjust confidence based on run results
│   └── Retire      — Remove tests that have proven stable
│
├── Reporter        — Outputs results
│   ├── JSON        — Agent-parseable structured output
│   ├── Text        — Human-readable terminal output
│   └── Markdown    — Report file generation
│
└── Governance      — Optional hooks
    ├── Waymark     — Agent action auditing
    └── Policy      — Test quality validation
```

### Technology Stack (Recommended)

| Component | Choice | Reason |
|-----------|--------|--------|
| Language | Go | Single binary, fast, cross-platform, great CLI ecosystem |
| CLI Framework | Cobra | Standard Go CLI framework |
| AI Client | Anthropic SDK + OpenAI SDK | Claude primary, OpenAI fallback |
| Memory Storage | JSON files in .teststop/ | Human-readable, version-controllable |
| Config | YAML | Simple, readable |
| Output | JSON + ANSI text | Agent-parseable + human-readable |

---

## MVP Scope (v0.1)

### In Scope
- `teststop run` on a local directory
- Code reading (static, no execution required)
- Mandate engine (base mandate, no context enrichment yet)
- Claude API integration
- Scenario generation (JSON output)
- Basic memory layer (read/write confidence)
- Text and JSON output modes
- Works on: Python, Node.js, Go projects

### Out of Scope for MVP
- Dynamic execution (running the actual tests)
- Waymark integration
- DocuFlow integration
- Local model support
- Windows support (macOS + Linux first)
- CI/CD plugins
- Web UI

---

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| AI generates irrelevant scenarios | Medium | High | Mandate refinement is the core work — iterate fast |
| AI API costs at scale | Medium | Medium | Local model support in v0.2 |
| Memory grows too large | Low | Low | Auto-archive runs older than 90 days |
| Mandate becomes too opinionated | Medium | Medium | Keep mandate composable and community-editable |
| Complexity creep | High | High | Weekly check: "Does this make the loop harder to exit?" |

---

## Success Metrics

| Metric | Target | Timeframe |
|--------|--------|-----------|
| Time to first run on new project | < 60 seconds | v0.1 |
| Scenario relevance (human review) | > 70% rated useful | v0.1 |
| Confidence reduction on mature project | > 50% test surface reduction at 2 years | v1.0 |
| Agent adoption | Used by 3+ AI coding agents without modification | v0.2 |
| Community mandate contributions | 10+ community improvements to mandate | v1.0 |

---

*Document version: 0.1*
*Project: teststop*
*Author: Shaiful Shabuj*
