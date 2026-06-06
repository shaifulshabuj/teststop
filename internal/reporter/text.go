package reporter

import (
	"fmt"
	"io"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// allCodes lists every ANSI escape sequence used, for stripping in no-color mode.
var allCodes = []string{
	colorReset, colorRed, colorGreen, colorYellow, colorBlue, colorBold, colorDim,
}

// stripANSI removes all known ANSI escape codes from s.
func stripANSI(s string) string {
	for _, code := range allCodes {
		s = strings.ReplaceAll(s, code, "")
	}
	return s
}

// WriteText writes a human-readable ANSI terminal report to w.
// If noColor is true, ANSI codes are stripped from the output.
func WriteText(w io.Writer, result RunResult, noColor bool) error {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf(
		"%steststop%s — %s%s%s (%s, %s)\n",
		colorBold, colorReset,
		colorBold, result.ProjectName, colorReset,
		result.SystemType, result.Language,
	))
	sb.WriteString(fmt.Sprintf(
		"%sRun at %s  Duration: %dms%s\n\n",
		colorDim,
		result.Timestamp.Format("2006-01-02 15:04:05"),
		result.Duration.Milliseconds(),
		colorReset,
	))

	// Scenarios section
	sb.WriteString(fmt.Sprintf("%sSCENARIOS (%d total)%s\n", colorBold, len(result.Scenarios), colorReset))
	sb.WriteString("─────────────────────────────\n")

	// Build a set of failed scenario IDs for quick lookup
	failedIDs := make(map[string]bool, len(result.Failures))
	for _, f := range result.Failures {
		failedIDs[f.ScenarioID] = true
	}

	for _, sc := range result.Scenarios {
		var marker, markerColor string
		if failedIDs[sc.ScenarioID] {
			marker = "✗"
			markerColor = colorRed
		} else {
			marker = "✓"
			markerColor = colorGreen
		}

		priorityColor := ""
		if sc.Priority == "critical" {
			priorityColor = colorRed + colorBold
		}

		sb.WriteString(fmt.Sprintf(
			"  %s%s%s %s%s%s %s\n",
			markerColor, marker, colorReset,
			priorityColor, sc.Priority, colorReset,
			sc.Title,
		))
	}
	sb.WriteString("\n")

	// Execution section
	es := result.ExecSummary
	target := es.Target
	if target == "" {
		target = "(none — static validation)"
	}
	sb.WriteString(fmt.Sprintf("%sEXECUTION%s\n", colorBold, colorReset))
	sb.WriteString("─────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  Target:  %s\n", target))
	sb.WriteString(fmt.Sprintf(
		"  Results: %s%d passed%s, %s%d failed%s of %d executed\n\n",
		colorGreen, es.Passed, colorReset,
		colorRed, es.Failed, colorReset,
		es.Executed,
	))

	// Failures section
	sb.WriteString(fmt.Sprintf("%sFAILURES (%d)%s\n", colorBold, len(result.Failures), colorReset))
	sb.WriteString("─────────────────────────────\n")
	if len(result.Failures) == 0 {
		sb.WriteString(fmt.Sprintf("  %s(none)%s\n", colorDim, colorReset))
	}
	for _, f := range result.Failures {
		priorityColor := colorRed
		if f.Priority == "critical" {
			priorityColor = colorRed + colorBold
		}
		sb.WriteString(fmt.Sprintf(
			"  %s✗ %s%s\n",
			colorRed, f.Title, colorReset,
		))
		sb.WriteString(fmt.Sprintf(
			"    %sArea: %s%s\n",
			colorDim, f.Area, colorReset,
		))
		sb.WriteString(fmt.Sprintf(
			"    %s%s%s\n",
			priorityColor, f.Description, colorReset,
		))
	}
	sb.WriteString("\n")

	// Memory section
	sb.WriteString(fmt.Sprintf("%sMEMORY%s\n", colorBold, colorReset))
	sb.WriteString("─────────────────────────────\n")

	stableList := "(none)"
	if len(result.StableAreas) > 0 {
		stableList = strings.Join(result.StableAreas, ", ")
	}
	volatileList := "(none)"
	if len(result.VolatileAreas) > 0 {
		volatileList = strings.Join(result.VolatileAreas, ", ")
	}
	retiredList := "(none)"
	if len(result.RetiredAreas) > 0 {
		retiredList = strings.Join(result.RetiredAreas, ", ")
	}

	sb.WriteString(fmt.Sprintf("  Stable areas:   %s\n", stableList))
	sb.WriteString(fmt.Sprintf("  Volatile areas: %s\n", volatileList))
	sb.WriteString(fmt.Sprintf("  Retired areas:  %s\n", retiredList))
	sb.WriteString("\n")

	// Confidence line
	scorePercent := result.ConfidenceScore * 100
	confidenceColor := colorGreen
	if scorePercent < 80 {
		confidenceColor = colorYellow
	}
	if len(result.Failures) > 0 {
		for _, f := range result.Failures {
			if f.Priority == "critical" {
				confidenceColor = colorRed
				break
			}
		}
	}

	status := "OK"
	switch ExitCodeFor(result, 0.80) {
	case 1:
		status = "REVIEW NEEDED"
	case 2:
		status = "CRITICAL FAILURES"
	}

	sb.WriteString(fmt.Sprintf(
		"%sCONFIDENCE: %s%.1f%%%s %s\n",
		colorBold, confidenceColor, scorePercent, colorReset,
		status,
	))

	out := sb.String()
	if noColor {
		out = stripANSI(out)
	}

	_, err := io.WriteString(w, out)
	if err != nil {
		return fmt.Errorf("reporter: %w", err)
	}
	return nil
}
