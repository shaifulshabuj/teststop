---
hide:
  - navigation
  - toc
---

<div class="ts-hero" markdown>

<div class="ts-hero-badge">v0.3.1 — Now Available</div>

# **teststop**<br><span class="accent">Break it before your users do.</span>

<p class="ts-hero-subtitle">
Trigger AI to test any software system the way a real adversarial user would —
someone who never read the docs, retries when things are slow, and does
what no spec ever imagined.
</p>

<div class="ts-hero-actions">
<a href="getting-started/installation/" class="ts-btn ts-btn-primary">
  Get Started →
</a>
<a href="guide/how-it-works/" class="ts-btn ts-btn-secondary">
  How It Works
</a>
<a href="https://github.com/shaifulshabuj/teststop" class="ts-btn ts-btn-secondary" target="_blank" rel="noopener">
  GitHub ↗
</a>
</div>

```bash
go install github.com/shaifulshabuj/teststop/cmd/teststop@latest
teststop run
```

</div>

---

## The Problem With Test Coverage

Every test ever written was written by someone who *knew* how the system works.

Real users don't know. That's why production still surprises us.

We call it test coverage. What we actually have is **assumption coverage**.

<div class="ts-features">

<div class="ts-feature" markdown>
<div class="ts-feature-icon">🎯</div>

### Zero Configuration

`teststop run` works on any project with no setup. Point it at any directory — Go, Python, TypeScript, Ruby, anything — and it learns the system by reading it.
</div>

<div class="ts-feature" markdown>
<div class="ts-feature-icon">🤖</div>

### AI-Native, No API Keys

Uses `claude` or `copilot` CLI already on your PATH. No SDK, no secrets to manage, no lock-in. The AI thinks adversarially because the mandate tells it to.
</div>

<div class="ts-feature" markdown>
<div class="ts-feature-icon">📉</div>

### Tests Reduce Over Time

Confidence scores persist per area. Proven stable areas get tested less. New or changed areas get tested more. After ~15 clean passes, an area retires.
</div>

<div class="ts-feature" markdown>
<div class="ts-feature-icon">🔗</div>

### Agent-Native Output

JSON output is designed for AI coding agents. Structured exit codes signal deploy safety. teststop fits cleanly into any automated workflow.
</div>

<div class="ts-feature" markdown>
<div class="ts-feature-icon">🛡️</div>

### Sandbox Isolation

On macOS with Apple Container, teststop runs the AI inside an isolated VM. The AI cannot touch your host filesystem beyond the mounted project path.
</div>

<div class="ts-feature" markdown>
<div class="ts-feature-icon">🌍</div>

### Universal

Works on any language, any age, any system type — web apps, APIs, CLIs, data pipelines, libraries. If it has code, teststop can test it.
</div>

</div>

---

## How It Works in 30 Seconds

<div class="ts-pipeline">

<div class="ts-pipeline-step">
<div class="ts-pipeline-num">1</div>
<div class="ts-pipeline-content">
<h4>Scan</h4>
<p>teststop walks your project tree, detects the language and system type, and extracts routes, flows, and dependencies — all statically, no code execution.</p>
</div>
</div>

<div class="ts-pipeline-step">
<div class="ts-pipeline-num">2</div>
<div class="ts-pipeline-content">
<h4>Compose Mandate</h4>
<p>Injects project context and accumulated memory into the base mandate — the adversarial instruction that tells the AI how a real user would break this specific system.</p>
</div>
</div>

<div class="ts-pipeline-step">
<div class="ts-pipeline-num">3</div>
<div class="ts-pipeline-content">
<h4>Generate Scenarios</h4>
<p>Calls <code>claude -p</code> or <code>copilot -p</code> with the mandate. The AI returns structured JSON: scenario IDs, steps, chaos factors, failure modes, and priorities.</p>
</div>
</div>

<div class="ts-pipeline-step">
<div class="ts-pipeline-num">4</div>
<div class="ts-pipeline-content">
<h4>Update Memory</h4>
<p>Confidence scores update per area. High confidence → less testing next run. Retirement at 0.95+ with ≥15 passes. The system gets smarter with every run.</p>
</div>
</div>

<div class="ts-pipeline-step">
<div class="ts-pipeline-num">5</div>
<div class="ts-pipeline-content">
<h4>Report &amp; Exit</h4>
<p>Outputs JSON, text, or markdown. Exits with a machine-readable code: 0 = safe to deploy, 1 = review needed, 2 = critical failures found.</p>
</div>
</div>

</div>

---

## Quick Install

=== "Go Install"

    ```bash
    go install github.com/shaifulshabuj/teststop/cmd/teststop@latest
    ```

=== "Binary Release"

    Download the latest binary for your platform from
    [GitHub Releases](https://github.com/shaifulshabuj/teststop/releases/latest).

    ```bash
    # macOS arm64 example
    curl -L https://github.com/shaifulshabuj/teststop/releases/latest/download/teststop_Darwin_arm64.tar.gz \
      | tar xz
    sudo mv teststop /usr/local/bin/
    ```

=== "Build from Source"

    ```bash
    git clone https://github.com/shaifulshabuj/teststop
    cd teststop
    go build -o teststop ./cmd/teststop
    ```

**Prerequisite:** `claude` or `copilot` CLI must be on your PATH.

[Full installation guide →](getting-started/installation.md){ .md-button }

---

## Part of a Trilogy

```
DocuFlow  → gives AI the context to act with purpose
Waymark   → gives humans the reason to trust and step back
teststop  → gives systems the confidence to prove themselves
```

| Tool | Role | Link |
|------|------|------|
| [DocuFlow](https://github.com/shaifulshabuj/docuflow-mcp) | MCP server — LLM wiki for AI context | GitHub |
| [Waymark](https://github.com/shaifulshabuj/waymark) | MCP middleware — AI agent governance | GitHub |
| **teststop** | CLI — adversarial testing trigger | **This repo** |
