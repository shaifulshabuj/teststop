package executor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

func TestConfigPick_Dispatch(t *testing.T) {
	httpSc := scenario.Scenario{Exec: &scenario.ExecSpec{Mode: scenario.ExecHTTP}}
	proseSc := scenario.Scenario{ScenarioID: "p"}

	cases := []struct {
		name string
		cfg  Config
		sc   scenario.Scenario
		want any
	}{
		{"http when exec+target", Config{Target: "http://x"}, httpSc, &HTTPExecutor{}},
		{"ai when target+adapter, no exec", Config{Target: "http://x", Adapter: &fakeAdapter{}}, proseSc, &AIExecutor{}},
		{"static when no target", Config{}, httpSc, &StaticExecutor{}},
		{"static when target but no adapter and no exec", Config{Target: "http://x"}, proseSc, &StaticExecutor{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.cfg.pick(c.sc)
			if fmt.Sprintf("%T", got) != fmt.Sprintf("%T", c.want) {
				t.Errorf("pick = %T, want %T", got, c.want)
			}
		})
	}
}

func TestRun_OrderStableAndComplete(t *testing.T) {
	// Static-only run (no target): every well-formed scenario passes, order preserved.
	n := 20
	scenarios := make([]scenario.Scenario, n)
	for i := range scenarios {
		scenarios[i] = scenario.Scenario{
			ScenarioID:       fmt.Sprintf("s%02d", i),
			ConfidenceArea:   "area",
			Steps:            []string{"a", "b"},
			ExpectedBehavior: "ok",
		}
	}

	results := Run(context.Background(), Config{Concurrency: 5}, scenarios)

	if len(results) != n {
		t.Fatalf("got %d results, want %d", len(results), n)
	}
	for i, r := range results {
		want := fmt.Sprintf("s%02d", i)
		if r.ScenarioID != want {
			t.Errorf("results[%d].ScenarioID = %q, want %q (order not preserved)", i, r.ScenarioID, want)
		}
		if !r.Passed {
			t.Errorf("results[%d] should pass static validation", i)
		}
	}
}

func TestRun_ConcurrencyIsBounded(t *testing.T) {
	var inflight, maxInflight int32
	var mu = make(chan struct{}, 1)
	mu <- struct{}{} // token

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-mu
		inflight++
		if inflight > maxInflight {
			maxInflight = inflight
		}
		mu <- struct{}{}
		time.Sleep(20 * time.Millisecond)
		<-mu
		inflight--
		mu <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := 12
	scenarios := make([]scenario.Scenario, n)
	for i := range scenarios {
		scenarios[i] = httpScenario(http.MethodGet, "/", 200)
		scenarios[i].ScenarioID = fmt.Sprintf("s%d", i)
	}

	const limit = 3
	results := Run(context.Background(), Config{Target: srv.URL, Concurrency: limit, Timeout: 2 * time.Second}, scenarios)

	if len(results) != n {
		t.Fatalf("got %d results, want %d", len(results), n)
	}
	if maxInflight > limit {
		t.Errorf("max in-flight = %d, exceeds concurrency limit %d", maxInflight, limit)
	}
	for _, r := range results {
		if !r.Passed {
			t.Errorf("scenario %s failed: %s", r.ScenarioID, r.FailureReason)
		}
	}
}

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	scenarios := []scenario.Scenario{
		{ScenarioID: "s0", ConfidenceArea: "a", Steps: []string{"x"}, ExpectedBehavior: "ok"},
	}
	results := Run(ctx, Config{}, scenarios)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	// With a cancelled context, static execution may still complete; the key
	// guarantee is we return exactly one result per scenario without panicking.
	if results[0].ScenarioID != "s0" {
		t.Errorf("ScenarioID = %q, want s0", results[0].ScenarioID)
	}
}
