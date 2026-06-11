# teststop v1.0 Contracts

This document freezes the public-facing contracts of teststop v1.0.
**Breaking changes require a major version bump.** Additive changes (new fields,
new flags with safe defaults) are allowed in minor/patch releases.

---

## Exit Codes

| Code | Meaning | Agent action |
|------|---------|-------------|
| `0` | Confidence threshold met — all scenarios passed (or confidence ≥ threshold) | Safe to proceed / deploy |
| `1` | Below threshold — confidence below threshold or non-critical failures found | Human review required |
| `2` | Critical failures — at least one `critical`-priority scenario failed | **Do NOT deploy** |
| `3` | teststop internal error — scan, mandate, AI CLI, or I/O failure | Debug teststop, not the target |

Exit codes are stable. Agents and CI pipelines may gate on them unconditionally.

---

## Scenario Schema (`pkg/scenario/types.go`)

The JSON schema produced by `teststop run --output json` under the `"scenarios"` key is
frozen. All fields below are present in every scenario object (omitempty fields may be
absent when the value is a zero-value).

```json
{
  "scenario_id":       "string  — stable ID for deduplication and memory keying",
  "title":             "string  — human-readable short description",
  "user_perspective":  "string  — who the actor is and what they want",
  "preconditions":     ["string, ..."],
  "steps":             ["string, ..."],
  "chaos_factors":     ["string, ..."],
  "expected_behavior": "string  — what a correct system should do",
  "failure_modes":     ["string, ..."],
  "priority":          "critical | high | medium | low",
  "confidence_area":   "string  — memory key; area of the system under test",
  "is_edge_case":      "bool",
  "exec":              "optional ExecSpec object (v0.2+, omitted when absent)"
}
```

**Stability rules:**
- Fields above may NOT be removed or renamed in v1.x.
- The `exec` sub-object (v0.2 additive extension) may gain new optional fields.
- New top-level optional fields may be added in minor releases without a major bump.

---

## JSON Run Output (`--output json`)

Top-level fields in the JSON envelope are stable:

| Field | Type | Notes |
|-------|------|-------|
| `project_name` | string | |
| `project_path` | string | |
| `language` | string | |
| `system_type` | string | |
| `timestamp` | RFC3339 | |
| `duration_ms` | number | milliseconds |
| `scenarios` | array | see Scenario schema above |
| `executions` | array | present when `--target` set (v0.2+) |
| `exec_summary` | object | `executed`, `count`, `passed`, `failed`, `skipped`, `target` |
| `failures` | array | `scenario_id`, `title`, `area`, `priority`, `description` |
| `stable_areas` | array of string | |
| `volatile_areas` | array of string | |
| `retired_areas` | array of string | |
| `exit_code` | number | mirrors process exit code |
| `confidence_score` | number | 0.0–1.0 |
| `adapter_name` | string | `claude` or `copilot` |
| `depth` | string | `light`, `normal`, or `aggressive` |

---

## Memory File Format (`.teststop/memory.json`)

The memory file is stable. AI agents and CI pipelines may read and commit it.
Keys are `confidence_area` strings; values contain `confidence` (float64),
`pass_count`, `fail_count`, `retired` (bool), and `last_updated` (RFC3339).

---

## Environment Variables

The following variables are stable CLI contracts:

| Variable | Description |
|----------|-------------|
| `TESTSTOP_CLI` | `auto` \| `claude` \| `copilot` — AI CLI selection |
| `TESTSTOP_MODEL` | optional model flag passed to the AI CLI |
| `TESTSTOP_SANDBOX` | `auto` \| `required` \| `none` — container isolation mode |
| `TESTSTOP_RUN_DEPTH` | override for `--depth` |
| `TESTSTOP_RUN_OUTPUT` | override for `--output` |
| `TESTSTOP_RUN_THRESHOLD` | override for `--threshold` |
| `TESTSTOP_RUN_NO_COLOR` | override for `--no-color` |
| `TESTSTOP_RUN_QUIET` | override for `--quiet` |
| `TESTSTOP_RUN_TARGET` | override for `--target` |
| `TESTSTOP_RUN_CONCURRENCY` | override for `--concurrency` |
| `TESTSTOP_RUN_AI_CONCURRENCY` | override for `--ai-concurrency` |
| `TESTSTOP_RUN_EXEC_TIMEOUT` | override for `--exec-timeout` |
| `TESTSTOP_RUN_MAX_RETRIES` | override for `--max-retries` |
