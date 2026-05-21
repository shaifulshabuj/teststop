# Environment Variables

All teststop behavior can be controlled via environment variables — no config file needed.

---

## Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `TESTSTOP_CLI` | `auto` | Which AI CLI to use |
| `TESTSTOP_MODEL` | _(empty)_ | Model to pass to the AI CLI |
| `TESTSTOP_SANDBOX` | `auto` | Sandbox isolation mode |

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
| `claude` | ✅ Yes | `TESTSTOP_MODEL=claude-opus-4-7` |
| `copilot` | ❌ Ignored | Model is controlled by GitHub Copilot subscription |

```bash
TESTSTOP_MODEL=claude-sonnet-4-6 teststop run
TESTSTOP_MODEL=claude-opus-4-7 teststop run --depth aggressive
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

---

## No Config File

teststop intentionally has no configuration file. The six design non-negotiables include:

> **ZERO CONFIGURATION** — `teststop run` must work with no setup on any project.

Environment variables are the only customization layer.
