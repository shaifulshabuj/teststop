package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/shaifulshabuj/teststop/internal/sandbox"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// CopilotCLI implements AIAdapter by shelling out to the `copilot` CLI.
type CopilotCLI struct {
	runner *sandbox.Runner
}

// NewCopilotCLI creates a CopilotCLI with auto sandbox detection.
func NewCopilotCLI() *CopilotCLI {
	return &CopilotCLI{runner: sandbox.New(sandbox.ModeFromEnv())}
}

func (c *CopilotCLI) Name() string { return "copilot" }

// Prompt sends an arbitrary instruction to the copilot CLI and returns raw stdout.
func (c *CopilotCLI) Prompt(input string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result := c.runner.Run(ctx, sandbox.RunConfig{}, "copilot", "-p", input, "-s", "--no-ask-user")
	if result.Err != nil {
		return nil, fmt.Errorf("copilot: %w\nstderr: %s", result.Err, result.Stderr)
	}
	return result.Stdout, nil
}

// GenerateScenarios sends the mandate to the copilot CLI and parses the returned JSON.
func (c *CopilotCLI) GenerateScenarios(mandate string) ([]scenario.Scenario, error) {
	out, err := c.Prompt(mandate)
	if err != nil {
		return nil, err
	}
	return ParseScenariosFromJSON(out)
}
