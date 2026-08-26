---
description: Create a GitHub PR for completed teststop work. Handles branch creation, commit, push, and PR opening with proper template. Invoke with: /teststop-pr <phase> "<description>"
---

# teststop-pr

Create a properly formatted pull request for completed teststop work.

## Usage
```
/teststop-pr 2 "Implement mandate engine"
/teststop-pr "0-2" "Phase 0-2: infrastructure, foundation, mandate"
```

## What I Do

### 1. Pre-PR Quality Gate
```bash
go build ./...          # must pass
go test -race ./...     # must pass
go vet ./...            # must pass
```
If any fails: **stop and fix before creating PR**.

### 2. Create Branch
```bash
git checkout -b feat/phase-<N>-<slug>
# e.g.: feat/phase-2-mandate-engine
```

### 3. Stage and Commit
```bash
git add -A
git status   # review what's being committed
git commit -m "feat(phase-<N>): <description>

- List key changes
- Files added/modified

Implements: #<issue-numbers>

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

### 4. Push
```bash
git push -u origin feat/phase-<N>-<slug>
```

### 5. Open PR
```bash
gh pr create \
  --title "feat(phase-<N>): <description>" \
  --body "$(cat <<'EOF'
## Summary
<brief description of what this phase implements>

## Related Issues
Closes #<n1>, Closes #<n2>, ...

## Changes Made
- pkg/scenario/types.go — Scenario contract
- internal/mandate/ — composer + templates
- mandate/base.md — adversarial user mandate

## Testing
- [x] `go build ./...` passes
- [x] `go test -race ./...` passes
- [x] `go vet ./...` passes

## Mandate Impact
<does/does not affect mandate behavior>

## The One Question
Does this make the user's problem easier to solve? ✅ Yes
EOF
)" \
  --base main
```

### 6. Report PR URL
Print the PR URL for the user to review.

## Branch Naming
- `feat/phase-0-infra`
- `feat/phase-1-foundation`
- `feat/phase-2-mandate`
- `feat/phase-3-reader`
- `feat/phase-4-memory`
- `feat/phase-5-ai`
- `feat/phase-6-reporter`
- `feat/phase-7-wiring`
- `feat/phase-0-7-v01-complete` (for a combined PR)
