package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// WriteText renders a compact, terminal-friendly summary. The format is
// intentionally close to what the README documents so the on-disk shape
// and the on-screen shape match.
func WriteText(w io.Writer, r Run) error {
	var b strings.Builder

	fmt.Fprintf(&b, "teststop v0.1.0\n\n")
	fmt.Fprintf(&b, "Project:  %s (%s, %s)\n", r.Project, r.Language, r.ProjectType)
	fmt.Fprintf(&b, "Stage:    %s\n", r.MaturityStage)
	fmt.Fprintf(&b, "Memory:   %d stable, %d volatile\n", len(r.StableAreas), len(r.VolatileAreas))
	b.WriteString("\n")

	fmt.Fprintf(&b, "Scenarios: %d generated\n", r.ScenariosGenerated)
	if r.ScenariosPassed+r.ScenariosFailed+r.ScenariosUnknown > 0 {
		fmt.Fprintf(&b, "  Passed:   %d\n", r.ScenariosPassed)
		fmt.Fprintf(&b, "  Failed:   %d\n", r.ScenariosFailed)
		if r.ScenariosUnknown > 0 {
			fmt.Fprintf(&b, "  Unknown:  %d\n", r.ScenariosUnknown)
		}
	}
	b.WriteString("\n")

	if len(r.Failures) > 0 {
		b.WriteString("Failures:\n")
		for _, f := range r.Failures {
			fmt.Fprintf(&b, "  [%s] %s — %s\n",
				strings.ToUpper(string(f.Priority)), f.ScenarioID, f.Title)
		}
		b.WriteString("\n")
	}

	deltaStr := ""
	switch {
	case r.ConfidenceDelta > 0:
		deltaStr = fmt.Sprintf(" (+%.2f)", r.ConfidenceDelta)
	case r.ConfidenceDelta < 0:
		deltaStr = fmt.Sprintf(" (%.2f)", r.ConfidenceDelta)
	}
	fmt.Fprintf(&b, "Confidence: %.2f%s\n", r.OverallConfidence, deltaStr)
	fmt.Fprintf(&b, "Threshold:  %.2f\n", r.Threshold)

	status := "Below threshold — review required"
	switch r.ExitCode() {
	case 0:
		status = "Above threshold — safe to proceed"
	case 2:
		status = "Critical failures — do not deploy"
	}
	fmt.Fprintf(&b, "Status:     %s\n", status)

	if len(r.RetiredThisRun) > 0 {
		fmt.Fprintf(&b, "\nRetired this run: %s\n", strings.Join(r.RetiredThisRun, ", "))
	}

	for _, n := range r.Notes {
		fmt.Fprintf(&b, "\nNote: %s\n", n)
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// SummariseFailures groups failures by priority for the text/markdown
// renderers when they want a clearer breakdown than a flat list.
func SummariseFailures(fs []Failure) map[scenario.Priority][]Failure {
	out := map[scenario.Priority][]Failure{}
	for _, f := range fs {
		out[f.Priority] = append(out[f.Priority], f)
	}
	return out
}
