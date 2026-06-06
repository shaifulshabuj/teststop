package ai

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shaifulshabuj/teststop/internal/sandbox"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// ClaudeCLI implements AIAdapter by shelling out to the `claude` CLI.
type ClaudeCLI struct {
	runner *sandbox.Runner
}

// NewClaudeCLI creates a ClaudeCLI with auto sandbox detection.
func NewClaudeCLI() *ClaudeCLI {
	return &ClaudeCLI{runner: sandbox.New(sandbox.ModeFromEnv())}
}

func (c *ClaudeCLI) Name() string { return "claude" }

// Prompt sends an arbitrary instruction to the claude CLI and returns raw stdout.
func (c *ClaudeCLI) Prompt(input string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{"-p", input}

	// Optional model override.
	if model := os.Getenv("TESTSTOP_MODEL"); model != "" {
		args = append(args, "--model", model)
	}

	result := c.runner.Run(ctx, sandbox.RunConfig{}, "claude", args...)
	if result.Err != nil {
		return nil, fmt.Errorf("claude: %w\nstderr: %s", result.Err, result.Stderr)
	}
	return result.Stdout, nil
}

// GenerateScenarios sends the mandate to the claude CLI and parses the returned JSON.
func (c *ClaudeCLI) GenerateScenarios(mandate string) ([]scenario.Scenario, error) {
	out, err := c.Prompt(mandate)
	if err != nil {
		return nil, err
	}
	return ParseScenariosFromJSON(out)
}
