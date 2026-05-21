# DevLoop Project Configuration — edit to match your stack

PROJECT_NAME="$(basename "$PWD")"
PROJECT_STACK="Go"
PROJECT_PATTERNS="SOLID, Clean Architecture"
PROJECT_CONVENTIONS="explicit error checks, small interfaces, idiomatic packages"
TEST_FRAMEWORK="go test"

# Provider routing
# main  = orchestrator / architect / reviewer (requires remote control: claude | copilot)
# worker = work / fix (any CLI provider: claude | copilot | opencode | pi)
# opencode and pi are worker-only — they have no remote-control support
DEVLOOP_MAIN_PROVIDER="claude"
DEVLOOP_WORKER_PROVIDER="copilot"

# Auto-failover: when a provider hits its rate limit, DevLoop automatically
# switches to the next provider in the chain and restores as soon as available.
# Main chain:   claude → copilot
# Worker chain: copilot → opencode → pi
DEVLOOP_FAILOVER_ENABLED="true"
DEVLOOP_PROBE_INTERVAL="5"   # minutes between availability probes on limited providers

# Smart permission system
# smart  — BLOCK dangerous, ALLOW safe ops, ESCALATE unknown to user (default)
# auto   — ALLOW everything (fastest, no interruptions to the pipeline)
# strict — ALLOW only known-safe ops, BLOCK everything else
# off    — disable permission hook (Claude's built-in behaviour applies)
DEVLOOP_PERMISSION_MODE="smart"
DEVLOOP_PERMISSION_TIMEOUT="60"  # seconds to wait for user response before auto-deny

# Worker mode
# cli          — use copilot or claude CLI locally (default)
# github-agent — create a GitHub Issue; Copilot coding agent works on it and opens a PR
DEVLOOP_WORKER_MODE="cli"

# Claude model settings
# CLAUDE_MODEL is the base default used by all Claude roles.
# Override per-role to use different models for main (architect/reviewer) vs worker.
#   "sonnet" = faster/cheaper   "opus" = more capable   "haiku" = fastest
CLAUDE_MODEL="sonnet"
# CLAUDE_MAIN_MODEL="opus"     # architect, reviewer, orchestrator (uncomment to override)
# CLAUDE_WORKER_MODEL="sonnet" # worker and fix passes (uncomment to override)

# Copilot model: the Copilot CLI does not expose a --model flag for non-interactive use.
# The model is determined by your GitHub Copilot subscription and plan settings.
# To change the Copilot model, update it in: https://github.com/settings/copilot

# Version checks and self-update use GitHub by default (no config needed).
# DEVLOOP_GITHUB_REPO="shaifulshabuj/devloop"   # override to use a fork
# Override with a custom VERSION file URL (plain semver text):
# DEVLOOP_VERSION_URL="https://raw.githubusercontent.com/you/devloop/main/VERSION"
# Override with a custom script URL for 'devloop update':
# DEVLOOP_SOURCE_URL="https://raw.githubusercontent.com/you/devloop/main/devloop.sh"
