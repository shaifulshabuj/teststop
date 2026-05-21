package sandbox_test

import (
	"context"
	"testing"

	"github.com/shaifulshabuj/teststop/internal/sandbox"
)

func TestModeFromEnv_defaults(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "")
	if sandbox.ModeFromEnv() != sandbox.ModeAuto {
		t.Error("empty TESTSTOP_SANDBOX should be ModeAuto")
	}
}

func TestModeFromEnv_none(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "none")
	if sandbox.ModeFromEnv() != sandbox.ModeDisabled {
		t.Error("TESTSTOP_SANDBOX=none should be ModeDisabled")
	}
}

func TestModeFromEnv_required(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "required")
	if sandbox.ModeFromEnv() != sandbox.ModeRequired {
		t.Error("TESTSTOP_SANDBOX=required should be ModeRequired")
	}
}

func TestModeFromEnv_auto(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "auto")
	if sandbox.ModeFromEnv() != sandbox.ModeAuto {
		t.Error("TESTSTOP_SANDBOX=auto should be ModeAuto")
	}
}

func TestRunner_runDirect(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "none")
	r := sandbox.New(sandbox.ModeDisabled)
	result := r.Run(context.Background(), sandbox.RunConfig{}, "echo", "hello")
	if result.Err != nil {
		t.Fatalf("echo should succeed: %v", result.Err)
	}
	if string(result.Stdout) != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", result.Stdout)
	}
}

func TestRunner_directFallback_noContainer(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "none")
	r := sandbox.New(sandbox.ModeDisabled)
	// Should run 'echo' directly without container.
	result := r.Run(context.Background(), sandbox.RunConfig{}, "echo", "test")
	if result.Err != nil {
		t.Fatal(result.Err)
	}
}

func TestRunner_commandNotFound(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "none")
	r := sandbox.New(sandbox.ModeDisabled)
	result := r.Run(context.Background(), sandbox.RunConfig{}, "this-command-does-not-exist-teststop")
	if result.Err == nil {
		t.Error("expected error for missing command")
	}
}

func TestRunner_stderrCaptured(t *testing.T) {
	t.Setenv("TESTSTOP_SANDBOX", "none")
	r := sandbox.New(sandbox.ModeDisabled)
	// 'ls' on a non-existent path writes to stderr and exits non-zero.
	result := r.Run(context.Background(), sandbox.RunConfig{}, "ls", "/this/path/definitely/does/not/exist/teststop")
	if result.Err == nil {
		t.Skip("ls did not return an error (unexpected OS behaviour)")
	}
	// Stderr should be non-empty.
	if len(result.Stderr) == 0 {
		t.Error("expected stderr output for failed ls")
	}
}
