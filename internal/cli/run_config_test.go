package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// newRunCmdForTest builds a command wired to the same package-level flag vars
// and defaults as the real runCmd, but isolated per test so we can exercise
// resolveRunSettings without invoking the full pipeline.
func newRunCmdForTest() *cobra.Command {
	// Reset the package-level flag vars to their documented defaults so each
	// test starts from a known baseline (these globals are shared with runCmd).
	runDepth = "normal"
	runOutput = "text"
	runThreshold = 80
	runNoColor = false
	runQuiet = false
	runTarget = ""
	runConcurrency = 4
	runAIConcurrency = 1
	runExecTimeout = 10 * time.Second
	runMaxRetries = 2

	c := &cobra.Command{Use: "run"}
	c.Flags().StringVar(&runDepth, "depth", runDepth, "")
	c.Flags().StringVar(&runOutput, "output", runOutput, "")
	c.Flags().IntVar(&runThreshold, "threshold", runThreshold, "")
	c.Flags().BoolVar(&runNoColor, "no-color", runNoColor, "")
	c.Flags().BoolVar(&runQuiet, "quiet", runQuiet, "")
	c.Flags().StringVar(&runTarget, "target", runTarget, "")
	c.Flags().IntVar(&runConcurrency, "concurrency", runConcurrency, "")
	c.Flags().IntVar(&runAIConcurrency, "ai-concurrency", runAIConcurrency, "")
	c.Flags().DurationVar(&runExecTimeout, "exec-timeout", runExecTimeout, "")
	c.Flags().IntVar(&runMaxRetries, "max-retries", runMaxRetries, "")
	return c
}

func writeProjectConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if content != "" {
		cfgDir := filepath.Join(dir, ".teststop")
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(content), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}
	}
	return dir
}

func TestResolveRunSettings_precedence(t *testing.T) {
	const fileContent = `depth: aggressive
threshold: 90
target: http://from-file
concurrency: 8
`

	t.Run("file only — file values applied over defaults", func(t *testing.T) {
		clearRunEnv(t)
		dir := writeProjectConfig(t, fileContent)
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if runDepth != "aggressive" {
			t.Errorf("depth: want file value 'aggressive', got %q", runDepth)
		}
		if runThreshold != 90 {
			t.Errorf("threshold: want 90, got %d", runThreshold)
		}
		// A key absent from the file keeps its built-in default.
		if runMaxRetries != 2 {
			t.Errorf("max_retries: want default 2, got %d", runMaxRetries)
		}
	})

	t.Run("absent file — defaults preserved, no error", func(t *testing.T) {
		clearRunEnv(t)
		dir := writeProjectConfig(t, "") // no config.yaml
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatalf("resolve should not error on absent file: %v", err)
		}
		if runDepth != "normal" || runThreshold != 80 {
			t.Errorf("absent file should keep defaults, got depth=%q threshold=%d", runDepth, runThreshold)
		}
	})

	t.Run("env overrides file", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_THRESHOLD", "55")
		t.Setenv("TESTSTOP_RUN_DEPTH", "light")
		dir := writeProjectConfig(t, fileContent)
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if runThreshold != 55 {
			t.Errorf("threshold: env should win over file, want 55, got %d", runThreshold)
		}
		if runDepth != "light" {
			t.Errorf("depth: env should win over file, want 'light', got %q", runDepth)
		}
		// Not overridden by env → falls back to file value.
		if runTarget != "http://from-file" {
			t.Errorf("target: want file value, got %q", runTarget)
		}
	})

	t.Run("flag overrides both file and env", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_THRESHOLD", "55")
		dir := writeProjectConfig(t, fileContent)
		cmd := newRunCmdForTest()
		// Explicit CLI flag — highest precedence.
		if err := cmd.ParseFlags([]string{"--threshold", "33", "--depth", "normal"}); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if runThreshold != 33 {
			t.Errorf("threshold: explicit flag should win over env+file, want 33, got %d", runThreshold)
		}
		if runDepth != "normal" {
			t.Errorf("depth: explicit flag should win over file, want 'normal', got %q", runDepth)
		}
	})

	t.Run("malformed env value errors", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_THRESHOLD", "not-an-int")
		dir := writeProjectConfig(t, "")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err == nil {
			t.Fatal("expected error for malformed TESTSTOP_RUN_THRESHOLD")
		}
	})

	t.Run("file sets output, no-color, quiet, exec-timeout, max-retries", func(t *testing.T) {
		clearRunEnv(t)
		dir := writeProjectConfig(t, "output: json\nno_color: true\nquiet: true\nexec_timeout: 5s\nmax_retries: 3\n")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if runOutput != "json" {
			t.Errorf("output: want %q, got %q", "json", runOutput)
		}
		if !runNoColor {
			t.Error("no_color: want true")
		}
		if !runQuiet {
			t.Error("quiet: want true")
		}
		if runExecTimeout != 5*time.Second {
			t.Errorf("exec_timeout: want 5s, got %v", runExecTimeout)
		}
		if runMaxRetries != 3 {
			t.Errorf("max_retries: want 3, got %d", runMaxRetries)
		}
	})

	t.Run("malformed config file errors", func(t *testing.T) {
		clearRunEnv(t)
		dir := writeProjectConfig(t, "depth: [unterminated\n")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err == nil {
			t.Fatal("expected error for malformed config.yaml")
		}
	})
}

// clearRunEnv unsets every TESTSTOP_RUN_* var for the duration of the test so
// the host environment can't leak into precedence assertions.
func clearRunEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"TESTSTOP_RUN_DEPTH",
		"TESTSTOP_RUN_OUTPUT",
		"TESTSTOP_RUN_THRESHOLD",
		"TESTSTOP_RUN_NO_COLOR",
		"TESTSTOP_RUN_QUIET",
		"TESTSTOP_RUN_TARGET",
		"TESTSTOP_RUN_CONCURRENCY",
		"TESTSTOP_RUN_AI_CONCURRENCY",
		"TESTSTOP_RUN_EXEC_TIMEOUT",
		"TESTSTOP_RUN_MAX_RETRIES",
	} {
		t.Setenv(k, "") // ensures present-but-empty is removed below
		os.Unsetenv(k)
	}
}

// TestResolveRunSettings_aiConcurrency covers the three-tier precedence for the
// new --ai-concurrency / ai_concurrency / TESTSTOP_RUN_AI_CONCURRENCY setting.
func TestResolveRunSettings_aiConcurrency(t *testing.T) {
	t.Run("default is 1", func(t *testing.T) {
		clearRunEnv(t)
		dir := writeProjectConfig(t, "")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatal(err)
		}
		if runAIConcurrency != 1 {
			t.Errorf("default ai-concurrency: want 1, got %d", runAIConcurrency)
		}
	})

	t.Run("config file sets value", func(t *testing.T) {
		clearRunEnv(t)
		dir := writeProjectConfig(t, "ai_concurrency: 3\n")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatal(err)
		}
		if runAIConcurrency != 3 {
			t.Errorf("config file: want 3, got %d", runAIConcurrency)
		}
	})

	t.Run("env overrides config file", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_AI_CONCURRENCY", "5")
		dir := writeProjectConfig(t, "ai_concurrency: 3\n")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatal(err)
		}
		if runAIConcurrency != 5 {
			t.Errorf("env: want 5, got %d", runAIConcurrency)
		}
	})

	t.Run("explicit flag overrides env and config", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_AI_CONCURRENCY", "5")
		dir := writeProjectConfig(t, "ai_concurrency: 3\n")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags([]string{"--ai-concurrency", "2"}); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err != nil {
			t.Fatal(err)
		}
		if runAIConcurrency != 2 {
			t.Errorf("explicit flag: want 2, got %d", runAIConcurrency)
		}
	})

	t.Run("malformed env errors", func(t *testing.T) {
		clearRunEnv(t)
		t.Setenv("TESTSTOP_RUN_AI_CONCURRENCY", "notanint")
		dir := writeProjectConfig(t, "")
		cmd := newRunCmdForTest()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatal(err)
		}
		if err := resolveRunSettings(cmd, dir); err == nil {
			t.Fatal("expected error for malformed TESTSTOP_RUN_AI_CONCURRENCY")
		}
	})
}
