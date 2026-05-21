// Package mandate composes the canonical adversarial-user mandate with
// project context and accumulated memory so the AI receives a single,
// fully-resolved instruction.
//
// The composer is intentionally substitutional, not templated. Adding a
// template engine would make the mandate harder to read at the source
// and harder to audit with `teststop mandate --show`. Placeholders are
// the literal substrings the base mandate contains, e.g. {{PROJECT_NAME}}.
package mandate

import (
	"fmt"
	"sort"
	"strings"

	embedded "github.com/shaifulshabuj/teststop/mandate"
	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
)

// Depth tunes how many scenarios the AI is asked to produce.
type Depth string

const (
	DepthLight      Depth = "light"
	DepthNormal     Depth = "normal"
	DepthAggressive Depth = "aggressive"
)

// Input is everything the composer needs to fill the mandate. Keeping
// it explicit (rather than a long function signature) makes the call
// sites in the CLI easier to read.
type Input struct {
	Project *reader.ProjectContext
	Memory  *memory.Memory
	Depth   Depth
}

// Compose returns the fully resolved mandate text ready to send to an AI.
func Compose(in Input) (string, error) {
	if in.Project == nil {
		return "", fmt.Errorf("mandate: project context required")
	}
	if in.Memory == nil {
		in.Memory = memory.New()
	}
	if in.Depth == "" {
		in.Depth = DepthNormal
	}

	text := embedded.Base
	replacements := buildReplacements(in)
	for k, v := range replacements {
		text = strings.ReplaceAll(text, k, v)
	}
	return text, nil
}

func buildReplacements(in Input) map[string]string {
	p := in.Project
	m := in.Memory

	return map[string]string{
		"{{PROJECT_NAME}}":              valueOr(p.Name, "the project"),
		"{{DETECTED_LANGUAGE}}":         valueOr(p.Language, "unknown"),
		"{{DETECTED_TYPE}}":             valueOr(p.Type, "service"),
		"{{DETECTED_ENTRY_POINTS}}":     listOrNone(p.EntryPoints),
		"{{DETECTED_FLOWS}}":            formatFlows(p.KeyFlows),
		"{{DETECTED_COMPLEXITY}}":       valueOr(p.Complexity, "moderate"),
		"{{DETECTED_FILE_COUNT}}":       fmt.Sprintf("%d", p.FileCount),
		"{{MEMORY_STABLE_AREAS}}":       formatAreas(m.StableAreas(), "none yet — every area is unproven"),
		"{{MEMORY_VOLATILE_AREAS}}":     formatAreas(m.VolatileAreas(), "none — pick the surfaces a new user would touch first"),
		"{{MEMORY_MATURITY_STAGE}}":     valueOr(m.MaturityStage, memory.StageNew),
		"{{MEMORY_OVERALL_CONFIDENCE}}": fmt.Sprintf("%.2f", m.OverallConfidence),
		"{{SCENARIO_COUNT}}":            fmt.Sprintf("%d", scenarioCount(in)),
	}
}

// scenarioCount picks a scenario budget from project complexity and
// maturity. Mature projects get fewer because most surface is already
// proven; new/complex projects get more.
func scenarioCount(in Input) int {
	base := 12
	switch in.Project.Complexity {
	case "simple":
		base = 8
	case "complex":
		base = 20
	}

	switch in.Memory.MaturityStage {
	case memory.StageMature:
		base = int(float64(base) * 0.6)
	case memory.StageLegacy:
		base = int(float64(base) * 0.4)
	}

	switch in.Depth {
	case DepthLight:
		base = int(float64(base) * 0.5)
	case DepthAggressive:
		base = int(float64(base) * 1.75)
	}

	if base < 3 {
		base = 3
	}
	if base > 50 {
		base = 50
	}
	return base
}

func valueOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func listOrNone(items []string) string {
	if len(items) == 0 {
		return "none detected"
	}
	cp := append([]string(nil), items...)
	sort.Strings(cp)
	return strings.Join(cp, ", ")
}

func formatFlows(flows []reader.Flow) string {
	if len(flows) == 0 {
		return "none detected — treat the project as an opaque box and explore its surface"
	}
	// Group by kind so the mandate is easier for the AI to scan.
	byKind := map[string][]string{}
	for _, f := range flows {
		byKind[f.Kind] = append(byKind[f.Kind], fmt.Sprintf("%s (%s)", f.Name, f.Location))
	}

	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)

	var b strings.Builder
	for i, k := range kinds {
		if i > 0 {
			b.WriteString("\n  ")
		}
		sort.Strings(byKind[k])
		fmt.Fprintf(&b, "%s: %s", k, strings.Join(byKind[k], "; "))
	}
	return b.String()
}

func formatAreas(areas []string, empty string) string {
	if len(areas) == 0 {
		return empty
	}
	cp := append([]string(nil), areas...)
	sort.Strings(cp)
	var b strings.Builder
	for _, a := range cp {
		fmt.Fprintf(&b, "- %s\n", a)
	}
	return strings.TrimRight(b.String(), "\n")
}
