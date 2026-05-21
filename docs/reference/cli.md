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
```

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
