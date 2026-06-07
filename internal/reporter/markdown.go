package reporter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteMarkdown writes a Markdown report to w.
func WriteMarkdown(w io.Writer, result RunResult) error {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# teststop Report — %s\n\n", result.ProjectName))

	es := result.ExecSummary

	// Metadata
	sb.WriteString(fmt.Sprintf("**Date:** %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Duration:** %dms\n", result.Duration.Milliseconds()))
	sb.WriteString(fmt.Sprintf("**System:** %s (%s)\n", result.SystemType, result.Language))
	if es.Executed {
		sb.WriteString(fmt.Sprintf("**Mode:** executed against %s\n", es.Target))
		sb.WriteString(fmt.Sprintf("**Confidence:** %.1f%%\n", result.ConfidenceScore*100))
	} else {
		sb.WriteString("**Mode:** predicted (no `--target` — structural only, not executed)\n")
		sb.WriteString(fmt.Sprintf("**Predicted confidence:** %.1f%% _(run with `--target` to verify)_\n", result.ConfidenceScore*100))
	}

	exitMeaning := exitCodeMeaning(result.ExitCode)
	sb.WriteString(fmt.Sprintf("**Exit Code:** %d (%s)\n\n", result.ExitCode, exitMeaning))

	// Scenarios table — "Scenarios" when executed, "Predicted Risks" otherwise.
	scenTitle := "Scenarios"
	if !es.Executed {
		scenTitle = "Predicted Risks"
	}
	sb.WriteString(fmt.Sprintf("## %s (%d total)\n\n", scenTitle, len(result.Scenarios)))
	sb.WriteString("| Priority | Title | Area | Edge Case |\n")
	sb.WriteString("|----------|-------|------|-----------|\n")
	for _, sc := range result.Scenarios {
		edgeCase := "no"
		if sc.IsEdgeCase {
			edgeCase = "yes"
		}
		sb.WriteString(fmt.Sprintf(
			"| %s | %s | %s | %s |\n",
			escapeMarkdown(sc.Priority),
			escapeMarkdown(sc.Title),
			escapeMarkdown(sc.ConfidenceArea),
			edgeCase,
		))
	}
	sb.WriteString("\n")

	// Execution section
	sb.WriteString("## Execution\n\n")
	if es.Executed {
		sb.WriteString(fmt.Sprintf("- **Target:** %s\n", es.Target))
		sb.WriteString(fmt.Sprintf(
			"- **Results:** %d passed, %d failed of %d executed\n\n",
			es.Passed, es.Failed, es.Count,
		))
	} else {
		sb.WriteString("- **Target:** _none — predicted only, not executed_\n")
		sb.WriteString(fmt.Sprintf(
			"- **Results:** %d scenarios predicted. Run with `--target <url>` to execute and verify.\n\n",
			es.Count,
		))
	}

	// Failures / predicted-issues section
	failTitle := "Failures"
	if !es.Executed {
		failTitle = "Predicted Failure Modes"
	}
	sb.WriteString(fmt.Sprintf("## %s (%d)\n\n", failTitle, len(result.Failures)))
	if len(result.Failures) == 0 {
		sb.WriteString("_(none)_\n\n")
	}
	for _, f := range result.Failures {
		sb.WriteString(fmt.Sprintf("### %s\n\n", f.Title))
		sb.WriteString(fmt.Sprintf("- **Area:** %s\n", f.Area))
		sb.WriteString(fmt.Sprintf("- **Priority:** %s\n", f.Priority))
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n\n", f.Description))
	}

	// Memory state section
	sb.WriteString("## Memory State\n\n")

	stableList := "none"
	if len(result.StableAreas) > 0 {
		stableList = strings.Join(result.StableAreas, ", ")
	}
	volatileList := "none"
	if len(result.VolatileAreas) > 0 {
		volatileList = strings.Join(result.VolatileAreas, ", ")
	}
	retiredList := "none"
	if len(result.RetiredAreas) > 0 {
		retiredList = strings.Join(result.RetiredAreas, ", ")
	}

	sb.WriteString(fmt.Sprintf("- **Stable areas:** %s\n", stableList))
	sb.WriteString(fmt.Sprintf("- **Volatile areas:** %s\n", volatileList))
	sb.WriteString(fmt.Sprintf("- **Retired areas:** %s\n", retiredList))
	sb.WriteString("\n")

	_, err := io.WriteString(w, sb.String())
	if err != nil {
		return fmt.Errorf("reporter: %w", err)
	}
	return nil
}

// SaveMarkdownReport saves a markdown report to projectPath/.teststop/reports/YYYY-MM-DD-HH-MM-SS.md.
// It returns the full path to the created file.
func SaveMarkdownReport(projectPath string, result RunResult) (string, error) {
	dir := filepath.Join(projectPath, ".teststop", "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("reporter: create reports dir: %w", err)
	}

	filename := time.Now().Format("2006-01-02-15-04-05") + ".md"
	fullPath := filepath.Join(dir, filename)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("reporter: create report file: %w", err)
	}
	defer f.Close()

	if err := WriteMarkdown(f, result); err != nil {
		return "", err
	}

	return fullPath, nil
}

// exitCodeMeaning returns a short description of an exit code.
func exitCodeMeaning(code int) string {
	switch code {
	case 0:
		return "confidence threshold met"
	case 1:
		return "review needed"
	case 2:
		return "critical failures found"
	case 3:
		return "teststop internal error"
	default:
		return "unknown"
	}
}

// escapeMarkdown escapes pipe characters in table cells.
func escapeMarkdown(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
