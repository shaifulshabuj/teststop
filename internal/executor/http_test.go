package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

func httpScenario(method, path string, expected int) scenario.Scenario {
	return scenario.Scenario{
		ScenarioID:     "s1",
		ConfidenceArea: "api",
		Priority:       scenario.PriorityHigh,
		Exec: &scenario.ExecSpec{
			Mode:           scenario.ExecHTTP,
			Method:         method,
			Path:           path,
			ExpectedStatus: expected,
		},
	}
}

func TestHTTPExecutor_Pass(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 2 * time.Second, MaxRetries: 0}
	res := ex.Execute(context.Background(), httpScenario(http.MethodGet, "/health", 200))

	if !res.Passed {
		t.Fatalf("expected pass, got fail: %s", res.FailureReason)
	}
	if res.Mode != ModeHTTP {
		t.Errorf("mode = %q, want %q", res.Mode, ModeHTTP)
	}
}

func TestHTTPExecutor_StatusMismatchFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200, but scenario expects 400
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 2 * time.Second, MaxRetries: 0}
	res := ex.Execute(context.Background(), httpScenario(http.MethodPost, "/api/login", 400))

	if res.Passed {
		t.Fatal("expected fail on status mismatch, got pass")
	}
	if res.FailureReason == "" {
		t.Error("expected a failure reason")
	}
}

func TestHTTPExecutor_ExpectedZeroAcceptsNon5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418, non-5xx
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 2 * time.Second, MaxRetries: 0}
	res := ex.Execute(context.Background(), httpScenario(http.MethodGet, "/", 0))

	if !res.Passed {
		t.Fatalf("expected pass for non-5xx with expected=0, got: %s", res.FailureReason)
	}
}

func TestHTTPExecutor_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 20 * time.Millisecond, MaxRetries: 0}
	res := ex.Execute(context.Background(), httpScenario(http.MethodGet, "/slow", 200))

	if res.Passed {
		t.Fatal("expected fail on timeout, got pass")
	}
}

func TestHTTPExecutor_RetryThenSucceed(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError) // first call 500
			return
		}
		w.WriteHeader(http.StatusOK) // subsequent calls succeed
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 2 * time.Second, MaxRetries: 2}
	res := ex.Execute(context.Background(), httpScenario(http.MethodGet, "/flaky", 200))

	if !res.Passed {
		t.Fatalf("expected pass after retry, got: %s", res.FailureReason)
	}
	if got := calls.Load(); got < 2 {
		t.Errorf("expected at least 2 calls (retry), got %d", got)
	}
}

func raceScenario(n, expected int) scenario.Scenario {
	return scenario.Scenario{
		ScenarioID:     "race1",
		ConfidenceArea: "actions/approve",
		Priority:       scenario.PriorityCritical,
		Exec: &scenario.ExecSpec{
			Mode:           scenario.ExecHTTP,
			Method:         http.MethodPost,
			Path:           "/approve",
			ExpectedStatus: expected,
			Concurrency:    n,
		},
	}
}

func TestHTTPExecutor_RaceGuardedOneWinner(t *testing.T) {
	var claimed atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claimed.CompareAndSwap(false, true) {
			w.WriteHeader(http.StatusOK) // exactly one winner
			return
		}
		w.WriteHeader(http.StatusConflict) // 409 for the rest
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 2 * time.Second}
	res := ex.Execute(context.Background(), raceScenario(10, 200))

	if !res.Passed {
		t.Fatalf("guarded race should pass, got fail: %s (%s)", res.FailureReason, res.ActualBehavior)
	}
}

func TestHTTPExecutor_RaceUnguardedFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // no guard — everyone wins
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 2 * time.Second}
	res := ex.Execute(context.Background(), raceScenario(8, 200))

	if res.Passed {
		t.Fatal("unguarded race should fail (multiple winners), got pass")
	}
	if !strings.Contains(res.FailureReason, "race not guarded") {
		t.Errorf("expected 'race not guarded' reason, got: %s", res.FailureReason)
	}
}

func TestHTTPExecutor_RaceServerErrorFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ex := &HTTPExecutor{BaseURL: srv.URL, Timeout: 2 * time.Second}
	res := ex.Execute(context.Background(), raceScenario(5, 200))

	if res.Passed {
		t.Fatal("server errors in a race should fail")
	}
}

func TestJoinURL(t *testing.T) {
	cases := []struct{ base, path, want string }{
		{"http://x", "/a", "http://x/a"},
		{"http://x/", "/a", "http://x/a"},
		{"http://x", "a", "http://x/a"},
		{"http://x", "", "http://x"},
	}
	for _, c := range cases {
		if got := joinURL(c.base, c.path); got != c.want {
			t.Errorf("joinURL(%q,%q) = %q, want %q", c.base, c.path, got, c.want)
		}
	}
}
