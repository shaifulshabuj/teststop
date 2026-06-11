# teststop × waymark demo

Adversarial AI testing of the [waymark](https://github.com/shaifulshabuj/waymark) MCP
middleware using teststop's real AI-driven path. Captured 2026-06-11.

## What's in this directory

| File | Description |
|------|-------------|
| `mandate.md` | The exact mandate sent to the AI — composed from waymark's scanned context + memory |
| `run_report.json` | Full JSON output from `teststop run` (15 scenarios, exit code 1 = review needed) |
| `run_report.md` | Markdown summary auto-saved by teststop |

## How to replay

Requires: `teststop` ≥ v1.0.1 and `claude` CLI on PATH, authenticated.

```bash
# From this repo root:
teststop run --path ../waymark --depth light --output json

# Or for a text summary:
teststop run --path ../waymark --depth light
```

**Notes:**
- waymark is MCP middleware, not an HTTP service — **do not use `--target`**; teststop
  runs in static-validation mode (predicts risks, does not execute HTTP calls)
- Confidence state accumulates in `../waymark/.teststop/memory.json`; each run the AI
  focuses more on areas with lower confidence scores
- `--depth light` produces a shorter mandate (~200 lines vs ~617 for `normal`) which
  reduces AI response time and token cost; use `normal` or `aggressive` for deeper coverage

## About this run

```
adapter:    claude
depth:      light
scenarios:  15  (re-captured 2026-06-11 after fix/blank-scenarios; prior artifact had hollow structs)
exit code:  1   (confidence below 80% threshold — memory partially populated from prior runs)
```

The exit code of 1 is expected for a first run: teststop's confidence memory starts at
zero and rises toward the 0.95 retirement threshold as scenarios repeatedly pass. Run
`teststop status --path ../waymark` to see the current memory state.

## Mandate source

The mandate in `mandate.md` was produced by:

```bash
teststop mandate --show ../waymark
```

It inlines: (1) the embedded `base.md` adversarial instruction, (2) the project context
scanned from waymark's source tree (language, entry points, routes, flows), and (3) the
current confidence-memory state so the AI focuses on volatile areas.
