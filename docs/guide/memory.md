# Memory System

teststop maintains a **confidence score per system area** that persists across runs. This is the mechanism that makes tests reduce over time as confidence builds.

---

## Overview

When teststop runs, it updates a JSON file at `.teststop/memory.json`. Each system area (e.g., `auth`, `checkout`, `api`) gets a confidence score between 0.0 and 1.0.

The principle is simple:

- **Passes** increase confidence
- **Failures** decrease confidence  
- **Proven areas** are retired and stop being tested aggressively
- **New or changed areas** start at 0% and require proof before they're trusted

---

## Confidence Scoring Formula

teststop uses **exponential approach** — the same math used in signal processing to converge on a stable value.

### Pass (confidence increase)

```
new = old + PassWeight × (1.0 − old)
```

Where `PassWeight = 0.19`.

This means:
- Each pass brings confidence 19% closer to 1.0
- The first pass: 0.0 → 0.19
- Second pass: 0.19 → 0.34
- Fifth pass: ~0.62
- Tenth pass: ~0.87
- **Fifteenth pass: ~0.96 (retirement threshold: 0.95)**

### Failure (confidence decrease)

```
new = old − FailPenalty
```

Where `FailPenalty = 0.30`.

A single failure drops confidence by 0.30 — significant. An area that has been building confidence for 10 runs can be sent back to "needs attention" status by one failure.

### Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `PassWeight` | `0.19` | Confidence increase per pass |
| `FailPenalty` | `0.30` | Confidence drop per failure |
| `RetirementThreshold` | `0.95` | Retire when confidence ≥ this |
| `VolatileThreshold` | `0.75` | Focus testing below this |
| `StableThreshold` | `0.95` | Reduce testing at or above this |

---

## Maturity Stages

Each area has a maturity stage that reflects where it is in the confidence journey:

| Stage | Confidence Range | Meaning |
|-------|-----------------|---------|
| `new` | < 40% | Untested or recently failed — needs full attention |
| `growing` | 40% – 70% | Building confidence — still needs regular testing |
| `mature` | 70% – 90% | Well-tested — lighter coverage is appropriate |
| `legacy` | ≥ 90% | Proven reliable — minimal additional testing needed |

---

## Retirement

An area is **retired** when both conditions are met:

1. Confidence ≥ 0.95
2. Test count ≥ 15

The dual gate prevents false retirement from a lucky streak on a new, under-tested area. After ~15 clean passes (~95% confidence), the area is genuinely proven.

Retired areas are:
- Written to `.teststop/retired.json`
- Marked `"retired": true` in `memory.json`
- Excluded from future mandate generation
- **Never deleted** — they remain as a historical record

If a retired area regresses (introduced in the mandate again with failures), it immediately loses its retired status on the next run.

---

## The Memory File

`.teststop/memory.json` is a human-readable JSON file. Here's an example after several runs:

```json
{
  "areas": {
    "auth": {
      "name": "auth",
      "confidence": 0.6156,
      "test_count": 5,
      "pass_count": 5,
      "fail_count": 0,
      "last_tested_at": "2025-05-21T10:30:00Z",
      "maturity_stage": "growing",
      "retired": false
    },
    "checkout": {
      "name": "checkout",
      "confidence": 0.3439,
      "test_count": 2,
      "pass_count": 2,
      "fail_count": 0,
      "last_tested_at": "2025-05-21T10:30:00Z",
      "maturity_stage": "new",
      "retired": false
    }
  },
  "created_at": "2025-05-21T10:00:00Z",
  "updated_at": "2025-05-21T10:30:00Z",
  "version": 1
}
```

---

## What to Commit

**Commit `memory.json`** — it's the accumulated proof that your system works. Your team shares the same confidence baseline when you check it in.

**Do not commit `.teststop/runs/`** — run history is gitignored by default.

Suggested `.gitignore` additions (teststop adds these automatically):

```gitignore
.teststop/runs/
```

Keep:

```
.teststop/memory.json     ← commit this
.teststop/retired.json    ← commit this
```

---

## Memory Commands

```bash
# Show confidence state as a table
teststop status

# Show raw memory.json (pretty-printed)
teststop memory

# Reset memory (with confirmation)
teststop memory --reset

# Skip confirmation prompt
teststop memory --reset --yes
```

---

## How the Mandate Uses Memory

The memory drives how aggressively each area gets tested:

- **Volatile areas** (< 75% confidence) → the mandate says "focus here, these are unproven"
- **Stable areas** (≥ 95% confidence) → the mandate says "these are proven, reduce coverage"
- **Retired areas** → not mentioned in the mandate at all

This means the AI naturally generates more scenarios for weak areas and fewer for strong ones — without you configuring anything.

---

## Resetting Memory

If you've made significant changes to a system area and want teststop to re-evaluate it from scratch:

```bash
teststop memory --reset --yes
```

This deletes `memory.json` and `retired.json`. The next run starts from zero confidence on all areas.

!!! warning
    Resetting memory deletes all accumulated confidence proof. Only do this after major refactors or when you want a clean slate. The historical data cannot be recovered after reset.
