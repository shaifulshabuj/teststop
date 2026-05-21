# teststop — Project Goals

---

## One-Line Goal

> *Make any AI agent test any software system the way a real user would break it — with zero configuration, for any codebase of any age.*

---

## Primary Goals

### G1 — Reality-Based Testing
Replace assumption-based testing with reality-based testing.

AI agents must test from the perspective of a real human user — not from the perspective of the developer who built the system. The agent must think like someone who has never read the docs, is impatient, makes mistakes, and does unexpected things.

**Success metric:** Tests generated reflect actual user failure patterns, not developer assumptions.

---

### G2 — Zero Configuration
teststop must work on any project with zero setup.

No config files. No production logs required. No environment setup. No framework-specific plugins. No onboarding time.

```bash
# This is the entire setup
teststop run
```

**Success metric:** A developer can run teststop on a project they have never seen before and get meaningful results within 60 seconds.

---

### G3 — Universal Compatibility
teststop must work on any system regardless of:
- Language (Go, Python, Rust, C#, Java, COBOL, anything)
- Age (2 days old or 50 years old)
- Type (web app, CLI tool, API, mobile backend, desktop app)
- Size (side project or enterprise system)
- State (new, growing, mature, legacy, unmaintained)

**Success metric:** teststop produces meaningful output on a COBOL system, a Next.js app, and a Rust CLI tool — without any system-specific configuration.

---

### G4 — Self-Reducing Test Surface
Tests must reduce over time, not grow.

teststop builds memory of what has been proven stable. It retires tests that no longer matter. It focuses progressively on new, changed, or fragile areas. As the system matures, the testing effort shrinks.

**Success metric:** On a 2-year-old project, teststop produces fewer, higher-confidence tests than on a new project of the same size.

---

### G5 — Agent-Native Design
teststop is built to be invoked by AI agents — not just humans.

Claude Code, GitHub Copilot, Codex, OpenCode, and future agents must be able to invoke teststop autonomously as part of their development workflow. Output must be machine-readable (JSON/structured) by default, with human-readable output as an option.

**Success metric:** An AI coding agent can invoke teststop, parse results, and make a deploy/no-deploy decision without human input.

---

### G6 — Designed to Disappear
teststop's long-term success is measured by how little of it remains.

A mature system running teststop should need it less and less over time. The tool accumulates confidence, not complexity. The goal state is a system that validates itself with near-zero testing overhead.

**Success metric:** On a stable 5-year-old project, teststop's testing surface is less than 10% of what it was at project start.

---

## Secondary Goals

### G7 — Open Source and Community-Driven
teststop is open source from day one. The mandate (the core prompt/instruction that drives AI testing) is the primary intellectual contribution and must be transparent, auditable, and improvable by the community.

### G8 — Waymark Integration (Governance)
teststop integrates with Waymark for agent governance. When AI agents invoke teststop as part of an automated pipeline, Waymark ensures the testing action is auditable and within defined governance policies.

### G9 — DocuFlow Integration (Context)
When DocuFlow is present in a project, teststop uses the captured WHY — the living theory of the system — to generate better, more contextually accurate test scenarios.

### G10 — No New Loop
teststop must not introduce a new maintenance loop. It must not become the thing that needs to be maintained, configured, and cleaned up. If teststop ever requires a "teststop cleanup sprint," the design has failed.

---

## Non-Goals (What teststop Will Never Be)

| Not This | Why |
|----------|-----|
| A test runner | Test runners already exist. teststop triggers AI to use them. |
| A test framework | Frameworks already exist. teststop is language-agnostic. |
| A CI/CD platform | CI/CD already exists. teststop plugs into it. |
| An observability tool | teststop does not require production data to function. |
| A replacement for testers | Testers define confidence. teststop executes it. |
| A complex system | If it needs its own admin, it has failed. |

---

## Definition of Done (For v1.0)

- [ ] `teststop run` works on any project with no configuration
- [ ] AI agent reads codebase and generates reality-based test scenarios
- [ ] Output is structured JSON parseable by other AI agents
- [ ] Memory layer (`.teststop/`) persists confidence across runs
- [ ] Tests reduce over repeated runs on stable code
- [ ] Works on at minimum: Python, Node.js, Go, Rust projects
- [ ] Waymark policy hook is available (even if optional)
- [ ] README explains the mandate — the core philosophy — clearly

---

*Document version: 1.0*
*Project: teststop*
*Author: Shaiful Shabuj*
