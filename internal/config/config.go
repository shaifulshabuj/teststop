// Package config loads optional project configuration from
// .teststop/config.yaml. The file is entirely optional: a missing file yields
// an empty config and no error. Every key maps one-to-one onto an existing
// `teststop run` flag — config introduces no new settings, it only lets the
// existing flag defaults be set per-project.
//
// Settings resolve with this precedence (lowest to highest):
//
//	config file  <  environment variable  <  explicit CLI flag
//
// The config package owns only the file tier. The env and flag tiers are
// applied by the caller (see internal/cli/run.go) so the precedence stays
// visible at the call site.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FileName is the project-relative path of the optional config file.
const FileName = "config.yaml"

// Config mirrors the `teststop run` flags. Every field is a pointer so we can
// tell "key absent from file" (nil) apart from "key explicitly set to the zero
// value" (e.g. threshold: 0). Only non-nil fields participate in precedence.
type Config struct {
	// Note: there is deliberately no `path` key. The project path determines
	// where config.yaml is read from, so making the file relocate that path
	// would be circular. Path stays a flag/arg-only setting.
	Depth       *string        `yaml:"depth"`
	Output      *string        `yaml:"output"`
	Threshold   *int           `yaml:"threshold"`
	NoColor     *bool          `yaml:"no_color"`
	Quiet       *bool          `yaml:"quiet"`
	Target      *string        `yaml:"target"`
	Concurrency    *int           `yaml:"concurrency"`
	AIConcurrency  *int           `yaml:"ai_concurrency"`
	ExecTimeout    *time.Duration `yaml:"exec_timeout"`
	MaxRetries     *int           `yaml:"max_retries"`
}

// configPath returns the path to config.yaml for the given project.
func configPath(projectPath string) string {
	return filepath.Join(projectPath, ".teststop", FileName)
}

// Load reads .teststop/config.yaml from projectPath. A missing file is not an
// error — it returns an empty (all-nil) Config. Malformed YAML returns an
// error with the offending path for context.
func Load(projectPath string) (*Config, error) {
	data, err := os.ReadFile(configPath(projectPath))
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var c Config
	// KnownFields rejects typos / unsupported keys so a misspelled setting
	// fails loudly instead of being silently ignored.
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		// An empty (or whitespace/comment-only) file decodes to io.EOF; treat
		// that as an empty config rather than an error.
		if errors.Is(err, io.EOF) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("parsing config file %s: %w", configPath(projectPath), err)
	}
	return &c, nil
}
