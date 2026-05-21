# Sandbox Isolation

teststop can run the AI CLI inside an isolated Apple Container VM. This prevents the AI from accessing your host filesystem beyond the mounted project directory.

---

## Why Sandbox?

When teststop calls an AI CLI, that process inherits your user's full permissions by default. The AI itself doesn't do anything malicious — but isolation is better security hygiene, especially in sensitive environments.

The sandbox ensures:

- The AI cannot read your SSH keys, AWS credentials, or other projects
- Credential paths are mounted **read-only**
- The container is **ephemeral** — destroyed after each run (`--rm`)

---

## How It Works

```
teststop run
  └─ sandbox.Runner.Run(mandate)
       ├─ [container available] → container run --rm teststop-agent:latest claude -p "..."
       │       Isolated VM: AI receives mandate, outputs JSON → captured by teststop
       └─ [no container] → exec.Command("claude", "-p", mandate)   ← direct fallback
```

The runtime image (`teststop-agent`) is a minimal Ubuntu 24.04 container with only the AI CLIs installed — no Go, no dev tools.

---

## Sandbox Modes

Control sandbox behavior with `TESTSTOP_SANDBOX`:

| Mode | Value | Behavior |
|------|-------|----------|
| Auto _(default)_ | `auto` | Use container if `container system status` reports running; fall back to direct otherwise |
| Required | `required` | Error if container is not available — use in high-security environments |
| Disabled | `none` | Always run AI CLI directly — use in CI, Docker-in-Docker, Linux |

```bash
# Always run direct (CI environments, non-macOS)
TESTSTOP_SANDBOX=none teststop run

# Enforce isolation — error if container unavailable
TESTSTOP_SANDBOX=required teststop run
```

---

## Installation (macOS only)

Apple Container requires macOS with Apple Silicon or Intel.

```bash
brew install container
container system start   # starts on boot after first run
```

Verify it's running:

```bash
container system status
# → running
```

---

## Credential Mounts

When running in a container, teststop mounts credentials read-only:

| Host Path | Container Path | Purpose |
|-----------|----------------|---------|
| `~/.claude` | `/root/.claude:ro` | Claude Code authentication |
| `~/.config/gh` | `/root/.config/gh:ro` | GitHub Copilot CLI token |

The `:ro` flag makes these mounts read-only — the container cannot write back to your host credentials.

---

## What the AI Cannot Access in Container Mode

| Resource | Accessible? |
|----------|-------------|
| Your project directory (via `--path`) | ✅ Read (v0.1 — mandate only) |
| `~/.claude` and `~/.config/gh` | ✅ Read-only |
| `~/.ssh` | ❌ No |
| Other projects on your machine | ❌ No |
| `~/.aws`, `~/.zshrc`, system files | ❌ No |
| Host network | ✅ (needed for AI CLI to call APIs) |

---

## Container Name

Each run gets a unique container name: `teststop-<8 random hex chars>`. This ensures concurrent runs don't interfere with each other.

---

## Fallback Behavior

If `TESTSTOP_SANDBOX=auto` (the default) and the container daemon is not running, teststop falls back silently to running the AI CLI directly. No error, no warning — it just works.

To see which mode was used, check the run output header or the markdown report.

---

## CI/CD and Linux

Containers are macOS-only. For CI environments, set:

```bash
TESTSTOP_SANDBOX=none
```

Or it will auto-detect that no container is available and fall back to direct.

GitHub Actions example:

```yaml
- name: Run teststop
  env:
    TESTSTOP_SANDBOX: none
  run: teststop run --output json --threshold 80
```
