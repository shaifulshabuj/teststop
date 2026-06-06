package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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
