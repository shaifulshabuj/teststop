package mandate

import (
	"fmt"
	"strings"

	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
	basemandatepkg "github.com/shaifulshabuj/teststop/mandate"
)

// Options controls how the mandate is composed.
type Options struct {
	// Depth controls testing depth: "light" | "normal" | "aggressive"
	Depth string
}

// scenarioCount returns the target number of scenarios for the given depth and flow count.
//
// Table (depth × complexity → N):
//
//	depth\complexity  simple(0-2)  moderate(3-7)  complex(8+)
//	light             5            10             15
//	normal            10           20             30
//	aggressive        20           35             50
func scenarioCount(depth string, flowCount int) int {
	table := [3][3]int{
		{5, 10, 15},  // light
		{10, 20, 30}, // normal
		{20, 35, 50}, // aggressive
	}

	var row int
	switch depth {
	case "light":
		row = 0
	case "aggressive":
		row = 2
	default: // "normal" and anything unrecognised
		row = 1
	}

	var col int
	switch {
	case flowCount <= 2:
		col = 0 // simple
	case flowCount <= 7:
		col = 1 // moderate
	default:
		col = 2 // complex
	}

	return table[row][col]
}

// Compose builds the final mandate string by injecting project context and accumulated
// memory into the embedded base mandate template.
//
// Placeholder tokens replaced (exact strings as they appear in mandate/base.md):
//
//	[SYSTEM_NAME]           → ctx.Name
//	[PROJECT_NAME]          → ctx.Name
//	[DETECTED_LANGUAGE]     → ctx.Language
//	[DETECTED_TYPE]         → ctx.Type
//	[DETECTED_ENTRY_POINTS] → comma-joined ctx.EntryPoints
//	[DETECTED_FLOWS]        → newline-joined "- Name: Description" entries
//	[MEMORY_STABLE_AREAS]   → comma-joined areas with Confidence >= 0.95 ("none" if empty)
//	[MEMORY_VOLATILE_AREAS] → comma-joined areas with Confidence < 0.75  ("none" if empty)
//	[N]                     → scenario count derived from depth × complexity table
func Compose(ctx reader.ProjectContext, mem *memory.Memory, opts Options) (string, error) {
	result := basemandatepkg.BaseMandateContent

	// Basic context substitutions.
	result = strings.ReplaceAll(result, "[SYSTEM_NAME]", ctx.Name)
	result = strings.ReplaceAll(result, "[PROJECT_NAME]", ctx.Name)
	result = strings.ReplaceAll(result, "[DETECTED_LANGUAGE]", ctx.Language)
	result = strings.ReplaceAll(result, "[DETECTED_TYPE]", ctx.Type)
	result = strings.ReplaceAll(result, "[DETECTED_ENTRY_POINTS]", strings.Join(ctx.EntryPoints, ", "))

	// Format flows as "- Name: Description" lines.
	flowLines := make([]string, 0, len(ctx.Flows))
	for _, f := range ctx.Flows {
		flowLines = append(flowLines, fmt.Sprintf("- %s: %s", f.Name, f.Description))
	}
	result = strings.ReplaceAll(result, "[DETECTED_FLOWS]", strings.Join(flowLines, "\n"))

	// Memory area classification.
	var stableAreas, volatileAreas []string
	if mem != nil && mem.Areas != nil {
		for name, area := range mem.Areas {
			if area == nil {
				continue
			}
			switch {
			case area.Confidence >= 0.95:
				stableAreas = append(stableAreas, name)
			case area.Confidence < 0.75:
				volatileAreas = append(volatileAreas, name)
			}
		}
	}

	stableStr := "none"
	if len(stableAreas) > 0 {
		stableStr = strings.Join(stableAreas, ", ")
	}
	volatileStr := "none"
	if len(volatileAreas) > 0 {
		volatileStr = strings.Join(volatileAreas, ", ")
	}

	result = strings.ReplaceAll(result, "[MEMORY_STABLE_AREAS]", stableStr)
	result = strings.ReplaceAll(result, "[MEMORY_VOLATILE_AREAS]", volatileStr)

	// Scenario count derived from depth × flow-complexity table.
	n := scenarioCount(opts.Depth, len(ctx.Flows))
	result = strings.ReplaceAll(result, "[N]", fmt.Sprintf("%d", n))

	return result, nil
}
