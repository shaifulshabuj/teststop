# Changelog

All notable changes to teststop are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

### v0.2 (planned)

- **Scenario executor** — actually run generated scenarios against a live system
- **Ollama adapter** — local model support via `TESTSTOP_CLI=ollama`
- **`teststop watch`** — file-watching mode that re-runs on code changes

### v1.0 (planned)

- **Waymark integration** — governance hooks for AI agent workflows
- **DocuFlow integration** — feed project documentation into mandate context
- **CI/CD plugins** — native GitHub Actions, GitLab CI support
- **`teststop diff`** — scenario comparison between runs

---

[v0.1.0]: https://github.com/shaifulshabuj/teststop/releases/tag/v0.1.0
