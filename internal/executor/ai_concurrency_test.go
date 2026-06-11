package executor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// slowAdapter is a test double whose Prompt blocks for delay, letting us measure
// real concurrency.
type slowAdapter struct {
	delay     time.Duration
	inflight  atomic.Int32
	maxSeen   atomic.Int32
	callCount atomic.Int32
}

func (s *slowAdapter) GenerateScenarios(string) ([]scenario.Scenario, error) { return nil, nil }
func (s *slowAdapter) Name() string                                          { return "slow" }
func (s *slowAdapter) Prompt(string) ([]byte, error) {
	s.callCount.Add(1)
	cur := s.inflight.Add(1)
	// Track the maximum concurrent calls observed.
	for {
		max := s.maxSeen.Load()
		if cur <= max || s.maxSeen.CompareAndSwap(max, cur) {
			break
		}
	}
	time.Sleep(s.delay)
	s.inflight.Add(-1)
	// Return a valid verdict so the result is not skipped.
	return []byte(`{"passed": true, "actual_behavior": "ok", "failure_reason": ""}`), nil
}

func aiModeScenarios(n int) []scenario.Scenario {
	out := make([]scenario.Scenario, n)
	for i := range out {
		out[i] = scenario.Scenario{
			ScenarioID:       string(rune('a' + i)),
			ConfidenceArea:   "area",
			Steps:            []string{"step"},
			ExpectedBehavior: "ok",
		}
	}
	return out
}

// TestAIConcurrency_DefaultIsOne verifies that with default AIConcurrency the
// AI adapter is never called more than once concurrently.
func TestAIConcurrency_DefaultIsOne(t *testing.T) {
	adapter := &slowAdapter{delay: 30 * time.Millisecond}
	cfg := Config{
		Target:      "http://localhost:9999",
		Adapter:     adapter,
		Concurrency: 8, // high general concurrency
		// AIConcurrency not set → should default to 1
	}
	scenarios := aiModeScenarios(6)

	Run(context.Background(), cfg, scenarios)

	if max := adapter.maxSeen.Load(); max > 1 {
		t.Errorf("max concurrent AI calls = %d, want ≤ 1 (default AIConcurrency)", max)
	}
	if cnt := adapter.callCount.Load(); int(cnt) != len(scenarios) {
		t.Errorf("expected %d AI calls, got %d", len(scenarios), cnt)
	}
}

// TestAIConcurrency_ExplicitCap verifies that AIConcurrency=2 allows up to 2
// concurrent AI calls even when general Concurrency is higher.
func TestAIConcurrency_ExplicitCap(t *testing.T) {
	const aiCap = 2
	adapter := &slowAdapter{delay: 30 * time.Millisecond}
	cfg := Config{
		Target:        "http://localhost:9999",
		Adapter:       adapter,
		Concurrency:   8,
		AIConcurrency: aiCap,
	}
	scenarios := aiModeScenarios(8)

	Run(context.Background(), cfg, scenarios)

	if max := adapter.maxSeen.Load(); max > aiCap {
		t.Errorf("max concurrent AI calls = %d, exceeds AIConcurrency cap %d", max, aiCap)
	}
}

// TestIsAIMode verifies the isAIMode helper classifies scenarios correctly.
func TestIsAIMode(t *testing.T) {
	httpSc := scenario.Scenario{Exec: &scenario.ExecSpec{Mode: scenario.ExecHTTP}}
	proseSc := scenario.Scenario{ScenarioID: "prose"}

	cases := []struct {
		name string
		cfg  Config
		sc   scenario.Scenario
		want bool
	}{
		{"no target → not AI", Config{}, proseSc, false},
		{"target+adapter+prose → AI", Config{Target: "http://x", Adapter: &fakeAdapter{}}, proseSc, true},
		{"target+adapter+http-exec → not AI", Config{Target: "http://x", Adapter: &fakeAdapter{}}, httpSc, false},
		{"target+no-adapter → not AI", Config{Target: "http://x"}, proseSc, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.cfg.isAIMode(c.sc); got != c.want {
				t.Errorf("isAIMode = %v, want %v", got, c.want)
			}
		})
	}
}

// TestDefaultAIConcurrency verifies the constant and withDefaults behaviour.
func TestDefaultAIConcurrency(t *testing.T) {
	if DefaultAIConcurrency != 1 {
		t.Errorf("DefaultAIConcurrency = %d, want 1", DefaultAIConcurrency)
	}
	cfg := Config{}.withDefaults()
	if cfg.AIConcurrency != 1 {
		t.Errorf("withDefaults AIConcurrency = %d, want 1", cfg.AIConcurrency)
	}
}
