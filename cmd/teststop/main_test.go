package main

import "testing"

// TestBuildDefaults verifies the package-level build-metadata vars have their
// documented default values (injected at release time; stay as defaults in tests).
func TestBuildDefaults(t *testing.T) {
	if version != "dev" {
		t.Errorf("version: want default %q, got %q", "dev", version)
	}
	if commit != "none" {
		t.Errorf("commit: want default %q, got %q", "none", commit)
	}
	if date != "unknown" {
		t.Errorf("date: want default %q, got %q", "unknown", date)
	}
}
