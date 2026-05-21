# Agent Sync Report — 2026-05-21 10:24

**Providers refreshed:** claude copilot

## DevLoop CLI Provider Update Report — 2026-05-21

### 1. New CLI Features / Flags for Non-Interactive Usage

**GitHub Copilot CLI** (formerly the retired `gh copilot` extension, now a standalone CLI):
- **Parallel task execution ("Fleet")** — multiple agent tasks can run concurrently; relevant for batch pipeline use.
- **Autonomous task completion ("Autopilot")** — headless/non-interactive mode explicitly supported.
- **Session data ("Chronicle")** — session state is persisted and queryable, enabling resumable pipelines.
- **LSP server integration** — language servers can be attached for richer context in non-interactive runs.
- **Programmatic/Actions automation** — dedicated `automate-copilot-cli` section with `run-cli-programmatically` and `automate-with-actions` guides.
- **Remote steering** — sessions can be steered via API while running, useful for CI orchestration.

**Claude Code**: The fetched page returned only minified JavaScript (Mintlify SPA shell); no parseable CLI reference content was extractable from the 80-line window.

### 2. Breaking Changes

- **`gh copilot` extension is retired.** Any DevLoop invocation using `gh copilot suggest` or `gh copilot explain` will fail. Must migrate to the new standalone **Copilot CLI** binary.

### 3. Best Practices for Large Prompts / Spec Files

- Use the **`--context` / context management** feature (now documented) to attach files rather than piping raw text into stdin.
- For spec-heavy tasks, leverage **Copilot Spaces** or **custom instructions files** stored in the config directory rather than inline prompt strings.

### 4. Recommended DevLoop Improvements

- **Update Copilot invocation path** from `gh copilot` to the new standalone CLI binary immediately — the old extension is retired and will break.
- **Adopt Chronicle (session data)** for long-running DevLoop tasks so interrupted runs can resume rather than restart.
- **Claude Code doc re-fetch needed** — the current fetch only returned client-side JavaScript; schedule a targeted fetch of `/docs/en/cli-reference` or equivalent to capture actual flag changes.
