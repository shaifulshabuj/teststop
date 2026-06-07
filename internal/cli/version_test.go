package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveVersion_PrefersInjected(t *testing.T) {
	if got := resolveVersion("v1.2.3"); got != "v1.2.3" {
		t.Errorf("resolveVersion(\"v1.2.3\") = %q, want %q", got, "v1.2.3")
	}
}

func TestResolveVersion_FallsBackWhenDev(t *testing.T) {
	// With "dev" injected, resolveVersion consults build info; in a test binary
	// that may or may not be present, but it must always return non-empty.
	if got := resolveVersion("dev"); got == "" {
		t.Error("resolveVersion(\"dev\") returned empty string")
	}
}

func TestVersionCommand_PrintsBuildInfo(t *testing.T) {
	buildVersion, buildCommit, buildDate = "v9.9.9", "deadbeef", "2026-06-07"

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)
	versionCmd.Run(versionCmd, nil)

	out := buf.String()
	for _, want := range []string{"teststop v9.9.9", "deadbeef", "2026-06-07", "os/arch"} {
		if !strings.Contains(out, want) {
			t.Errorf("version output missing %q\ngot:\n%s", want, out)
		}
	}
}
