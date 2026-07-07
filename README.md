# teststop

> The adversarial AI tester that acts like your most unpredictable user, autonomously finding cracks in your LLM apps and agents before production.

## The Problem

Traditional unit tests, static benchmarks, and grading frameworks fail to catch the chaotic, emergent failures caused by real-world human interactions with AI agents. We usually rely on them for CI/CD grading, but they suffer from "assumption coverage"—they only test the edge cases the developer already knew about. 

teststop solves this by dynamically generating adversarial scenarios custom-tailored to *your* codebase context, and executing them against your running app via HTTP or the CLI. It reads your codebase, figures out what you're trying to do, and then tries to break it from the outside.

## ⚠️ Status: Early Pivot

teststop is currently in an early pivot toward being a general LLM/agent eval tool. It was originally built to test a sibling project, but we've realized the core engine—dynamic, autonomous red-teaming based on your specific application logic—is more generally useful.

We are actively looking for feedback from developers building LLM agents. Please see [docs/PIVOT-SPEC.md](docs/PIVOT-SPEC.md) for our direction and the things we are working on.

## Install

```bash
go install github.com/shaifulshabuj/teststop/cmd/teststop@latest
```

Or build from source:

```bash
git clone https://github.com/shaifulshabuj/teststop
cd teststop
go build -o teststop ./cmd/teststop
```

**Requirements:** [ollama](https://ollama.com) (default, free and private) **or** `claude`/`copilot` CLI on your PATH.

## Quickstart

teststop uses your local codebase to generate scenarios, and maintains its eval cache and config inside a `.teststop/` directory at your project root. 

```bash
# Run on the current directory
teststop run

# Run on a specific path
teststop run --path ./src

# Execute scenarios against a RUNNING system
teststop run --target http://localhost:8080

# Show the current confidence state of the project
teststop status
```

*Note: Because this dynamically generates and executes adversarial scenarios, runs can be slow compared to static assertions.*

## CI/CD and Automation

teststop is built to act as a quality gate in your pipeline. It uses frozen exit codes and provides structured machine-readable output.

```bash
teststop run --target http://localhost:8080 --output json
```

### Exit Codes

| Code | Meaning | Agent Action |
|------|---------|-------------|
| `0` | Confidence threshold met — all scenarios passed. | Safe to proceed / deploy |
| `1` | Below threshold — non-critical failures found. | Human review required |
| `2` | Critical failures — at least one `critical` scenario failed. | **Do NOT deploy** |
| `3` | teststop internal error (e.g. CLI or I/O failure). | Debug teststop |

The JSON output schema produced by `--output json` is strictly versioned and safe for agent consumption.



*teststop v1.1.0*

## License

[MIT](LICENSE)
