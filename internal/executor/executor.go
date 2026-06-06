// Package executor runs generated scenarios against a live system and reports
// real pass/fail outcomes. It is the v0.2 evolution that turns teststop from a
// scenario GENERATOR into a scenario RUNNER.
//
// Execution is hybrid (graceful degradation per scenario):
//
//	exec(http) + live target  -> HTTPExecutor   (deterministic, net/http)
//	live target, no exec      -> AIExecutor     (AI drives the request)
//	no live target            -> StaticExecutor (structural validation, v0.1 semantics)
package executor

import (
	"context"
	"time"

	"github.com/shaifulshabuj/teststop/internal/ai"
	"github.com/shaifulshabuj/teststop/internal/sandbox"
	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// Execution mode labels (recorded on ExecutionResult.Mode).
const (
	ModeHTTP   = "http"
	ModeAI     = "ai"
	ModeStatic = "static"
)

// Defaults applied by Config.withDefaults when fields are zero.
const (
	DefaultTimeout     = 10 * time.Second
	DefaultMaxRetries  = 2
	DefaultConcurrency = 4
)

// ExecutionResult is the outcome of executing a single scenario.
type ExecutionResult struct {
	ScenarioID     string        `json:"scenario_id"`
	Area           string        `json:"area"`
	Mode           string        `json:"mode"` // http | ai | static
	Passed         bool          `json:"passed"`
	ActualBehavior string        `json:"actual_behavior"`
	FailureReason  string        `json:"failure_reason,omitempty"`
	Priority       string        `json:"priority"`
	Duration       time.Duration `json:"duration_ms"`
}

// Executor executes one scenario and reports its outcome.
type Executor interface {
	Execute(ctx context.Context, s scenario.Scenario) ExecutionResult
}

// Config controls a Run.
type Config struct {
	Target      string          // base URL of the running system; "" disables live execution
	Timeout     time.Duration   // per-request timeout
	MaxRetries  int             // retries for transient HTTP failures
	Concurrency int             // max scenarios executed in parallel
	Adapter     ai.AIAdapter    // AI backend for AI-driven execution
	Runner      *sandbox.Runner // reserved for future sandbox-aware execution
}

func (c Config) withDefaults() Config {
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	if c.MaxRetries < 0 {
		c.MaxRetries = DefaultMaxRetries
	}
	if c.Concurrency <= 0 {
		c.Concurrency = DefaultConcurrency
	}
	return c
}

// pick selects the executor for a single scenario given the config (hybrid dispatch).
func (c Config) pick(s scenario.Scenario) Executor {
	if c.Target != "" && s.Exec != nil && s.Exec.Mode == scenario.ExecHTTP {
		return &HTTPExecutor{BaseURL: c.Target, Timeout: c.Timeout, MaxRetries: c.MaxRetries}
	}
	if c.Target != "" && c.Adapter != nil {
		return &AIExecutor{Adapter: c.Adapter, Target: c.Target}
	}
	return &StaticExecutor{}
}

// Run executes every scenario using a bounded worker pool. Results are returned
// in the same order as the input scenarios. Respects ctx cancellation.
func Run(ctx context.Context, cfg Config, scenarios []scenario.Scenario) []ExecutionResult {
	cfg = cfg.withDefaults()

	results := make([]ExecutionResult, len(scenarios))
	sem := make(chan struct{}, cfg.Concurrency)
	done := make(chan int, len(scenarios))

	for i, s := range scenarios {
		select {
		case <-ctx.Done():
			// Mark remaining scenarios as not executed due to cancellation.
			results[i] = ExecutionResult{
				ScenarioID:    s.ScenarioID,
				Area:          s.ConfidenceArea,
				Priority:      s.Priority,
				Mode:          ModeStatic,
				Passed:        false,
				FailureReason: "execution cancelled: " + ctx.Err().Error(),
			}
			done <- i
			continue
		case sem <- struct{}{}:
		}

		go func(i int, s scenario.Scenario) {
			defer func() { <-sem }()
			ex := cfg.pick(s)
			r := ex.Execute(ctx, s)
			// Ensure identity fields are always populated regardless of executor.
			r.ScenarioID = s.ScenarioID
			r.Area = s.ConfidenceArea
			r.Priority = s.Priority
			results[i] = r
			done <- i
		}(i, s)
	}

	for range scenarios {
		<-done
	}
	return results
}
