# CLI Commands

Complete reference for all teststop commands and flags.

---

## `teststop run`

Run adversarial user testing on a project.

```
teststop run [path] [flags]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `path` | Path to the project. Overrides `--path` if both are provided. |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--path <dir>` | `.` | Path to the project to test |
| `--depth <level>` | `normal` | Testing depth: `light` \| `normal` \| `aggressive` |
| `--output <format>` | `text` | Output format: `json` \| `text` \| `markdown` |
| `--threshold <n>` | `80` | Confidence threshold 0–100. Exit `0` when average confidence ≥ this. |
| `--no-color` | `false` | Disable ANSI color output (useful for agents reading stdout) |
| `--quiet` | `false` | Minimal output — prints only `OK`, `REVIEW`, `CRITICAL`, or `ERROR` |
| `--target <url>` | _(none)_ | Base URL of a **running** system to execute scenarios against. Empty = static validation only. |
| `--concurrency <n>` | `4` | Max scenarios executed in parallel |
| `--ai-concurrency <n>` | `1` | Max concurrent AI-mode executions. Keep at 1 to avoid rate-limit exhaustion; increase only when the AI backend supports higher parallelism. |
| `--exec-timeout <dur>` | `10s` | Per-request execution timeout (e.g. `15s`, `500ms`) |
| `--max-retries <n>` | `2` | Retries for transient HTTP execution failures (transport errors, 5xx) |

### Examples

```bash
# Run on current directory
teststop run

# Run on a specific path
teststop run ./api

# Equivalent with flag
teststop run --path ./api

# Aggressive depth for pre-release testing
teststop run --depth aggressive --threshold 90

# Agent-friendly: JSON output, no color, quiet
teststop run --output json --no-color --quiet

# Markdown report to stdout
teststop run --output markdown

# Execute scenarios against a running system
TESTSTOP_SANDBOX=none teststop run --target http://localhost:8080

# Tune execution against a staging instance
teststop run --target https://staging.example.com \
  --concurrency 8 --exec-timeout 15s --max-retries 3
```

!!! tip "Execution vs. generation"
    Without `--target`, `teststop run` only **generates and validates**
    scenarios. With `--target`, it also **executes** them and feeds real
    pass/fail into confidence memory. See [Execution](../guide/execution.md).

### Testing Depth

| Depth | Scenarios | Use When |
|-------|-----------|----------|
| `light` | 3–4 | Quick sanity check during development |
| `normal` | 5–6 | Standard CI run or pre-merge check |
| `aggressive` | 7–9 | Pre-release, major refactors, or first run on a new codebase |

---

## `teststop status`

Show the confidence state of all tracked areas.

```
teststop status [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--path <dir>` | `.` | Path to the project |

### Output

```
Area         Confidence   Tests   Maturity   Status
──────────────────────────────────────────────────────────
auth         62%          5       growing    active
checkout     38%          2       new        active
api-routes   91%          12      legacy     active
payments     95%          15      legacy     retired
```

---

## `teststop memory`

Show or reset the accumulated confidence memory.

```
teststop memory [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--path <dir>` | `.` | Path to the project |
| `--reset` | `false` | Reset memory (asks for confirmation) |
| `--yes` | `false` | Skip confirmation prompt (use with `--reset`) |

### Examples

```bash
# Show memory as pretty-printed JSON
teststop memory

# Reset with confirmation prompt
teststop memory --reset

# Reset without prompt (for scripts)
teststop memory --reset --yes
```

!!! warning "Irreversible"
    `--reset` permanently deletes `memory.json` and `retired.json`. This cannot be undone unless you have the files in version control.

---

## `teststop report`

Show the most recent run report.

```
teststop report [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--path <dir>` | `.` | Path to the project |
| `--format <fmt>` | `text` | Output format: `text` \| `md` |

### Examples

```bash
# Show last report in terminal
teststop report

# Show last report as markdown
teststop report --format md
```

Reports are stored in `.teststop/reports/YYYY-MM-DD-HH-MM-SS.md` and are available after any `teststop run`.

---

## `teststop mandate`

Show the exact mandate that will be sent to the AI.

```
teststop mandate [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--show` | required | Print the mandate (required flag) |
| `--path <dir>` | `.` | Path to the project to scan |
| `--depth <level>` | `normal` | Testing depth (affects scenario count `[N]`) |

### Examples

```bash
# Show the mandate for the current directory
teststop mandate --show

# Show the mandate for a specific project with aggressive depth
teststop mandate --show --path ./api --depth aggressive
```

!!! tip
    Use this to understand exactly what instructions the AI receives. It's the best way to debug unexpected scenario output or to iterate on the mandate.

---

## `teststop version`

Print the version and build information.

```
teststop version
```

```console
$ teststop version
teststop v0.3.1
  commit:  a1b2c3d
  built:   2026-06-11T00:00:00Z
  go:      go1.26.3
  os/arch: darwin/arm64
```

The version is injected at release time by GoReleaser. For `go install` builds it
is recovered from the module build info, so it still reports the installed tag.
The same value is available via the `--version` / `-v` flag:

```bash
teststop --version    # teststop v0.3.1
teststop -v           # teststop v0.3.1
```

---

## Global Flags

These flags apply to all commands:

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help for a command |
| `--version`, `-v` | Show teststop version |

---

## Environment Variable Override

All key behaviors can be overridden via environment variables without modifying flags:

```bash
TESTSTOP_CLI=claude          # Force Claude adapter
TESTSTOP_MODEL=claude-opus-4-7  # Override AI model
TESTSTOP_SANDBOX=none        # Disable container isolation
```

See [Environment Variables](configuration.md) for the full list.
