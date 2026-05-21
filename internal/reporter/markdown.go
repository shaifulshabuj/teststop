package reporter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/shaifulshabuj/teststop/pkg/scenario"
)

// WriteMarkdown renders a Markdown report suitable for a PR comment or
// a wiki page. It is intentionally section-heavy so a reader can skim
// to the part they care about.
func WriteMarkdown(w io.Writer, r Run) error {
	var b strings.Builder

	fmt.Fprintf(&b, "# teststop report — %s\n\n", r.Project)
	fmt.Fprintf(&b, "_Run: `%s` · %s_\n\n", r.RunID, r.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"))

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Project:** `%s` — %s (%s)\n", r.Project, r.Language, r.ProjectType)
	fmt.Fprintf(&b, "- **Maturity:** `%s`\n", r.MaturityStage)
	fmt.Fprintf(&b, "- **Confidence:** **%.2f** (Δ %+0.2f) against threshold **%.2f**\n",
		r.OverallConfidence, r.ConfidenceDelta, r.Threshold)
	fmt.Fprintf(&b, "- **Scenarios:** %d generated · %d passed · %d failed",
		r.ScenariosGenerated, r.ScenariosPassed, r.ScenariosFailed)
	if r.ScenariosUnknown > 0 {
		fmt.Fprintf(&b, " · %d unknown", r.ScenariosUnknown)
	}
	b.WriteString("\n")

	switch r.ExitCode() {
	case 0:
		b.WriteString("- **Status:** ✅ above threshold\n")
	case 2:
		b.WriteString("- **Status:** ⛔ critical failures — do not deploy\n")
	default:
		b.WriteString("- **Status:** ⚠️ below threshold — review required\n")
	}
	b.WriteString("\n")

	if len(r.Failures) > 0 {
		b.WriteString("## Failures\n\n")
		grouped := SummariseFailures(r.Failures)
		for _, p := range []scenario.Priority{
			scenario.PriorityCritical, scenario.PriorityHigh,
			scenario.PriorityMedium, scenario.PriorityLow,
		} {
			fs := grouped[p]
			if len(fs) == 0 {
				continue
			}
			fmt.Fprintf(&b, "### %s\n\n", strings.ToUpper(string(p)))
			for _, f := range fs {
				fmt.Fprintf(&b, "- `%s` — %s", f.ScenarioID, f.Title)
				if f.ConfidenceArea != "" {
					fmt.Fprintf(&b, " _(area: `%s`)_", f.ConfidenceArea)
				}
				b.WriteString("\n")
				if f.Notes != "" {
					fmt.Fprintf(&b, "  - %s\n", f.Notes)
				}
			}
			b.WriteString("\n")
		}
	}

	if len(r.StableAreas) > 0 {
		b.WriteString("## Stable areas\n\n")
		stable := append([]string(nil), r.StableAreas...)
		sort.Strings(stable)
		for _, a := range stable {
			fmt.Fprintf(&b, "- `%s`\n", a)
		}
		b.WriteString("\n")
	}

	if len(r.VolatileAreas) > 0 {
		b.WriteString("## Volatile areas\n\n")
		vol := append([]string(nil), r.VolatileAreas...)
		sort.Strings(vol)
		for _, a := range vol {
			fmt.Fprintf(&b, "- `%s`\n", a)
		}
		b.WriteString("\n")
	}

	if len(r.RetiredThisRun) > 0 {
		b.WriteString("## Retired this run\n\n")
		for _, a := range r.RetiredThisRun {
			fmt.Fprintf(&b, "- `%s`\n", a)
		}
		b.WriteString("\n")
	}

	for _, n := range r.Notes {
		fmt.Fprintf(&b, "> %s\n\n", n)
	}

	_, err := io.WriteString(w, b.String())
	return err
}
