# Scenario Schema

The scenario schema is the **stable JSON contract** between teststop and any system that consumes its output. The v0.1 core fields are locked — changes to them are breaking changes. The optional `exec` field was **added in v0.2** as a non-breaking, additive extension: legacy scenarios without it parse unchanged.

---

## Full Schema

```json
{
  "scenario_id": "string",
  "title": "string",
  "user_perspective": "string",
  "preconditions": ["string"],
  "steps": ["string"],
  "chaos_factors": ["string"],
  "expected_behavior": "string",
  "failure_modes": ["string"],
  "priority": "critical | high | medium | low",
  "confidence_area": "string",
  "is_edge_case": true,
  "exec": {
    "mode": "http",
    "method": "POST",
    "path": "/api/login",
    "headers": { "Content-Type": "application/json" },
    "body": "{\"username\":\"\",\"password\":\"\"}",
    "expected_status": 400
  }
}
```

!!! note "`exec` is optional (v0.2)"
    Every field except `exec` is required. The AI emits `exec` **only** when a
    scenario maps cleanly to a single concrete request that can be replayed
    deterministically. See [Execution](../guide/execution.md) for how it is used.

---

## Field Reference

### `scenario_id`

**Type:** `string`

A unique identifier for the scenario within a run. Format: `TS-NNN` (e.g., `TS-001`, `TS-007`).

Use this field to reference specific scenarios in reports, issues, or CI annotations.

---

### `title`

**Type:** `string`

A short, descriptive name for the scenario. Written from the user's perspective, not the developer's.

**Good:** `"Double-submit on slow network"`
**Avoid:** `"Test idempotency of POST /checkout"`

---

### `user_perspective`

**Type:** `string`

A narrative description of who the user is and what they're trying to do. This is the adversarial framing that makes the scenario realistic.

**Example:**
```
"A frustrated mobile user with a slow 3G connection who clicked submit, 
saw nothing happen for 5 seconds, and clicked again"
```

---

### `preconditions`

**Type:** `string[]`

The state that must exist before the user begins. Each entry is one condition.

**Example:**
```json
[
  "User is logged in",
  "Shopping cart contains 3 items",
  "User is on the checkout page"
]
```

---

### `steps`

**Type:** `string[]`

The specific actions the adversarial user takes. Each step is one action.

**Example:**
```json
[
  "Fill in credit card number",
  "Click the Submit button",
  "Wait 4 seconds (simulating network lag)",
  "Click Submit again before the response arrives"
]
```

---

### `chaos_factors`

**Type:** `string[]`

Environmental and behavioral conditions that increase the likelihood of failure.

**Common examples:**
```json
[
  "3G mobile network (2–4 second latency)",
  "Two browser tabs open with the same session",
  "Browser back button used mid-flow",
  "Payment provider API responding slowly",
  "User pasted data from Excel with invisible characters"
]
```

---

### `expected_behavior`

**Type:** `string`

What a correct system should do in this scenario.

**Example:**
```
"The second submit is rejected with a 409 Conflict or ignored by the client. 
Exactly one order is created. The user sees a clear confirmation with order ID."
```

---

### `failure_modes`

**Type:** `string[]`

Specific ways the system could fail in this scenario. Each entry is one failure mode.

**Example:**
```json
[
  "Duplicate order created",
  "Inventory decremented twice",
  "User charged twice",
  "500 Internal Server Error with no user-facing message"
]
```

!!! note "Exit code 2 trigger"
    When a scenario is **executed** against a `--target`, a `critical`-priority
    scenario that **fails** triggers exit code `2` (critical failures). Without a
    `--target`, scenarios are validated structurally and report no execution
    failures. See [Execution](../guide/execution.md).

---

### `priority`

**Type:** `"critical" | "high" | "medium" | "low"`

The severity level of the scenario.

| Priority | Meaning | Action |
|----------|---------|--------|
| `critical` | Could cause data loss, security breach, or revenue impact | Investigate immediately |
| `high` | Likely to affect user experience significantly | Prioritize in next sprint |
| `medium` | Notable but recoverable failure | Address before release |
| `low` | Minor annoyance or edge case | Track in backlog |

---

### `confidence_area`

**Type:** `string`

Which system area this scenario covers. teststop uses this field to update confidence scores.

This should be a stable, human-readable identifier — typically a system capability:
`auth`, `checkout`, `api`, `payments`, `search`, `notifications`, etc.

---

### `is_edge_case`

**Type:** `boolean`

Whether this scenario represents an unusual or boundary condition.

`true` — the scenario tests behavior outside the happy path or at system limits.
`false` — the scenario tests normal or common usage.

---

### `exec`

**Type:** `object | undefined` — **optional, added in v0.2**

Present only when the scenario maps to a single concrete, replayable request.
When present, teststop's [HTTP executor](../guide/execution.md) runs the request
deterministically against `--target`. When absent, the scenario is executed by
the AI driver (if `--target` is set) or validated structurally.

| Sub-field | Type | Description |
|-----------|------|-------------|
| `mode` | `"http" \| "cli"` | Execution mode |
| `method` | `string` | HTTP verb (`GET`, `POST`, …) |
| `path` | `string` | Path appended to the `--target` base URL |
| `headers` | `object` | Request headers (string → string) |
| `body` | `string` | Request body |
| `expected_status` | `int` | Status a correct system returns; `0`/omitted means "any non-5xx passes" |
| `command` | `string[]` | CLI argv (when `mode` is `cli`) |
| `expected_exit` | `int` | Expected process exit code (CLI mode) |

---

## Full Example

```json
{
  "scenario_id": "TS-003",
  "title": "Session expiry mid-checkout",
  "user_perspective": "A user who left their checkout form open overnight and returned the next morning to submit it",
  "preconditions": [
    "User is authenticated with a 24-hour session token",
    "Shopping cart contains items",
    "Checkout form is fully filled in",
    "Session expires while the user is idle on the page"
  ],
  "steps": [
    "Open checkout page in the morning",
    "Fill in all form fields",
    "Leave the page idle for 25 hours",
    "Return to the page",
    "Click Submit without refreshing"
  ],
  "chaos_factors": [
    "JWT session token expired 1 hour ago",
    "Form state persisted in browser memory from yesterday",
    "API returns 401 on submission"
  ],
  "expected_behavior": "User is redirected to login, cart contents are preserved, user can complete checkout after re-authenticating.",
  "failure_modes": [
    "Order submitted without authentication — security bypass",
    "Cart contents lost after login redirect",
    "Generic 401 error with no explanation or recovery path",
    "Infinite redirect loop between login and checkout"
  ],
  "priority": "critical",
  "confidence_area": "auth",
  "is_edge_case": false
}
```

---

## Consuming the Schema

teststop outputs the full scenario array in JSON format:

```bash
teststop run --output json | jq '.scenarios[].priority'
```

```bash
# Filter critical scenarios only
teststop run --output json | jq '[.scenarios[] | select(.priority == "critical")]'
```

The schema is defined in `pkg/scenario/types.go` and is imported by any Go tool that wants to parse teststop output directly.
