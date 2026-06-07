# Changelog

All notable changes to teststop are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.2.1] — 2026-06-07

### Added

- `teststop version` command and `--version` / `-v` flags, reporting the
  GoReleaser-injected version, commit, build date, Go version, and os/arch. For
  `go install` builds the version is recovered from the module build info.
- `--help` now organizes commands into **Core** and **Meta** groups and includes
  a usage examples section.

### Changed

- GoReleaser now injects `main.commit` and `main.date` in addition to
  `main.version`.

---

## [v0.2.0] — 2026-06-06

teststop becomes a scenario **runner**, not just a scenario **generator**.

### Added

**Dynamic Scenario Execution (`internal/executor/`)**

- `teststop run --target <url>` — execute generated scenarios against a running
  system and feed real pass/fail outcomes into confidence memory
- **Hybrid execution**, chosen per scenario:
  - `HTTPExecutor` — deterministic `net/http` execution for scenarios carrying a
    structured `exec` block (retries on transport errors and 5xx, per-request
    timeout, status-code judging)
  - `AIExecutor` — AI-driven execution for prose-only scenarios when `--target`
    is set; the AI performs the steps and returns a structured verdict
  - `StaticExecutor` — structural validation only (the no-`--target` default,
    preserving v0.1 behavior)
- Bounded, order-stable concurrent execution with context cancellation
- New `run` flags: `--target`, `--concurrency` (4), `--exec-timeout` (10s),
  `--max-retries` (2)

**Scenario Schema (additive, non-breaking)**

- Optional `exec` field on `Scenario` (`mode`, `method`, `path`, `headers`,
  `body`, `expected_status`, `command`, `expected_exit`). Legacy v0.1 scenario
  JSON without `exec` continues to parse unchanged.

**AI Adapter**

- `Prompt(input)` added to the `AIAdapter` interface for AI-driven execution;
  `GenerateScenarios` now builds on it. Both Claude and Copilot adapters updated.

**Reporting**

- `RunResult` gains `executions` and an `exec_summary` (executed/passed/failed +
  target); text and Markdown reports render an execution summary. Failures are
  now derived from real execution outcomes.

**Mandate**

- `mandate/base.md` now invites the AI to emit an optional `exec` block when a
  scenario maps cleanly to a single concrete HTTP request.

### Changed

- The `run` pipeline executes scenarios and updates confidence from **real**
  outcomes instead of granting every area an automatic pass. A failed
  `critical` scenario now yields exit code `2`.

### Notes

- A sandboxed (Apple Container) AI tester cannot reach the host's `localhost`;
  use `TESTSTOP_SANDBOX=none` for local targets, or target a reachable
  staging/production-like URL. Sandbox-network-aware execution (wiring the
  reserved `Config.Runner`) is tracked as future work.

---

## [v0.1.0] — 2025-05-21

First public release of teststop.

### Added

- `teststop run` — full adversarial testing pipeline: scan → mandate → generate
  → memory → report
- Static project scanner, mandate composer, confidence memory system with
  area retirement, and JSON / ANSI text / Markdown reporters
- Claude and GitHub Copilot CLI adapters with `TESTSTOP_CLI` auto-detection
- Apple Container sandbox isolation (`auto` / `required` / `none`) with
  read-only credential mounts and direct-execution fallback
- CLI commands: `run`, `status`, `memory`, `report`, `mandate --show`
- `mandate/base.md` — adversarial user mandate, embedded via `//go:embed`
- Stable scenario JSON schema (`pkg/scenario/types.go`)
- GoReleaser distribution (darwin/linux × amd64/arm64) and CI/Release workflows
- Exit codes: `0` ok, `1` review, `2` critical, `3` internal error

---

[v0.2.1]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.2.1
[v0.2.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.2.0
[v0.1.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.1.0
