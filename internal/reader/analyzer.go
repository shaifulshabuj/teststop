package reader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxFlows       = 50
	maxDeps        = 20
	maxAnalyzeSize = 500 * 1024 // 500 KB
)

// analyzeSourceExts are the extensions Analyze inspects for flow patterns.
var analyzeSourceExts = map[string]bool{
	".go": true,
	".py": true,
	".js": true,
	".ts": true,
	".rb": true,
}

// --- Go HTTP route patterns ---
var (
	reGoHandleFunc = regexp.MustCompile(`http\.HandleFunc\("([^"]+)"`)
	reGoRouter     = regexp.MustCompile(`(?:router|e|r|mux)\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"`)
	// Also catch cases where the variable name differs but the method is on a *mux.Router / gin etc.
	reGoMethod = regexp.MustCompile(`\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"`)
)

// --- Python route patterns ---
var (
	rePyAppRoute    = regexp.MustCompile(`@(?:app|router)\.(get|post|put|delete|patch)\("([^"]+)"`)
	rePyAppRouteSQ  = regexp.MustCompile(`@(?:app|router)\.(get|post|put|delete|patch)\('([^']+)'`)
)

// --- JavaScript/TypeScript route patterns ---
var (
	reJSAppRoute   = regexp.MustCompile(`(?:app|router)\.(get|post|put|delete|patch)\(['"]([^'"]+)['"]`)
)

// --- Cobra CLI patterns ---
var (
	reCobraUse = regexp.MustCompile(`Use:\s+"([^"]+)"`)
)

// --- Python argparse/click patterns ---
var (
	reArgparseAdd = regexp.MustCompile(`add_argument\(['"]([^'"]+)['"]`)
)

// Analyze enriches ctx.Flows and ctx.Dependencies by extracting patterns from source files.
func Analyze(ctx *ProjectContext) error {
	seen := make(map[string]bool) // deduplicate by "name|area"

	err := filepath.Walk(ctx.Path, func(p string, info os.FileInfo, werr error) error {
		if werr != nil {
			return nil
		}

		if info.IsDir() {
			base := filepath.Base(p)
			if p != ctx.Path && (skipDirs[base] || strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		// Check size limit
		if info.Size() > maxAnalyzeSize {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		base := filepath.Base(p)

		// Dependency extraction from config files
		extractDeps(ctx, p, base)

		// Flow extraction from source files
		if !analyzeSourceExts[ext] {
			return nil
		}

		if len(ctx.Flows) >= maxFlows {
			return nil
		}

		flows := extractFlows(p, ext)
		for _, f := range flows {
			key := f.Name + "|" + f.Area
			if !seen[key] {
				seen[key] = true
				ctx.Flows = append(ctx.Flows, f)
				if len(ctx.Flows) >= maxFlows {
					break
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("analyze walk: %w", err)
	}

	return nil
}

// extractFlows scans a single file and returns detected flows.
func extractFlows(path, ext string) []Flow {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var flows []Flow
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		switch ext {
		case ".go":
			flows = append(flows, extractGoFlows(line)...)
		case ".py":
			flows = append(flows, extractPyFlows(line)...)
		case ".js", ".ts", ".mjs", ".cjs":
			flows = append(flows, extractJSFlows(line)...)
		}
	}

	return flows
}

func extractGoFlows(line string) []Flow {
	var flows []Flow

	// http.HandleFunc("/path", ...)
	if m := reGoHandleFunc.FindStringSubmatch(line); m != nil {
		path := m[1]
		flows = append(flows, Flow{
			Name:        "HTTP " + path,
			Description: "HTTP handler for " + path,
			Area:        "api" + path,
		})
	}

	// router/e/r/mux.METHOD("/path", ...)
	if m := reGoMethod.FindStringSubmatch(line); m != nil {
		method := strings.ToUpper(m[1])
		path := m[2]
		flows = append(flows, Flow{
			Name:        method + " " + path,
			Description: method + " handler for " + path,
			Area:        "api" + path,
		})
	}

	// Use: "command-name" (cobra)
	if m := reCobraUse.FindStringSubmatch(line); m != nil {
		use := m[1]
		// Only capture single-word or hyphenated names (not full cobra help text)
		if !strings.Contains(use, "\n") {
			flows = append(flows, Flow{
				Name:        "cmd: " + use,
				Description: "CLI command: " + use,
				Area:        "cli",
			})
		}
	}

	return flows
}

func extractPyFlows(line string) []Flow {
	var flows []Flow

	if m := rePyAppRoute.FindStringSubmatch(line); m != nil {
		method := strings.ToUpper(m[1])
		path := m[2]
		flows = append(flows, Flow{
			Name:        method + " " + path,
			Description: method + " handler for " + path,
			Area:        "api" + path,
		})
	}

	if m := rePyAppRouteSQ.FindStringSubmatch(line); m != nil {
		method := strings.ToUpper(m[1])
		path := m[2]
		flows = append(flows, Flow{
			Name:        method + " " + path,
			Description: method + " handler for " + path,
			Area:        "api" + path,
		})
	}

	if m := reArgparseAdd.FindStringSubmatch(line); m != nil {
		arg := m[1]
		flows = append(flows, Flow{
			Name:        "arg: " + arg,
			Description: "CLI argument: " + arg,
			Area:        "cli",
		})
	}

	return flows
}

func extractJSFlows(line string) []Flow {
	var flows []Flow

	if m := reJSAppRoute.FindStringSubmatch(line); m != nil {
		method := strings.ToUpper(m[1])
		path := m[2]
		flows = append(flows, Flow{
			Name:        method + " " + path,
			Description: method + " handler for " + path,
			Area:        "api" + path,
		})
	}

	return flows
}

// extractDeps extracts dependencies from known config files.
func extractDeps(ctx *ProjectContext, path, base string) {
	if len(ctx.Dependencies) >= maxDeps {
		return
	}

	switch base {
	case "go.mod":
		extractGoModDeps(ctx, path)
	case "package.json":
		extractPackageJSONDeps(ctx, path)
	case "requirements.txt":
		extractRequirementsDeps(ctx, path)
	case "Gemfile":
		extractGemfileDeps(ctx, path)
	case "Cargo.toml":
		extractCargoTomlDeps(ctx, path)
	case "pom.xml":
		extractPomXMLDeps(ctx, path)
	}
}

func extractGoModDeps(ctx *ProjectContext, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "github.com/") || strings.HasPrefix(line, "golang.org/") {
			// Extract just the module path, drop version
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				addDep(ctx, parts[0])
			}
		}
	}
}

func extractPackageJSONDeps(ctx *ProjectContext, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	for name := range pkg.Dependencies {
		addDep(ctx, name)
	}
	for name := range pkg.DevDependencies {
		addDep(ctx, name)
	}
}

func extractRequirementsDeps(ctx *ProjectContext, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip version specifiers: pkg==1.0, pkg>=1.0, pkg~=1.0
		for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">"} {
			if idx := strings.Index(line, sep); idx != -1 {
				line = line[:idx]
				break
			}
		}
		line = strings.TrimSpace(line)
		if line != "" {
			addDep(ctx, line)
		}
	}
}

func extractGemfileDeps(ctx *ProjectContext, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	reGem := regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"]`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := reGem.FindStringSubmatch(scanner.Text()); m != nil {
			addDep(ctx, m[1])
		}
	}
}

func extractCargoTomlDeps(ctx *ProjectContext, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	inDeps := false
	reDep := regexp.MustCompile(`^([a-zA-Z0-9_-]+)\s*=`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[dependencies]" || line == "[dev-dependencies]" {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}
		if inDeps {
			if m := reDep.FindStringSubmatch(line); m != nil {
				addDep(ctx, m[1])
			}
		}
	}
}

func extractPomXMLDeps(ctx *ProjectContext, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// Simple regex extraction of <artifactId> within <dependency> blocks
	reArtifact := regexp.MustCompile(`<artifactId>([^<]+)</artifactId>`)
	matches := reArtifact.FindAllStringSubmatch(string(data), -1)
	for _, m := range matches {
		addDep(ctx, m[1])
	}
}

// addDep appends dep to ctx.Dependencies if not already present and under the limit.
func addDep(ctx *ProjectContext, dep string) {
	if len(ctx.Dependencies) >= maxDeps {
		return
	}
	for _, existing := range ctx.Dependencies {
		if existing == dep {
			return
		}
	}
	ctx.Dependencies = append(ctx.Dependencies, dep)
}
