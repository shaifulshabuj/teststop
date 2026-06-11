# Changelog

All notable changes to teststop are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v1.0.0] — 2026-06-11

### Added

- **`.teststop/config.yaml` support.** The optional per-project config file is
  now real. Every key maps one-to-one onto an existing `teststop run` flag.
  Settings resolve with precedence **config file < `TESTSTOP_RUN_*` env var <
  explicit CLI flag**. A missing file is not an error; malformed YAML or an
  unknown key fails loudly. See `.teststop/config.example.yaml` for the keys.

- **`--ai-concurrency` flag** (default `1`). Caps the number of concurrently
  running AI-mode scenario executions to prevent rate-limit exhaustion. Config
  key `ai_concurrency`; env `TESTSTOP_RUN_AI_CONCURRENCY`.

- **Structured AI error detection.** The Claude adapter now calls
  `claude --output-format json` and parses the outer envelope, surfacing
  structured errors (rate-limit, auth failure, refusal) with context instead
  of raw stderr. Non-zero exits and envelope `is_error: true` both map to
  informative error messages.

- **E2E pipeline test** (`test/e2e/`). A full reader → mandate → adapter →
  memory → reporter → exit-code integration test using a fake-claude fixture
  script. Runs without real tokens; skippable via `-short`.

- **v1.0 contracts frozen** (`CONTRACTS.md`). Exit codes, scenario JSON schema,
  run output envelope, memory file format, and environment variables are
  declared stable. Breaking changes require a major version bump.

- **`examples/waymark-demo/`**. Replayable demo artifact: 52-scenario run
  against the waymark MCP middleware repo, with the captured JSON report,
  Markdown summary, the mandate used, and replay instructions.

### Changed

- `pkg/scenario/types.go` package comment declares the schema FROZEN at v1.0.
- `internal/cli` test coverage raised from 28.9% → 51.6%.

---

## [v0.3.1] — 2026-06-08

Correctness fixes from a Waymark usage review (#44).

### Fixed

- **AI infrastructure errors no longer count as scenario failures** (#44, finding 1).
  When the AI CLI errors (e.g. exit 1 on rate-limit exhaustion) or returns an
  unparseable verdict, the scenario is now marked **skipped** instead of failed.
  Skipped results are excluded from confidence scoring, the failures list, and the
  exit code, and are reported separately (`exec_summary.skipped`,
  `executions[].skipped`). Previously a rate-limited run fabricated "failures" that
  dragged confidence down (e.g. 68.7% instead of ~91%) while saying nothing about
  the system under test.
- **Spawned AI runs in a neutral working directory** (#44, finding 3). Direct
  (non-sandboxed) `claude`/`copilot` calls now run from the system temp dir, so they
  no longer inherit teststop's cwd and load the *target project's* `CLAUDE.md` / MCP
  configuration — which could contaminate behavior or fail when those MCP servers
  are unavailable to a subprocess.

### Notes

- Findings 2 (separate AI-mode concurrency cap) and 4 (`--output-format json` for
  structured rate-limit/error detection) from #44 are tracked as follow-up issues.

---

## [v0.3.0] — 2026-06-07

### Added

- **Concurrency exec mode** (#43) — `ExecSpec.concurrency`: when `> 1`, the HTTP
  executor fires N identical requests simultaneously and asserts the guard yields
  exactly one winner (the rest cleanly rejected). Deterministically verifies race
  guards like double-submit and claim-the-last-item. The mandate now invites the
  AI to emit `concurrency` for race scenarios.

### Changed

- **Reporter honesty** (#42) — runs without `--target` are now clearly labelled as
  **predicted** (risk surface), not executed. Text/Markdown reports show
  "PREDICTED RISKS" / "PREDICTED CONFIDENCE" with a caveat to run `--target` to
  verify; executed runs keep the verified ✓/✗ + CONFIDENCE framing.
- `exec_summary` JSON now carries `executed` (bool) and `count` (int) — previously
  `executed` held the count. Agents should read `executed` as "was this run
  executed against a live target."

### Notes

- Concurrency mode tests guards reachable from the target's current state;
  scenarios needing per-request setup remain future work.

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

[v1.0.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v1.0.0
[v0.3.1]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.3.1
[v0.3.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.3.0
[v0.2.1]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.2.1
[v0.2.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.2.0
[v0.1.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.1.0
