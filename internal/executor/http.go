package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// HTTPExecutor deterministically executes scenarios that carry a structured
// http exec block, by firing the request against BaseURL.
type HTTPExecutor struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

// Execute fires the scenario's HTTP request (with retries) and judges the result.
func (e *HTTPExecutor) Execute(ctx context.Context, s scenario.Scenario) ExecutionResult {
	start := time.Now()
	res := ExecutionResult{
		ScenarioID: s.ScenarioID,
		Area:       s.ConfidenceArea,
		Priority:   s.Priority,
		Mode:       ModeHTTP,
	}

	if s.Exec == nil {
		res.FailureReason = "http executor: scenario has no exec block"
		res.Duration = time.Since(start)
		return res
	}

	// Concurrency race mode: fire N identical requests at once, expect one winner.
	if s.Exec.Concurrency > 1 {
		return e.executeRace(ctx, s)
	}

	method := s.Exec.Method
	if method == "" {
		method = http.MethodGet
	}
	url := joinURL(e.BaseURL, s.Exec.Path)

	client := &http.Client{Timeout: e.Timeout}

	var lastErr error
	var status int
	attempts := e.MaxRetries + 1
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			// Simple linear backoff; abort early if context is done.
			select {
			case <-ctx.Done():
				res.FailureReason = "context cancelled: " + ctx.Err().Error()
				res.ActualBehavior = "request aborted"
				res.Duration = time.Since(start)
				return res
			case <-time.After(time.Duration(attempt) * 100 * time.Millisecond):
			}
		}

		var body io.Reader
		if s.Exec.Body != "" {
			body = strings.NewReader(s.Exec.Body)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			// Malformed request — not retryable.
			res.FailureReason = "build request: " + err.Error()
			res.ActualBehavior = "request not sent"
			res.Duration = time.Since(start)
			return res
		}
		for k, v := range s.Exec.Headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue // transport error — retry
		}
		status = resp.StatusCode
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if status >= 500 && attempt < attempts-1 {
			lastErr = fmt.Errorf("server error %d", status)
			continue // 5xx — retry
		}
		lastErr = nil
		break
	}

	res.Duration = time.Since(start)

	if lastErr != nil {
		res.Passed = false
		res.ActualBehavior = fmt.Sprintf("%s %s failed", method, url)
		res.FailureReason = lastErr.Error()
		return res
	}

	res.ActualBehavior = fmt.Sprintf("HTTP %d in %dms", status, res.Duration.Milliseconds())
	res.Passed = judgeStatus(status, s.Exec.ExpectedStatus)
	if !res.Passed {
		if s.Exec.ExpectedStatus > 0 {
			res.FailureReason = fmt.Sprintf("expected status %d, got %d", s.Exec.ExpectedStatus, status)
		} else {
			res.FailureReason = fmt.Sprintf("server error status %d", status)
		}
	}
	return res
}

// executeRace fires Exec.Concurrency identical requests simultaneously and judges
// whether the system yielded exactly one winner with the rest cleanly rejected.
// This is the deterministic test for race guards (double-submit, claim-last-item).
func (e *HTTPExecutor) executeRace(ctx context.Context, s scenario.Scenario) ExecutionResult {
	start := time.Now()
	res := ExecutionResult{
		ScenarioID: s.ScenarioID,
		Area:       s.ConfidenceArea,
		Priority:   s.Priority,
		Mode:       ModeHTTP,
	}

	n := s.Exec.Concurrency
	method := s.Exec.Method
	if method == "" {
		method = http.MethodGet
	}
	url := joinURL(e.BaseURL, s.Exec.Path)
	client := &http.Client{Timeout: e.Timeout}

	type outcome struct {
		status int
		err    error
	}
	outcomes := make([]outcome, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			var body io.Reader
			if s.Exec.Body != "" {
				body = strings.NewReader(s.Exec.Body)
			}
			req, err := http.NewRequestWithContext(ctx, method, url, body)
			if err != nil {
				outcomes[i] = outcome{err: err}
				return
			}
			for k, v := range s.Exec.Headers {
				req.Header.Set(k, v)
			}
			resp, err := client.Do(req)
			if err != nil {
				outcomes[i] = outcome{err: err}
				return
			}
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
			resp.Body.Close()
			outcomes[i] = outcome{status: resp.StatusCode}
		}(i)
	}
	wg.Wait()
	res.Duration = time.Since(start)

	// Tally outcomes into success / rejected / server-error / transport-error.
	var winners, rejected, serverErr, transportErr int
	counts := map[int]int{}
	for _, o := range outcomes {
		if o.err != nil {
			transportErr++
			continue
		}
		counts[o.status]++
		switch {
		case isSuccess(o.status, s.Exec.ExpectedStatus):
			winners++
		case o.status >= 500:
			serverErr++
		case o.status >= 400:
			rejected++
		default:
			// Non-success 2xx/3xx that doesn't match expected_status — treat as a
			// winner-class anomaly so multiple "successes" are caught.
			winners++
		}
	}

	res.ActualBehavior = fmt.Sprintf("%d concurrent %s %s: %s", n, method, url, histogram(counts, transportErr))

	switch {
	case transportErr > 0:
		res.Passed = false
		res.FailureReason = fmt.Sprintf("%d/%d requests failed to complete (transport error)", transportErr, n)
	case serverErr > 0:
		res.Passed = false
		res.FailureReason = fmt.Sprintf("%d/%d requests returned a server error (5xx)", serverErr, n)
	case winners == 1 && rejected == n-1:
		res.Passed = true
	case winners > 1:
		res.Passed = false
		res.FailureReason = fmt.Sprintf("race not guarded: %d concurrent requests succeeded, expected exactly 1", winners)
	case winners == 0:
		res.Passed = false
		res.FailureReason = "no request succeeded — every concurrent request was rejected"
	default:
		res.Passed = false
		res.FailureReason = fmt.Sprintf("ambiguous outcome: %d succeeded, %d rejected of %d", winners, rejected, n)
	}
	return res
}

// isSuccess reports whether status is in the success class for a scenario.
// expected>0 means "exactly this status"; expected==0 means "any 2xx".
func isSuccess(status, expected int) bool {
	if expected > 0 {
		return status == expected
	}
	return status >= 200 && status < 300
}

// histogram renders a compact "N×code" summary, sorted by status code.
func histogram(counts map[int]int, transportErr int) string {
	codes := make([]int, 0, len(counts))
	for c := range counts {
		codes = append(codes, c)
	}
	sort.Ints(codes)
	parts := make([]string, 0, len(codes)+1)
	for _, c := range codes {
		parts = append(parts, fmt.Sprintf("%d×%d", counts[c], c))
	}
	if transportErr > 0 {
		parts = append(parts, fmt.Sprintf("%d×error", transportErr))
	}
	return strings.Join(parts, ", ")
}

// judgeStatus returns true if the response status is acceptable.
// expected==0 means "any non-5xx is a pass".
func judgeStatus(status, expected int) bool {
	if expected > 0 {
		return status == expected
	}
	return status < 500
}

// joinURL concatenates a base URL and a path, tolerating slash mismatches.
func joinURL(base, path string) string {
	if path == "" {
		return base
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}
