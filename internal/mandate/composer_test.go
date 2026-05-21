package mandate_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shaifulshabuj/teststop/internal/mandate"
	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
)

func TestCompose_replacesPlaceholders(t *testing.T) {
	ctx := reader.ProjectContext{
		Name:        "myapp",
		Language:    "Go",
		Type:        "api",
		EntryPoints: []string{"cmd/server/main.go"},
		Flows:       []reader.Flow{{Name: "auth", Description: "User authentication", Area: "auth"}},
	}
	mem := memory.NewMemory()
	opts := mandate.Options{Depth: "normal"}

	result, err := mandate.Compose(ctx, mem, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "[SYSTEM_NAME]") {
		t.Error("placeholder [SYSTEM_NAME] was not replaced")
	}
	if strings.Contains(result, "[DETECTED_LANGUAGE]") {
		t.Error("placeholder [DETECTED_LANGUAGE] was not replaced")
	}
	if !strings.Contains(result, "myapp") {
		t.Error("expected project name 'myapp' in result")
	}
	if !strings.Contains(result, "Go") {
		t.Error("expected language 'Go' in result")
	}
}

func TestCompose_scenarioCount(t *testing.T) {
	ctx := reader.ProjectContext{Name: "app", Language: "Go", Type: "cli"}
	mem := memory.NewMemory()

	cases := []struct {
		depth string
		flows int
		n     int
	}{
		{"light", 0, 5},
		{"normal", 0, 10},
		{"aggressive", 0, 20},
		{"normal", 5, 20},
		{"aggressive", 10, 50},
	}
	for _, c := range cases {
		ctx.Flows = make([]reader.Flow, c.flows)
		result, err := mandate.Compose(ctx, mem, mandate.Options{Depth: c.depth})
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("%d", c.n)
		if !strings.Contains(result, expected) {
			t.Errorf("depth=%s flows=%d: expected N=%s in result", c.depth, c.flows, expected)
		}
	}
}

func TestCompose_memoryAreas(t *testing.T) {
	ctx := reader.ProjectContext{Name: "app", Language: "Go", Type: "api"}
	mem := memory.NewMemory()

	// Add a stable area (confidence >= 0.95).
	mem.Areas["auth"] = &memory.Area{Confidence: 0.97}
	// Add a volatile area (confidence < 0.75).
	mem.Areas["payments"] = &memory.Area{Confidence: 0.50}
	// Add a middle area (0.75 <= confidence < 0.95) — should not appear in either list.
	mem.Areas["profile"] = &memory.Area{Confidence: 0.80}

	opts := mandate.Options{Depth: "normal"}
	result, err := mandate.Compose(ctx, mem, opts)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(result, "[MEMORY_STABLE_AREAS]") {
		t.Error("placeholder [MEMORY_STABLE_AREAS] was not replaced")
	}
	if strings.Contains(result, "[MEMORY_VOLATILE_AREAS]") {
		t.Error("placeholder [MEMORY_VOLATILE_AREAS] was not replaced")
	}
	if !strings.Contains(result, "auth") {
		t.Error("expected stable area 'auth' in result")
	}
	if !strings.Contains(result, "payments") {
		t.Error("expected volatile area 'payments' in result")
	}
}

func TestCompose_nilMemory(t *testing.T) {
	ctx := reader.ProjectContext{Name: "app", Language: "Python", Type: "web_app"}
	opts := mandate.Options{Depth: "light"}

	result, err := mandate.Compose(ctx, nil, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "[MEMORY_STABLE_AREAS]") {
		t.Error("placeholder [MEMORY_STABLE_AREAS] was not replaced when mem is nil")
	}
	if strings.Contains(result, "[MEMORY_VOLATILE_AREAS]") {
		t.Error("placeholder [MEMORY_VOLATILE_AREAS] was not replaced when mem is nil")
	}
}

func TestCompose_flowFormatting(t *testing.T) {
	ctx := reader.ProjectContext{
		Name:     "svc",
		Language: "Go",
		Type:     "api",
		Flows: []reader.Flow{
			{Name: "login", Description: "User login flow", Area: "auth"},
			{Name: "checkout", Description: "Purchase flow", Area: "payments"},
		},
	}
	mem := memory.NewMemory()
	opts := mandate.Options{Depth: "normal"}

	result, err := mandate.Compose(ctx, mem, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "[DETECTED_FLOWS]") {
		t.Error("placeholder [DETECTED_FLOWS] was not replaced")
	}
	if !strings.Contains(result, "- login: User login flow") {
		t.Error("expected formatted flow '- login: User login flow' in result")
	}
	if !strings.Contains(result, "- checkout: Purchase flow") {
		t.Error("expected formatted flow '- checkout: Purchase flow' in result")
	}
}
