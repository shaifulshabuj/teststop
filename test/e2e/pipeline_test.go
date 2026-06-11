// Package e2e exercises the full teststop run pipeline using a fake AI CLI
// fixture instead of real tokens. Skippable via -short.
package e2e_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeCLIScript is a POSIX shell script that acts as a fake "claude" binary.
// It ignores all arguments and emits a single valid scenario inside the
// JSON envelope that ClaudeCLI expects from `claude --output-format json`.
const fakeCLIScript = `#!/bin/sh
cat <<'ENDJSON'
{"is_error":false,"result":"[{\"scenario_id\":\"sc-e2e-001\",\"title\":\"Fake E2E scenario\",\"user_perspective\":\"attacker probing auth boundaries\",\"preconditions\":[],\"steps\":[\"attempt unauthorized access\"],\"chaos_factors\":[\"bad input\",\"slow network\"],\"expected_behavior\":\"system rejects with 403\",\"failure_modes\":[\"panic\",\"data leak\"],\"priority\":\"medium\",\"confidence_area\":\"e2e-test-area\",\"is_edge_case\":false}]"}
ENDJSON
`

// moduleRoot walks up from the test's working directory to find the go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod walking up from " + dir)
		}
		dir = parent
	}
}

// buildTeststop compiles the teststop binary into a temp dir and returns the path.
func buildTeststop(t *testing.T, root string) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "teststop")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/teststop")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build teststop: %v\n%s", err, out)
	}
	return binPath
}

// filterEnv returns a copy of env with all TESTSTOP_* entries removed so the
// host environment cannot leak into the subprocess and override test settings.
func filterEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "TESTSTOP_") {
			out = append(out, e)
		}
	}
	return out
}

// TestPipelineE2E_FakeAI exercises the full teststop run pipeline:
//
//	reader → mandate → adapter → memory update → reporter → exit code
//
// It uses a fake claude CLI script that emits a valid scenario JSON envelope
// without making real AI calls.
func TestPipelineE2E_FakeAI(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e: skipped in short mode (builds binary)")
	}

	root := moduleRoot(t)
	binPath := buildTeststop(t, root)

	// Write the fake claude script into a private bin dir.
	fakeDir := t.TempDir()
	fakeScript := filepath.Join(fakeDir, "claude")
	if err := os.WriteFile(fakeScript, []byte(fakeCLIScript), 0o755); err != nil {
		t.Fatalf("write fake claude: %v", err)
	}

	// Create a minimal Go project for the reader to scan.
	projDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projDir, "main.go"),
		[]byte("package main\nfunc main() {}"), 0o644); err != nil {
		t.Fatalf("create project: %v", err)
	}

	// Run teststop run against the fake project.
	cmd := exec.Command(binPath, "run", "--path", projDir, "--output", "json")
	cmd.Env = append(
		filterEnv(os.Environ()),
		"TESTSTOP_CLI=claude",
		"TESTSTOP_SANDBOX=none",
		"PATH="+fakeDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	out, err := cmd.Output()

	// Expect exit code 1: confidence starts at 0, below the default 80% threshold.
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running teststop: %v", err)
		}
	}
	if exitCode != 1 {
		t.Errorf("exit code: want 1 (below threshold), got %d\nstdout: %s", exitCode, out)
	}

	// Output must be valid JSON.
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output not valid JSON: %v\nraw: %s", err, out)
	}

	// At least one scenario must have been generated and reported.
	scenarios, _ := result["scenarios"].([]any)
	if len(scenarios) == 0 {
		t.Errorf("expected ≥1 scenario in JSON output, got: %v", result["scenarios"])
	}

	// adapter_name must be "claude" (the fake CLI registered under that name).
	if name, _ := result["adapter_name"].(string); name != "claude" {
		t.Errorf("adapter_name: want %q, got %q", "claude", name)
	}

	// Memory file must have been written to the project dir.
	memPath := filepath.Join(projDir, ".teststop", "memory.json")
	if _, err := os.Stat(memPath); err != nil {
		t.Errorf("memory.json not created at %s: %v", memPath, err)
	}
}
