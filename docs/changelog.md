# Changelog

All notable changes to teststop are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.3.1] — 2026-06-08

Correctness fixes from a Waymark usage review (#44).

### Fixed

- **AI infrastructure errors no longer count as scenario failures** (#44). An AI
  CLI error (e.g. exit 1 on rate-limit exhaustion) or unparseable verdict now marks
  the scenario **skipped**, not failed — excluded from confidence, failures, and the
  exit code, and reported separately (`exec_summary.skipped`). A rate-limited run no
  longer fabricates failures that drag confidence down.
- **Spawned AI runs in a neutral working directory** (#44). Direct `claude`/`copilot`
  calls run from the system temp dir, so they no longer load the target project's
  `CLAUDE.md` / MCP configuration.

---

## [v0.3.0] — 2026-06-07

### Added

- **Concurrency exec mode** (#43) — `exec.concurrency`: when `> 1`, the HTTP
  executor fires N identical requests at once and asserts exactly one wins (the
  rest cleanly rejected), deterministically verifying race guards like
  double-submit and claim-the-last-item. The mandate invites the AI to emit
  `concurrency` for race scenarios.

### Changed

- **Reporter honesty** (#42) — runs without `--target` are clearly labelled
  **predicted** (a risk surface), not executed. Reports show "PREDICTED RISKS" /
  "PREDICTED CONFIDENCE" with a caveat to run `--target` to verify; executed runs
  keep the verified ✓/✗ + CONFIDENCE framing.
- `exec_summary` JSON now carries `executed` (bool) and `count` (int).

---

## [v0.2.1] — 2026-06-07

### Added

- `teststop version` command and `--version` / `-v` flags, reporting the
  GoReleaser-injected version, commit, build date, Go version, and os/arch. For
  `go install` builds the version is recovered from the module build info.
- `--help` now organizes commands into **Core** and **Meta** groups and includes
  a usage examples section.

### Changed

- GoReleaser now injects `main.commit` and `main.date` alongside `main.version`.

---

## [v0.2.0] — 2026-06-06

teststop becomes a scenario **runner**, not just a scenario **generator**.

### Added

**Dynamic Scenario Execution (`internal/executor/`)**

- `teststop run --target <url>` — execute generated scenarios against a running
  system and feed real pass/fail outcomes into confidence memory
- **Hybrid execution**, chosen per scenario:
    - **HTTP** — deterministic `net/http` execution for scenarios carrying a
      structured `exec` block (retries on transport errors and `5xx`,
      per-request timeout, status-code judging)
    - **AI-driven** — for prose-only scenarios when `--target` is set; the AI
      performs the steps and returns a structured verdict
    - **Static** — structural validation only (the no-`--target` default,
      preserving v0.1 behavior)
- Bounded, order-stable concurrent execution with context cancellation
- New `run` flags: `--target`, `--concurrency` (4), `--exec-timeout` (10s),
  `--max-retries` (2)

**Scenario Schema (additive, non-breaking)**

- Optional `exec` field on the scenario object (`mode`, `method`, `path`,
  `headers`, `body`, `expected_status`, `command`, `expected_exit`). Legacy v0.1
  JSON without `exec` parses unchanged.

**AI Adapter**

- `Prompt(input)` added to the adapter interface for AI-driven execution;
  `GenerateScenarios` builds on it. Claude and Copilot adapters updated.

**Reporting**

- `RunResult` gains `executions` and `exec_summary`; text and Markdown reports
  render an execution summary. Failures now derive from real execution outcomes.
  `ExecutionResult.duration_ms` is emitted in true milliseconds.

**Mandate**

- `mandate/base.md` invites an optional `exec` block when a scenario maps cleanly
  to a single concrete HTTP request.

### Changed

- The `run` pipeline updates confidence from **real** execution outcomes instead
  of granting every area an automatic pass. A failed `critical` scenario now
  yields exit code `2`.

### Notes

- A sandboxed (Apple Container) AI tester cannot reach the host's `localhost`;
  use `TESTSTOP_SANDBOX=none` for local targets, or target a reachable
  staging/production-like URL.

---

## [v0.1.0] — 2025-05-21

First public release of teststop.

### Added

**Core Pipeline**

- `teststop run` — full adversarial testing pipeline: scan → mandate → generate → memory → report
- Static project scanner (`internal/reader/`) — detects language, system type, routes, flows, and dependencies across Go, Python, TypeScript, Ruby, Rust, and more
- Mandate composer (`internal/mandate/`) — injects project context and memory into `mandate/base.md`
- Confidence memory system (`internal/memory/`) — per-area scoring with exponential approach formula
- Area retirement at ≥ 0.95 confidence AND ≥ 15 test count
- Reporter (`internal/reporter/`) — JSON, ANSI text, and Markdown output formats

**AI Adapters**

- Claude CLI adapter (`internal/ai/claudecli.go`) — calls `claude -p "<mandate>"` with optional `--model`
- GitHub Copilot CLI adapter (`internal/ai/copilotcli.go`) — calls `copilot -p "<mandate>" -s --no-ask-user`
- Auto-detection via `TESTSTOP_CLI` environment variable

**Sandbox Isolation**

- Apple Container integration (`internal/sandbox/`) — runs AI CLI in isolated VM
- Three modes: `auto`, `required`, `none` via `TESTSTOP_SANDBOX`
- Read-only credential mounts (`~/.claude`, `~/.config/gh`)
- Runtime image: `ghcr.io/shaifulshabuj/teststop-agent:latest` (Ubuntu 24.04 minimal)
- Automatic fallback to direct execution when container not available

**CLI Commands**

- `teststop run` — main test command with `--depth`, `--output`, `--threshold`, `--no-color`, `--quiet`
- `teststop status` — confidence state table
- `teststop memory` — show and reset memory
- `teststop report` — last run report
- `teststop mandate --show` — display composed mandate

**The Mandate**

- `mandate/base.md` — adversarial user mandate with 10 behavior patterns, 11 chaos conditions, 6 system type adaptations
- Embedded in binary via `//go:embed base.md`

**Scenario Schema**

- `pkg/scenario/types.go` — stable JSON contract for AI-generated scenarios
- Fields: `scenario_id`, `title`, `user_perspective`, `preconditions`, `steps`, `chaos_factors`, `expected_behavior`, `failure_modes`, `priority`, `confidence_area`, `is_edge_case`

**Distribution**

- GoReleaser configuration — 4 targets: `darwin/arm64`, `darwin/amd64`, `linux/arm64`, `linux/amd64`
- GitHub Actions CI workflow — build, test, vet on push and PR
- GitHub Actions Release workflow — test + GoReleaser on version tags

**Exit Codes**

- `0` — confidence threshold met
- `1` — below threshold (review required)
- `2` — critical failures found
- `3` — teststop internal error

---

## Roadmap

### v0.2 (in progress)

- ✅ **Scenario executor** — run generated scenarios against a live system _(shipped in v0.2.0)_
- **Ollama adapter** — local model support via `TESTSTOP_CLI=ollama`
- **`teststop watch`** — file-watching mode that re-runs on code changes
- **Sandbox-network-aware execution** — run the executor inside the sandbox network

### v1.0 (planned)

- **Waymark integration** — governance hooks for AI agent workflows
- **DocuFlow integration** — feed project documentation into mandate context
- **CI/CD plugins** — native GitHub Actions, GitLab CI support
- **`teststop diff`** — scenario comparison between runs

---

[v0.3.1]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.3.1
[v0.3.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.3.0
[v0.2.1]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.2.1
[v0.2.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.2.0
[v0.1.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.1.0
