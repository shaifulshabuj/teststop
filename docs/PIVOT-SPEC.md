# teststop Pivot Spec: General LLM/Agent Eval

## 1. Waymark-coupling audit

The codebase was audited for coupling to the "waymark" project or specific target shapes. Overall, there is no structural or architectural coupling in the core Go code, but documentation and examples are heavily tied to it.

*   **`internal/ai/adapter_test.go:179`**
    *   *Assumption:* Contains a mock string `We are generating exactly 30 scenarios for the waymark API system.`
    *   *Difficulty to generalize:* **Trivial**. Just change the string to a generic demo target.
*   **`.docuflow/` & `teststop-init/` documentation (e.g. `03-PRD.md:231`)**
    *   *Assumption:* Explicit references to "FR8/G8 — Waymark Integration (Optional)" mapping teststop as a governance agent for waymark.
    *   *Difficulty to generalize:* **Trivial**. Update or archive these documents to reflect the pivot.
*   **`CHANGELOG.md`**
    *   *Assumption:* Mentions "quality tradeoff against waymark API project".
    *   *Difficulty to generalize:* **Trivial**.
*   **`examples/waymark-demo/`**
    *   *Assumption:* The entire primary example directory is a captured run against the waymark repo, including `mandate.md` shaped around waymark's API endpoints.
    *   *Difficulty to generalize:* **Moderate**. We must delete this directory and replace it with a generic, relatable LLM application target (e.g., a simple LangChain or OpenAI-based agent).
*   **The `.teststop/` Convention (`internal/config/config.go:52`)**
    *   *Assumption:* teststop expects to drop `.teststop/config.yaml` and `.teststop/memory.json` into the root of the target project. While not strictly "waymark", it assumes a specific operational shape.
    *   *Difficulty to generalize:* **Moderate**. Must ensure users of general frameworks accept this cache/config pattern.

## 2. Positioning

*   **Target user:** LLM application builders and agent framework developers who need to test their systems against unpredictable human-like behaviors, edge cases, and prompt injections before deploying to production.
*   **Problem in one sentence:** Traditional unit tests and static benchmarks fail to catch the chaotic, emergent failures caused by real-world human interactions with AI agents.
*   **One-line positioning statement:** "The adversarial AI tester that acts like your most unpredictable user, autonomously finding cracks in your LLM apps and agents before production."

### Differentiation vs. Incumbents
*   **Promptfoo**: Promptfoo excels at CLI-driven matrix testing with static assertions. *teststop* dynamically generates adversarial scenarios based on your specific codebase context.
*   **DeepEval**: A Pytest-like framework using structured metrics for CI/CD. *teststop* focuses purely on autonomous red-teaming and discovering *unknown* edge cases rather than grading known ones.
*   **Inspect (UK AISI)**: A framework for formal research benchmarking and capability auditing. *teststop* is an agile developer tool meant to break your specific application logic, not benchmark the underlying foundational model.
*   **Braintrust**: An end-to-end observability and dataset management platform. *teststop* acts as the synthetic chaotic user to generate the data/traffic that a platform like Braintrust would observe.
*   **LangSmith Evals**: Deeply integrated observability for LangChain. *teststop* is framework-agnostic and attacks the system externally via HTTP or CLI.
*   **OpenAI Evals**: A registry of public, open-source benchmarks (standardized tasks). *teststop* generates tests custom-tailored to *your* implementation.

## 3. Untether + MVP cut

The MVP focuses on the smallest scope required to make teststop usable by a stranger on their own LLM app.

**What to cut:**
*   Delete `examples/waymark-demo/` entirely.
*   Remove all references to the "Waymark Governance Integration" from public PRDs and documentation.

**What to generalize:**
*   Replace `examples/` with a generic target (e.g., `examples/chatbot-demo`).
*   Update mock strings in `adapter_test.go`.
*   Frame the `.teststop/` directory explicitly as an "eval cache and config" folder, analogous to `.pytest_cache` or `.next`.

**What stays frozen (per `CONTRACTS.md`):**
*   **Scenario Schema (`pkg/scenario/types.go`)**: `scenario_id`, `title`, `steps`, `exec` blocks.
*   **Exit Codes**: 0, 1, 2, 3 remain the CI/CD gating mechanism.
*   **JSON Run Output (`--output json`)**
*   **Memory file format (`.teststop/memory.json`)**
*   **Environment variables** (`TESTSTOP_CLI`, etc.)

**OUT of scope for the MVP:**
*   A graphical dashboard or UI.
*   Integrations with external observability platforms (Braintrust, LangSmith).
*   Any form of runtime agent governance or policy enforcement.

## 4. Demand-check plan

Before building any new features or writing heavy code, we will validate demand for this pivot.

*   **Channels:** Post on Reddit (`r/LocalLLaMA`, `r/LangChain`), Hacker News ("Show HN: Adversarial tester for your AI agents"), and relevant AI developer Discord servers.
*   **Artifact to show:** A short, punchy `README.md` pitch and a 1-minute asciinema terminal recording (or GIF) showing teststop autonomously finding an edge case in a well-known demo application.
*   **Evidence of demand:** At least 50 GitHub stars, 10+ organic CLI installations, or 3+ developers opening genuine issues/PRs for their own use-cases within 14 days of posting.
*   **Kill threshold:** If the "Show HN" post receives < 10 upvotes and there is zero organic adoption or feature requests within two weeks, kill the project and do not build.

## 5. Risks + open questions

*   **Risk:** The `.teststop/` directory convention might be rejected by users who don't want another dotfile cluster in their repo.
*   **Risk:** Autonomous scenario generation can be slow/expensive. Will users bounce if their first run takes 2 minutes and costs $0.05 on Claude, instead of instantly returning like promptfoo?
*   **Open Question (NEEDS-DECISION):** Do we keep the name "teststop"? A general LLM eval tool might benefit from a more descriptive name in a crowded market.
*   **Open Question (NEEDS-DECISION):** Should we provide a way to override the `.teststop/` path globally so users don't have to dirty their repositories?
