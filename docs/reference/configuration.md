# Configuration

teststop is **zero-config by default** — `teststop run` works on any project with no setup.
When you need project-specific tuning, two optional layers are available:
an optional `.teststop/config.yaml` file and `TESTSTOP_RUN_*` environment variables.

---

## Precedence (lowest → highest)

```
.teststop/config.yaml  <  TESTSTOP_RUN_* env var  <  explicit CLI flag
```

An explicit CLI flag always wins. A missing config file is never an error.

---

## `.teststop/config.yaml`

An optional per-project configuration file. All keys map one-to-one onto `teststop run` flags — there are no settings here that don't also exist as flags. A missing file is ignored; a malformed file or unknown key is a hard error.

```yaml
# .teststop/config.yaml — all keys are optional
depth: normal          # --depth
output: text           # --output
threshold: 80          # --threshold
no_color: false        # --no-color
quiet: false           # --quiet
target: ""             # --target
concurrency: 4         # --concurrency
exec_timeout: 10s      # --exec-timeout
max_retries: 2         # --max-retries
```

A reference copy lives at `.teststop/config.example.yaml` in your project after `teststop run` creates the `.teststop/` directory.

---

## `TESTSTOP_RUN_*` Environment Variables

These variables mirror every `teststop run` flag. They override config.yaml but yield to an explicit CLI flag. Malformed values (wrong type, bad duration) are hard errors.

| Variable | Flag equivalent | Type | Default |
|----------|----------------|------|---------|
| `TESTSTOP_RUN_DEPTH` | `--depth` | string | `normal` |
| `TESTSTOP_RUN_OUTPUT` | `--output` | string | `text` |
| `TESTSTOP_RUN_THRESHOLD` | `--threshold` | integer | `80` |
| `TESTSTOP_RUN_NO_COLOR` | `--no-color` | boolean | `false` |
| `TESTSTOP_RUN_QUIET` | `--quiet` | boolean | `false` |
| `TESTSTOP_RUN_TARGET` | `--target` | string | _(none)_ |
| `TESTSTOP_RUN_CONCURRENCY` | `--concurrency` | integer | `4` |
| `TESTSTOP_RUN_EXEC_TIMEOUT` | `--exec-timeout` | duration | `10s` |
| `TESTSTOP_RUN_MAX_RETRIES` | `--max-retries` | integer | `2` |

```bash
# Example: set depth and output in CI without touching flags
TESTSTOP_RUN_DEPTH=aggressive TESTSTOP_RUN_OUTPUT=json teststop run
```

---

## AI / Sandbox Environment Variables

These variables control the AI adapter and sandbox isolation (not `run`-specific):

| Variable | Default | Description |
|----------|---------|-------------|
| `TESTSTOP_CLI` | `auto` | Which AI CLI to use: `auto` \| `claude` \| `copilot` |
| `TESTSTOP_MODEL` | _(empty)_ | Model flag passed to the AI CLI |
| `TESTSTOP_SANDBOX` | `auto` | Sandbox mode: `auto` \| `required` \| `none` |

---

## `TESTSTOP_CLI`

Controls which AI adapter teststop uses.

| Value | Behavior |
|-------|----------|
| `auto` _(default)_ | Try `claude` first, then `copilot`; error if neither found |
| `claude` | Use Claude CLI; error if not on PATH |
| `copilot` | Use GitHub Copilot CLI; error if not on PATH |

```bash
TESTSTOP_CLI=claude teststop run
TESTSTOP_CLI=copilot teststop run
```

---

## `TESTSTOP_MODEL`

Passed as `--model <value>` to the AI CLI. Only applies to adapters that support model selection.

| CLI | Support | Example |
|-----|---------|---------|
| `claude` | ✅ Yes | `TESTSTOP_MODEL=claude-opus-4-8` |
| `copilot` | ❌ Ignored | Model is controlled by GitHub Copilot subscription |

```bash
TESTSTOP_MODEL=claude-sonnet-4-6 teststop run
TESTSTOP_MODEL=claude-opus-4-8 teststop run --depth aggressive
```

If not set, the Claude CLI uses its configured default model.

---

## `TESTSTOP_SANDBOX`

Controls whether teststop runs the AI CLI inside an Apple Container VM.

| Value | Behavior |
|-------|----------|
| `auto` _(default)_ | Use container if `container system status` reports running; direct otherwise |
| `required` | Error immediately if container is not running |
| `none` | Always run directly — bypass container detection |

```bash
# CI/CD environments (container not available)
TESTSTOP_SANDBOX=none teststop run

# Enforce isolation — fail if container unavailable
TESTSTOP_SANDBOX=required teststop run
```

See [Sandbox Isolation](../guide/sandbox.md) for detailed explanation.

---

## Setting Variables

### Per command

```bash
TESTSTOP_CLI=claude TESTSTOP_SANDBOX=none teststop run
```

### In `.env` files (with direnv or similar)

```bash
# .envrc
export TESTSTOP_CLI=claude
export TESTSTOP_MODEL=claude-sonnet-4-6
export TESTSTOP_SANDBOX=none
```

### In GitHub Actions

```yaml
- name: Run teststop
  env:
    TESTSTOP_CLI: claude
    TESTSTOP_SANDBOX: none
  run: teststop run --output json --threshold 80
```

### In shell profile (permanent)

```bash
# ~/.zshrc or ~/.bashrc
export TESTSTOP_CLI=claude
export TESTSTOP_SANDBOX=none
```
