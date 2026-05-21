# Exit Codes

teststop exits with a structured exit code after every run. These codes are designed to be consumed by CI pipelines and AI coding agents.

---

## Exit Code Table

| Code | Name | Meaning | Recommended Action |
|------|------|---------|-------------------|
| `0` | OK | Confidence threshold met | Safe to proceed with deployment |
| `1` | REVIEW | Below confidence threshold | Human or agent review required before deploying |
| `2` | CRITICAL | Critical failures found | Do NOT deploy — investigate before proceeding |
| `3` | ERROR | teststop internal error | Debug teststop itself (not your project) |

---

## Code 0 — OK

The average confidence score across all active areas equals or exceeds the threshold (default: 80%).

```bash
teststop run
echo $?   # → 0
```

This is the "green light" signal. Safe to merge, safe to deploy.

---

## Code 1 — REVIEW

The average confidence score is below the threshold. This is **expected on the first run** — confidence starts at 0% and builds with each pass.

```bash
teststop run
echo $?   # → 1
```

This is a signal to review, not necessarily to block. The scenarios generated are valuable inputs even when confidence is low.

**Common causes:**
- First run on a project (no confidence history)
- Recently reset memory
- A new system area was discovered in this run
- Threshold set higher than current confidence

---

## Code 2 — CRITICAL

Critical-priority scenarios have `failure_modes` populated. This indicates the AI identified failure patterns that could cause data loss, security issues, or significant user impact.

```bash
teststop run
echo $?   # → 2
```

teststop will never exit `2` without providing the specific scenarios that triggered it. Review the output to understand what was flagged.

**Do not deploy** when exit code is `2` without investigating the flagged scenarios.

---

## Code 3 — ERROR

An internal error occurred before teststop could complete its run. This is a teststop problem, not a problem with your project.

**Common causes:**
- AI CLI not found on PATH
- AI CLI failed to authenticate
- AI CLI returned unparseable output
- Filesystem permission error on `.teststop/`

---

## Using Exit Codes in CI

### GitHub Actions

```yaml
- name: Run adversarial tests
  run: teststop run --output json --threshold 80
  # Non-zero exit fails the step automatically
```

```yaml
- name: Run adversarial tests
  id: teststop
  run: teststop run --output json --quiet
  continue-on-error: true

- name: Block on critical failures
  if: steps.teststop.outputs.exit_code == '2'
  run: |
    echo "CRITICAL failures found. Blocking deployment."
    exit 1
```

### Shell Scripts

```bash
#!/usr/bin/env bash
teststop run --output json --quiet
EXIT=$?

case $EXIT in
  0) echo "All good. Deploying..." ;;
  1) echo "Review required. Pausing for human check." ; exit 1 ;;
  2) echo "CRITICAL failures. Aborting deployment." ; exit 1 ;;
  3) echo "teststop error. Check your AI CLI setup." ; exit 1 ;;
esac
```

---

## Using Exit Codes in AI Agent Workflows

```bash
# Minimal agent-friendly invocation
teststop run --output json --no-color --quiet

# The agent reads exit code to decide next action:
# 0 → proceed with PR / merge
# 1 → include review note in PR description
# 2 → block merge, add critical label
# 3 → log error, skip teststop this run
```

---

## Threshold Tuning

The exit code `0` threshold defaults to 80%. Adjust it with `--threshold`:

```bash
# Strict: require 90% confidence to exit 0
teststop run --threshold 90

# Lenient: exit 0 at 60% (useful when building initial confidence)
teststop run --threshold 60

# Exit 0 as soon as any confidence exists (only blocks on critical)
teststop run --threshold 1
```

!!! tip "Start lenient, tighten over time"
    On a new project, set `--threshold 60` until confidence builds across all areas.
    Once `teststop status` shows most areas in `mature` stage, raise to 80–90%.
