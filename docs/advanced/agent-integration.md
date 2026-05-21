# Agent Integration

teststop is designed to be invoked by AI coding agents — not just humans. This page covers integration patterns for Claude Code, GitHub Copilot, and generic CI/agent workflows.

---

## Core Principle

teststop's output is deliberately machine-first:

- **Structured JSON** — parseable by any agent
- **Numeric exit codes** — no string parsing needed
- **`--quiet` mode** — single-line status (`OK` / `REVIEW` / `CRITICAL` / `ERROR`)
- **`--no-color`** — clean stdout for piping

---

## Minimal Agent Invocation

```bash
teststop run --output json --no-color --quiet
```

| Flag | Why |
|------|-----|
| `--output json` | Structured output for parsing |
| `--no-color` | No ANSI escape codes that break parsing |
| `--quiet` | Suppress human-oriented output; print only status word |

Exit code is the primary signal. Read JSON for details when needed.

---

## Claude Code Integration

Add teststop to your Claude Code workflow in `CLAUDE.md`:

```markdown
## Testing Protocol

After modifying code, run:

```bash
TESTSTOP_SANDBOX=none teststop run --output json --no-color --quiet
```

Exit code meanings:
- 0: Confidence met — safe to proceed
- 1: Review required — include scenario summary in PR
- 2: Critical failures — do not create PR until resolved
- 3: teststop error — skip this run, note in output
```

This tells Claude Code exactly when and how to invoke teststop without any manual prompting.

---

## GitHub Copilot Workspace

Include teststop in a `.github/copilot-instructions.md`:

```markdown
Before creating any pull request, run:

  teststop run --output json --no-color --threshold 80

If exit code is 2, list the critical scenarios in the PR description under "Adversarial Test Findings".
If exit code is 1, add a note "Confidence building — N runs needed for full confidence".
```

---

## GitHub Actions CI

### Basic check

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Claude CLI
        run: |
          # Install claude CLI (or copilot) however you distribute it
          # For Claude: download from releases
          curl -fsSL https://claude.ai/install.sh | sh

      - name: Run teststop
        env:
          TESTSTOP_SANDBOX: none
          TESTSTOP_CLI: claude
        run: |
          teststop run --output json --threshold 80
```

### With PR annotation

```yaml
      - name: Run teststop
        id: ts
        env:
          TESTSTOP_SANDBOX: none
        run: |
          teststop run --output json --quiet > ts-output.json || true
          echo "exit_code=$?" >> $GITHUB_OUTPUT
          echo "summary=$(jq -r '.scenarios | length' ts-output.json) scenarios" >> $GITHUB_OUTPUT

      - name: Comment on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const exitCode = '${{ steps.ts.outputs.exit_code }}';
            const summary = '${{ steps.ts.outputs.summary }}';
            const emoji = exitCode === '0' ? '✅' : exitCode === '2' ? '🚨' : '⚠️';
            
            github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body: `${emoji} **teststop**: ${summary} — exit ${exitCode}`
            });
```

---

## Parsing JSON Output

### Shell (with jq)

```bash
# Get the count of critical scenarios
teststop run --output json | jq '[.scenarios[] | select(.priority == "critical")] | length'

# List all failure modes
teststop run --output json | jq -r '.scenarios[].failure_modes[]'

# Get confidence score
teststop run --output json | jq '.confidence_score'

# Check for specific area
teststop run --output json | jq '.volatile_areas'
```

### Python

```python
import subprocess, json

result = subprocess.run(
    ["teststop", "run", "--output", "json", "--no-color"],
    capture_output=True, text=True
)

if result.returncode == 3:
    raise RuntimeError("teststop error: " + result.stderr)

data = json.loads(result.stdout)
exit_code = result.returncode
scenarios = data["scenarios"]
confidence = data["confidence_score"]

print(f"Confidence: {confidence:.0%}, Scenarios: {len(scenarios)}, Exit: {exit_code}")
```

### Go

```go
import (
    "encoding/json"
    "os/exec"
    "github.com/shaifulshabuj/teststop/internal/reporter"
)

cmd := exec.Command("teststop", "run", "--output", "json", "--no-color")
out, err := cmd.Output()
// exit code is in err.(*exec.ExitError).ExitCode()

var result reporter.RunResult
json.Unmarshal(out, &result)
```

---

## JSON Output Schema

The full `RunResult` JSON structure:

```json
{
  "project_name": "string",
  "project_path": "string",
  "language": "string",
  "system_type": "string",
  "timestamp": "RFC3339",
  "duration_ms": 42000,
  "adapter": "claude",
  "depth": "normal",
  "scenarios": [ /* []Scenario — see Scenario Schema */ ],
  "failures": [
    {
      "scenario_id": "TS-001",
      "title": "string",
      "area": "string",
      "priority": "critical",
      "description": "failure mode 1; failure mode 2"
    }
  ],
  "stable_areas": ["payments", "auth"],
  "volatile_areas": ["checkout"],
  "retired_areas": [],
  "confidence_score": 0.62,
  "exit_code": 1
}
```

---

## Committing Memory in CI

For teams where CI updates confidence automatically:

```yaml
      - name: Commit updated memory
        if: github.ref == 'refs/heads/main'
        run: |
          git config user.name "teststop-bot"
          git config user.email "bot@noreply"
          git add .teststop/memory.json .teststop/retired.json
          git diff --staged --quiet || git commit -m "chore(teststop): update confidence memory [skip ci]"
          git push
```

This keeps the shared confidence baseline up to date as CI runs accumulate.
