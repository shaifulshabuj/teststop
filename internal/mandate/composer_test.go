package mandate

import (
	"strings"
	"testing"

	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
)

func TestCompose_SubstitutesPlaceholders(t *testing.T) {
	out, err := Compose(Input{
		Project: &reader.ProjectContext{
			Name:        "demo",
			Language:    "Go",
			Type:        "cli",
			EntryPoints: []string{"cmd/demo/main.go"},
			KeyFlows: []reader.Flow{
				{Kind: "cli", Name: "demo run", Location: "cmd/demo/main.go:42"},
			},
			Complexity: "simple",
			FileCount:  17,
		},
		Memory: memory.New(),
		Depth:  DepthNormal,
	})
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	for _, want := range []string{
		"demo", "Go", "cli", "cmd/demo/main.go", "demo run",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
	for _, leftover := range []string{
		"{{PROJECT_NAME}}", "{{DETECTED_LANGUAGE}}",
		"{{DETECTED_TYPE}}", "{{DETECTED_ENTRY_POINTS}}",
		"{{DETECTED_FLOWS}}", "{{MEMORY_STABLE_AREAS}}",
		"{{MEMORY_VOLATILE_AREAS}}", "{{SCENARIO_COUNT}}",
	} {
		if strings.Contains(out, leftover) {
			t.Errorf("unresolved placeholder %s", leftover)
		}
	}
}

func TestScenarioCount_ScalesWithComplexityAndDepth(t *testing.T) {
	simple := scenarioCount(Input{
		Project: &reader.ProjectContext{Complexity: "simple"},
		Memory:  memory.New(),
		Depth:   DepthLight,
	})
	complex := scenarioCount(Input{
		Project: &reader.ProjectContext{Complexity: "complex"},
		Memory:  memory.New(),
		Depth:   DepthAggressive,
	})
	if simple >= complex {
		t.Fatalf("light/simple (%d) should be < aggressive/complex (%d)", simple, complex)
	}
	if simple < 3 {
		t.Fatalf("scenario count floor violated: %d", simple)
	}
	if complex > 50 {
		t.Fatalf("scenario count ceiling violated: %d", complex)
	}
}

func TestScenarioCount_MatureStageReducesBudget(t *testing.T) {
	mNew := memory.New()
	mMature := memory.New()
	mMature.MaturityStage = memory.StageMature

	new := scenarioCount(Input{
		Project: &reader.ProjectContext{Complexity: "moderate"},
		Memory:  mNew,
		Depth:   DepthNormal,
	})
	mature := scenarioCount(Input{
		Project: &reader.ProjectContext{Complexity: "moderate"},
		Memory:  mMature,
		Depth:   DepthNormal,
	})
	if mature >= new {
		t.Fatalf("mature stage should reduce budget: new=%d mature=%d", new, mature)
	}
}
