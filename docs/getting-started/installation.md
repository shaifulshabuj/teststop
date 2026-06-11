# Installation

teststop ships as a single, zero-dependency binary. No runtime, no daemon. Configuration is optional — `teststop run` works with zero setup on any project.

## Prerequisites

teststop shells out to an AI CLI — it does **not** use an API SDK or require API keys.

You need **one** of the following on your `PATH`:

| CLI | Install | Auth |
|-----|---------|------|
| `claude` | [claude.ai/download](https://claude.ai/download) | Claude account (free tier works) |
| `copilot` (GitHub) | [GitHub Copilot CLI](https://docs.github.com/en/copilot/github-copilot-in-the-cli) | GitHub Copilot subscription |

Verify your AI CLI works:

```bash
echo "Hello" | claude -p "Respond in one word"
```

---

## Install Methods

=== ":material-language-go: Go Install (Recommended)"

    If you have Go 1.21+ installed:

    ```bash
    go install github.com/shaifulshabuj/teststop/cmd/teststop@latest
    ```

    The binary lands in `$(go env GOPATH)/bin`. Make sure that's on your `PATH`:

    ```bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

=== ":material-download: Binary Release"

    Download a pre-compiled binary from the [latest release](https://github.com/shaifulshabuj/teststop/releases/latest).

    === "macOS (Apple Silicon)"

        ```bash
        curl -L https://github.com/shaifulshabuj/teststop/releases/latest/download/teststop_Darwin_arm64.tar.gz \
          | tar xz teststop
        sudo mv teststop /usr/local/bin/
        ```

    === "macOS (Intel)"

        ```bash
        curl -L https://github.com/shaifulshabuj/teststop/releases/latest/download/teststop_Darwin_x86_64.tar.gz \
          | tar xz teststop
        sudo mv teststop /usr/local/bin/
        ```

    === "Linux (amd64)"

        ```bash
        curl -L https://github.com/shaifulshabuj/teststop/releases/latest/download/teststop_Linux_x86_64.tar.gz \
          | tar xz teststop
        sudo mv teststop /usr/local/bin/
        ```

    === "Linux (arm64)"

        ```bash
        curl -L https://github.com/shaifulshabuj/teststop/releases/latest/download/teststop_Linux_arm64.tar.gz \
          | tar xz teststop
        sudo mv teststop /usr/local/bin/
        ```

=== ":material-source-branch: Build from Source"

    ```bash
    git clone https://github.com/shaifulshabuj/teststop
    cd teststop
    go build -o teststop ./cmd/teststop
    sudo mv teststop /usr/local/bin/
    ```

    **Requirements:** Go 1.21+

---

## Verify Installation

```bash
teststop --version
```

```
teststop v0.3.1
```

---

## Sandbox (Optional — macOS only)

teststop can run the AI inside an isolated Apple Container VM for extra security.
This is **optional** — without it, teststop runs the AI CLI directly.

```bash
brew install container
container system start
```

When the container daemon is running, teststop automatically uses it.
To disable sandbox regardless: `TESTSTOP_SANDBOX=none teststop run`

See [Sandbox Isolation](../guide/sandbox.md) for details.

---

## Next Step

[:octicons-arrow-right-24: Run your first test](quickstart.md)
