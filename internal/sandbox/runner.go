package sandbox

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Runner executes commands either in a container or directly.
type Runner struct {
	mode  Mode
	avail bool // whether container system is available
}

// New creates a Runner with the given mode, detecting container availability.
func New(mode Mode) *Runner {
	return &Runner{
		mode:  mode,
		avail: Detect(),
	}
}

// Run executes cmd with args in a container (if available and mode allows) or directly.
func (r *Runner) Run(ctx context.Context, cfg RunConfig, cmd string, args ...string) Result {
	if r.shouldUseContainer() {
		return r.runInContainer(ctx, cfg, cmd, args...)
	}
	return r.runDirect(ctx, cmd, args...)
}

// shouldUseContainer returns true if we should run in a container.
func (r *Runner) shouldUseContainer() bool {
	switch r.mode {
	case ModeDisabled:
		return false
	case ModeRequired:
		return true // will error in runInContainer if not avail
	default: // ModeAuto
		return r.avail
	}
}

// runDirect executes cmd directly via exec.CommandContext.
//
// The working directory is set to the system temp dir, NOT teststop's own cwd.
// Otherwise a spawned `claude`/`copilot` would inherit teststop's directory and
// load the target project's CLAUDE.md and MCP configuration — contaminating the
// AI's behavior (and potentially failing when those MCP servers are unavailable
// to a subprocess). Credentials live under $HOME, not cwd, so a neutral cwd is
// safe; the mandate/scenario prompt is fully self-contained.
func (r *Runner) runDirect(ctx context.Context, cmd string, args ...string) Result {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Dir = os.TempDir()
	stdout, err := c.Output()
	var stderr []byte
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	return Result{Stdout: stdout, Stderr: stderr, Err: err}
}

// runInContainer executes cmd inside an Apple Container VM.
func (r *Runner) runInContainer(ctx context.Context, cfg RunConfig, cmd string, args ...string) Result {
	if !r.avail {
		return Result{Err: fmt.Errorf("sandbox: container system not available (TESTSTOP_SANDBOX=required)")}
	}

	if cfg.Image == "" {
		cfg.Image = DefaultImage
	}

	// Generate unique container name using crypto/rand.
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return Result{Err: fmt.Errorf("sandbox: failed to generate container name: %w", err)}
	}
	name := "teststop-" + hex.EncodeToString(b)

	// Build: container run --rm --name <name> [mounts] [envs] <image> <cmd> <args...>
	containerArgs := []string{"run", "--rm", "--name", name}

	// Add default credential mounts.
	containerArgs = append(containerArgs, defaultMounts()...)

	// Add user-specified mounts.
	for _, m := range cfg.Mounts {
		containerArgs = append(containerArgs, "--volume", m)
	}

	// Add env vars.
	for _, e := range cfg.Env {
		containerArgs = append(containerArgs, "--env", e)
	}

	containerArgs = append(containerArgs, cfg.Image)
	containerArgs = append(containerArgs, cmd)
	containerArgs = append(containerArgs, args...)

	c := exec.CommandContext(ctx, "container", containerArgs...)
	stdout, err := c.Output()
	var stderr []byte
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	return Result{Stdout: stdout, Stderr: stderr, Err: err}
}

// defaultMounts returns the standard credential volume mounts.
func defaultMounts() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		"--volume", filepath.Join(home, ".claude") + ":/root/.claude:ro",
		"--volume", filepath.Join(home, ".config", "gh") + ":/root/.config/gh:ro",
	}
}
