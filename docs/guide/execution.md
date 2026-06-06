# Execution

> Added in **v0.2**.

By default, `teststop run` **generates** adversarial scenarios and validates them
structurally. Point it at a running system with `--target` and teststop will also
**execute** those scenarios and feed the real pass/fail outcomes into
[confidence memory](memory.md).

This is the jump from a scenario *generator* to a scenario *runner*.

---

## teststop tests what's running — it doesn't run your app

teststop never starts, builds, or manages the system under test. You run your app
however you like — a local process, a container, a staging deployment — and point
`--target` at it:

```bash
teststop run --target http://localhost:8080
```

This is deliberate. teststop is a thin trigger; managing your app's lifecycle
would make it "the new loop that needs maintaining," which the project explicitly
avoids. The upside: you can target **any** environment, including a
production-like instance, and teststop adapts to it.

---

## Hybrid execution

teststop chooses how to execute each scenario individually:

```mermaid
flowchart TD
    A[Scenario] --> B{exec block<br/>and --target set?}
    B -- yes --> C[HTTP executor<br/>deterministic net/http]
    B -- no --> D{--target set?}
    D -- yes --> E[AI executor<br/>AI performs the steps]
    D -- no --> F[Static executor<br/>structural validation]
```

| Condition | Executor | Behavior |
|-----------|----------|----------|
| Scenario has a structured `exec` block **and** `--target` is set | **HTTP** | Fires the exact request with `net/http`; judged on status code |
| `--target` set, scenario is prose-only | **AI-driven** | The AI actually performs the steps against the target and returns a verdict |
| No `--target` | **Static** | Validates scenario structure only (default; ≈ v0.1 behavior) |

The HTTP path is **deterministic and fast** (no AI call at execution time). The
AI path covers open-ended, chaos-heavy scenarios that can't be reduced to one
request. See the [`exec` field](../reference/scenarios.md#exec) for the structured
contract.

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--target <url>` | _(none)_ | Base URL of the running system. Empty = static only. |
| `--concurrency <n>` | `4` | Max scenarios executed in parallel |
| `--exec-timeout <dur>` | `10s` | Per-request timeout (`15s`, `500ms`, …) |
| `--max-retries <n>` | `2` | Retries for transport errors and `5xx` responses |

Execution runs concurrently with a bounded worker pool; results are returned in
scenario order.

---

## Sandbox and localhost

teststop runs the AI CLI inside an [Apple Container sandbox](sandbox.md) when
available. A sandboxed container **cannot reach your host's `localhost`**, so:

```bash
# Local target → run the AI tester directly
TESTSTOP_SANDBOX=none teststop run --target http://localhost:8080

# Remote / staging target → sandbox works fine
TESTSTOP_SANDBOX=required teststop run --target https://staging.example.com
```

The deterministic **HTTP executor always runs host-side**, so the sandbox setting
does not affect HTTP-mode scenarios.

!!! info "Future work"
    Wiring the executor to run *inside* the sandbox's network (so even local
    execution is fully isolated) is tracked as a follow-up. Today, use
    `TESTSTOP_SANDBOX=none` for local targets.

---

## What execution looks like

Running against a small Go API:

```bash
TESTSTOP_SANDBOX=none teststop run --path ./sample-api \
  --target http://localhost:8099 --depth light --output json
```

```json
{
  "exec_summary": { "executed": 5, "passed": 4, "failed": 1, "target": "http://localhost:8099" },
  "executions": [
    {
      "scenario_id": "login-empty-whitespace-username",
      "area": "api/login",
      "mode": "http",
      "passed": false,
      "actual_behavior": "HTTP 401 in 5ms",
      "failure_reason": "expected status 400, got 401",
      "priority": "medium",
      "duration_ms": 5
    }
  ]
}
```

Here teststop found a real issue: a login with a whitespace-only username was
authenticated (`401`) instead of being rejected as invalid input (`400`). That
failure lowers confidence for the `api/login` area and is reflected in the
[exit code](../reference/exit-codes.md).

---

## Effect on confidence and exit codes

- Each executed scenario updates its `confidence_area`: a **pass** raises
  confidence, a **failure** drops it (see [Memory](memory.md)).
- A failed **`critical`** scenario sets [exit code](../reference/exit-codes.md)
  `2` (do not deploy).
- Below-threshold average confidence sets exit code `1` (review).

Without `--target`, well-formed scenarios pass structural validation, so behavior
matches v0.1.
