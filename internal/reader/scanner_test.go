package reader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shaifulshabuj/teststop/internal/reader"
)

func TestScan_basicDirectory(t *testing.T) {
	// Create a temp dir with some files
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\nfunc main() {}"), 0644)
	os.WriteFile(filepath.Join(tmp, "server.go"), []byte("package main"), 0644)
	os.MkdirAll(filepath.Join(tmp, "node_modules", "pkg"), 0755)
	os.WriteFile(filepath.Join(tmp, "node_modules", "pkg", "index.js"), []byte(""), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.FileCount < 2 {
		t.Errorf("expected >= 2 files, got %d", ctx.FileCount)
	}
	// node_modules should be skipped
	if ctx.Languages["JavaScript"] > 0 {
		t.Error("node_modules should be skipped, no JS files should be counted")
	}
	if ctx.Languages["Go"] < 2 {
		t.Errorf("expected >= 2 Go files, got %d", ctx.Languages["Go"])
	}
}

func TestDetect_goAPI(t *testing.T) {
	tmp := t.TempDir()
	// Write a Go file with HTTP listener
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte(`package main
import "net/http"
func main() {
	http.HandleFunc("/health", healthHandler)
	http.ListenAndServe(":8080", nil)
}`), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Detect(&ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.Language != "Go" {
		t.Errorf("expected Go, got %s", ctx.Language)
	}
	if ctx.Type != "api" {
		t.Errorf("expected api, got %s", ctx.Type)
	}
}

func TestAnalyze_extractsRoutes(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "routes.go"), []byte(`package main
import "net/http"
func setup() {
	http.HandleFunc("/api/users", usersHandler)
	http.HandleFunc("/api/health", healthHandler)
}`), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Language = "Go"
	if err := reader.Analyze(&ctx); err != nil {
		t.Fatal(err)
	}
	if len(ctx.Flows) < 2 {
		t.Errorf("expected >= 2 flows, got %d: %v", len(ctx.Flows), ctx.Flows)
	}
}

func TestScan_nameAndPath(t *testing.T) {
	tmp := t.TempDir()

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Name == "" {
		t.Error("Name should not be empty")
	}
	if ctx.Path == "" {
		t.Error("Path should not be empty")
	}
	if ctx.Languages == nil {
		t.Error("Languages map should be initialized")
	}
}

func TestScan_skipsHiddenDirs(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".hidden"), 0755)
	os.WriteFile(filepath.Join(tmp, ".hidden", "secret.go"), []byte("package secret"), 0644)
	os.WriteFile(filepath.Join(tmp, "visible.go"), []byte("package main"), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	// Only visible.go should be counted
	if ctx.Languages["Go"] != 1 {
		t.Errorf("expected 1 Go file (hidden dir skipped), got %d", ctx.Languages["Go"])
	}
}

func TestDetect_primaryLanguage_tieBreak(t *testing.T) {
	tmp := t.TempDir()
	// Equal number of Go and Python files — Go should win
	os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmp, "b.py"), []byte("print()"), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Detect(&ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.Language != "Go" {
		t.Errorf("expected Go to win tie-break, got %s", ctx.Language)
	}
}

func TestDetect_webApp(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "index.html"), []byte("<html></html>"), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Detect(&ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.Type != "web_app" {
		t.Errorf("expected web_app, got %s", ctx.Type)
	}
}

func TestDetect_library(t *testing.T) {
	tmp := t.TempDir()
	// Write a file but no entry points
	os.WriteFile(filepath.Join(tmp, "lib.go"), []byte("package mylib\nfunc Foo() {}"), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	// lib.go is not an entry point, EntryPoints will be empty
	if err := reader.Detect(&ctx); err != nil {
		t.Fatal(err)
	}
	if ctx.Type != "library" {
		t.Errorf("expected library, got %s", ctx.Type)
	}
}

func TestAnalyze_cobraCommands(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "root.go"), []byte(`package cmd
import "github.com/spf13/cobra"
var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "My app",
}
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run stuff",
}
`), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Language = "Go"
	if err := reader.Analyze(&ctx); err != nil {
		t.Fatal(err)
	}
	// Should detect at least one cobra command
	hasCLI := false
	for _, f := range ctx.Flows {
		if f.Area == "cli" {
			hasCLI = true
			break
		}
	}
	if !hasCLI {
		t.Errorf("expected cobra CLI flows, got %v", ctx.Flows)
	}
}

func TestAnalyze_goModDeps(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(`module example.com/myapp

go 1.21

require (
	github.com/spf13/cobra v1.8.0
	github.com/gin-gonic/gin v1.9.1
	golang.org/x/text v0.14.0
)
`), 0644)

	ctx, err := reader.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Analyze(&ctx); err != nil {
		t.Fatal(err)
	}
	if len(ctx.Dependencies) == 0 {
		t.Error("expected dependencies from go.mod, got none")
	}
}

func TestScanProject_fullPipeline(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte(`package main
import (
	"net/http"
	"github.com/spf13/cobra"
)
func main() {
	http.HandleFunc("/api/users", usersHandler)
	http.ListenAndServe(":8080", nil)
}
`), 0644)
	os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(`module example.com/test
go 1.21
require github.com/spf13/cobra v1.8.0
`), 0644)

	ctx, err := reader.ScanProject(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Language != "Go" {
		t.Errorf("expected Go, got %s", ctx.Language)
	}
	if ctx.Type != "api" {
		t.Errorf("expected api, got %s", ctx.Type)
	}
	if len(ctx.Flows) == 0 {
		t.Error("expected flows to be extracted")
	}
}
