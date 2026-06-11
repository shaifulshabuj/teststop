package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/internal/config"
)

// writeConfig writes content to projectPath/.teststop/config.yaml and returns
// projectPath.
func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".teststop")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, config.FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return dir
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string // "" means: do not create the file (absent case)
		create  bool
		wantErr bool
		check   func(t *testing.T, c *config.Config)
	}{
		{
			name:   "file absent — no error, empty config",
			create: false,
			check: func(t *testing.T, c *config.Config) {
				if c.Depth != nil || c.Threshold != nil || c.Target != nil {
					t.Errorf("absent file should yield all-nil config, got %+v", c)
				}
			},
		},
		{
			name:    "empty file — no error, empty config",
			create:  true,
			content: "",
			check: func(t *testing.T, c *config.Config) {
				if c.Depth != nil {
					t.Errorf("empty file should yield all-nil config, got %+v", c)
				}
			},
		},
		{
			name:    "comment-only file — no error, empty config",
			create:  true,
			content: "# just a comment\n",
			check: func(t *testing.T, c *config.Config) {
				if c.Depth != nil {
					t.Errorf("comment-only file should yield all-nil config, got %+v", c)
				}
			},
		},
		{
			name:   "fully populated file parses every key",
			create: true,
			content: `depth: aggressive
output: json
threshold: 90
no_color: true
quiet: true
target: http://localhost:8080
concurrency: 8
exec_timeout: 15s
max_retries: 3
`,
			check: func(t *testing.T, c *config.Config) {
				assertStr(t, "depth", c.Depth, "aggressive")
				assertStr(t, "output", c.Output, "json")
				assertInt(t, "threshold", c.Threshold, 90)
				assertBool(t, "no_color", c.NoColor, true)
				assertBool(t, "quiet", c.Quiet, true)
				assertStr(t, "target", c.Target, "http://localhost:8080")
				assertInt(t, "concurrency", c.Concurrency, 8)
				if c.ExecTimeout == nil || *c.ExecTimeout != 15*time.Second {
					t.Errorf("exec_timeout: want 15s, got %v", c.ExecTimeout)
				}
				assertInt(t, "max_retries", c.MaxRetries, 3)
			},
		},
		{
			name:    "partial file leaves unset keys nil",
			create:  true,
			content: "threshold: 70\n",
			check: func(t *testing.T, c *config.Config) {
				assertInt(t, "threshold", c.Threshold, 70)
				if c.Depth != nil || c.Target != nil {
					t.Errorf("unset keys should stay nil, got %+v", c)
				}
			},
		},
		{
			name:    "malformed yaml — clear error",
			create:  true,
			content: "depth: [unterminated\n",
			wantErr: true,
		},
		{
			name:    "unknown key — rejected loudly",
			create:  true,
			content: "not_a_real_key: 1\n",
			wantErr: true,
		},
		{
			name:    "wrong type for threshold — error",
			create:  true,
			content: "threshold: not-a-number\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var projectPath string
			if tt.create {
				projectPath = writeConfig(t, tt.content)
			} else {
				projectPath = t.TempDir() // no .teststop/config.yaml inside
			}

			c, err := config.Load(projectPath)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

func assertStr(t *testing.T, key string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s: want %q, got %v", key, want, got)
	}
}

func assertInt(t *testing.T, key string, got *int, want int) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s: want %d, got %v", key, want, got)
	}
}

func assertBool(t *testing.T, key string, got *bool, want bool) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s: want %t, got %v", key, want, got)
	}
}
