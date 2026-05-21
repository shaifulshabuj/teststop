---
description: Implement a complete teststop phase from its GitHub issues. Handles Go implementation, tests, and closes issues. Invoke with: /implement-phase <phase-number>
---

# implement-phase

Implement a complete teststop phase end-to-end.

## Usage
```
/implement-phase 0     # Phase 0: infrastructure
/implement-phase 1     # Phase 1: foundation (scenario types + CLI scaffold)
/implement-phase 2     # Phase 2: mandate engine
/implement-phase 3     # Phase 3: reader (code scanner)
/implement-phase 4     # Phase 4: memory layer
/implement-phase 5     # Phase 5: AI adapter
/implement-phase 6     # Phase 6: reporter
/implement-phase 7     # Phase 7: wire-up & integration
```

## What I Do

1. **Fetch issues** for the requested phase from GitHub:
   ```bash
   gh issue list --label "phase/<N>-*" --state open --json number,title,body --limit 20
   ```

2. **Read the issue body** for acceptance criteria and implementation details

3. **Delegate to go-implementer subagent** for each implementation task

4. **Delegate to mandate-writer subagent** for Phase 2 (mandate/base.md)

5. **Run quality gates** after each file:
   ```bash
   go build ./...   # must pass after every file write
   ```

6. **Run full test suite** after the phase is complete:
   ```bash
   go test -race ./...
   go vet ./...
   ```

7. **Close issues** as each one is completed:
   ```bash
   gh issue close <number> --comment "Implemented in commit <sha>. go test ./... passes."
   ```

8. **Report phase completion** with:
   - Files created/modified
   - Test results
   - Issues closed
   - Next recommended phase

## Phase Order (dependencies)
```
P0 (infra) → P1 (foundation) → P2 (mandate) + P3 (reader) + P4 (memory) + P5 (ai) + P6 (reporter) → P7 (wiring)
```
P2, P3, P4, P5, P6 can be done in any order after P1.
P7 requires ALL of P2-P6 complete.

## Quality Standard
Never mark a phase complete unless:
- `go build ./...` exits 0
- `go test -race ./...` exits 0
- All acceptance criteria from the GitHub issues are met
