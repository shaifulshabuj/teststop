# Quickstart

Get your first adversarial test report in under five minutes.

---

## Step 1 — Point teststop at your project

Navigate to any software project:

```bash
cd /path/to/your/project
```

Or stay in your home directory and pass the path:

```bash
teststop run --path /path/to/your/project
```

teststop works on any project in any language. No setup needed.

---

## Step 2 — Run

```bash
teststop run
```

teststop will:

1. Scan your project (language, type, routes, dependencies)
2. Load any existing confidence memory
3. Compose the adversarial mandate
4. Call your AI CLI (`claude` or `copilot`)
5. Display the test scenarios and confidence report

The first run takes 20–60 seconds depending on your AI CLI.

---

## Step 3 — Read the output

A typical run looks like:

```
teststop v0.1.0 — adversarial user testing

Project:   my-api (Go · api)
Adapter:   claude
Depth:     normal
Scenarios: 9

SCENARIO  TS-001 · critical
  Title:   Double-submit on slow network
  User:    Someone with a poor connection clicking Submit twice
  Steps:
    1. Fill checkout form
    2. Click submit (simulate 3s latency)
    3. Click submit again before response
  Failure modes: Duplicate order created, inventory decremented twice

SCENARIO  TS-002 · high
  Title:   Session cookie after password change
  ...

──────────────────────────────────────────────────────────
Confidence:  0%  (0 areas tested — first run)
Areas:       volatile: auth, checkout, api
Exit code:   1 (REVIEW)
──────────────────────────────────────────────────────────
```

!!! tip "Exit 1 on the first run is normal"
    The first run always exits `1` (review required) because confidence starts at 0%.
    As you run teststop repeatedly and scenarios keep passing, confidence builds and
    eventually exits `0`.

---

## Step 4 — Run again

```bash
teststop run
```

teststop remembers what it tested. Confidence grows with each clean run.
After ~15 passes per area, that area is considered proven and retired.

Check your accumulated confidence state:

```bash
teststop status
```

```
Area         Confidence   Tests   Maturity   Status
──────────────────────────────────────────────────
auth         62%          5       growing    active
checkout     38%          2       new        active
api          19%          1       new        active
```

---

## Step 5 — Commit your memory

```bash
git add .teststop/memory.json
git commit -m "chore: add teststop confidence memory"
```

`.teststop/memory.json` is the accumulated proof that your system works.
Commit it so your team shares the same confidence baseline.

---

## Common flags

```bash
# Machine-readable output for CI / AI agents
teststop run --output json --no-color --quiet

# More thorough testing
teststop run --depth aggressive

# Custom confidence threshold (default 80%)
teststop run --threshold 90

# Show the exact mandate sent to the AI
teststop mandate --show

# See full command reference
teststop --help
```

---

## Next Steps

<div class="grid cards" markdown>

-   :material-brain:{ .lg .middle } **Understand the pipeline**

    ---

    Learn how teststop scans, composes, and generates scenarios end to end.

    [:octicons-arrow-right-24: How It Works](../guide/how-it-works.md)

-   :material-memory:{ .lg .middle } **Memory System**

    ---

    How confidence scores work, how areas retire, and what to commit.

    [:octicons-arrow-right-24: Memory System](../guide/memory.md)

-   :material-robot:{ .lg .middle } **Agent Integration**

    ---

    Wire teststop into Claude Code, Copilot, or any AI coding workflow.

    [:octicons-arrow-right-24: Agent Integration](../advanced/agent-integration.md)

-   :material-console:{ .lg .middle } **CLI Reference**

    ---

    Every command, flag, and option.

    [:octicons-arrow-right-24: CLI Reference](../reference/cli.md)

</div>
