package reader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// flowSignal pairs a regex with the Flow.Kind it produces. The patterns
// are intentionally permissive: a real-user-facing surface is enough,
// false positives are cheaper than false negatives because the mandate
// only needs orientation, not a complete index.
type flowSignal struct {
	kind    string
	pattern *regexp.Regexp
	name    func(match []string) string
}

var flowSignals = []flowSignal{
	// Go HTTP routers
	{
		kind:    "http",
		pattern: regexp.MustCompile(`(?i)\b(?:router|r|mux|app|e|g)\.(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|Handle|HandleFunc)\(\s*"([^"]+)"`),
		name: func(m []string) string {
			method := strings.ToUpper(m[1])
			if method == "HANDLE" || method == "HANDLEFUNC" {
				method = "ANY"
			}
			return method + " " + m[2]
		},
	},
	// Node/Express style
	{
		kind:    "http",
		pattern: regexp.MustCompile(`\b(?:app|router)\.(get|post|put|patch|delete|head|options|all)\(\s*['"]([^'"]+)['"]`),
		name: func(m []string) string {
			return strings.ToUpper(m[1]) + " " + m[2]
		},
	},
	// FastAPI / Flask decorators
	{
		kind:    "http",
		pattern: regexp.MustCompile(`@\w+\.(get|post|put|patch|delete|head|options|route)\(\s*['"]([^'"]+)['"]`),
		name: func(m []string) string {
			return strings.ToUpper(m[1]) + " " + m[2]
		},
	},
	// Django urls.py path(...)
	{
		kind:    "http",
		pattern: regexp.MustCompile(`\bpath\(\s*['"]([^'"]*)['"]\s*,`),
		name: func(m []string) string {
			return "GET /" + strings.TrimPrefix(m[1], "/")
		},
	},
	// Cobra CLI command definitions
	{
		kind:    "cli",
		pattern: regexp.MustCompile(`Use:\s*"([^"\s]+)"`),
		name: func(m []string) string {
			return m[1]
		},
	},
	// Click / Typer CLI decorators
	{
		kind:    "cli",
		pattern: regexp.MustCompile(`@(?:click|typer|app)\.command\(\s*(?:['"]([^'"]+)['"])?`),
		name: func(m []string) string {
			if m[1] != "" {
				return m[1]
			}
			return "<unnamed command>"
		},
	},
}

// extsToScan lists source extensions worth grepping for flow signals.
// Other files (assets, configs) rarely carry route definitions.
var extsToScan = map[string]struct{}{
	".go":   {},
	".js":   {},
	".mjs":  {},
	".cjs":  {},
	".jsx":  {},
	".ts":   {},
	".tsx":  {},
	".py":   {},
	".rb":   {},
	".java": {},
	".kt":   {},
	".php":  {},
	".rs":   {},
}

// detectEntryPoints lists the conventional entry points found in the
// project. The list is bounded so the mandate stays compact.
func detectEntryPoints(r *scanResult, language string) []string {
	candidates := map[string]bool{}
	for _, f := range r.files {
		base := filepath.Base(f)
		rel := relativeTo(f, r.root)
		switch base {
		case "main.go", "main.py", "app.py", "manage.py", "__main__.py",
			"index.js", "index.ts", "server.js", "server.ts", "app.js", "app.ts":
			candidates[rel] = true
		case "Main.java", "Application.java", "main.rs":
			candidates[rel] = true
		}
		// cmd/<name>/main.go convention
		if strings.HasPrefix(rel, "cmd"+string(os.PathSeparator)) &&
			(base == "main.go" || strings.HasSuffix(rel, ".go") && filepath.Dir(rel) != "cmd") {
			candidates[rel] = true
		}
		// src/main.rs convention
		if rel == filepath.Join("src", "main.rs") {
			candidates[rel] = true
		}
	}
	out := make([]string, 0, len(candidates))
	for k := range candidates {
		out = append(out, k)
	}
	return capStrings(out, 12)
}

// detectFlows greps a bounded slice of source files for route and
// command patterns. It is best-effort, not exhaustive — the goal is
// to give the mandate enough orientation that the AI can reason about
// the surface area the user actually touches.
func detectFlows(r *scanResult) []Flow {
	const maxFiles = 200
	const maxBytesPerFile = 1 << 20 // 1 MB
	seen := map[string]bool{}
	flows := []Flow{}

	scanned := 0
	for _, f := range r.files {
		ext := strings.ToLower(filepath.Ext(f))
		if _, ok := extsToScan[ext]; !ok {
			continue
		}
		if scanned >= maxFiles {
			break
		}
		scanned++

		file, err := os.Open(f)
		if err != nil {
			continue
		}
		// Tight read budget per file keeps the reader cheap on large
		// monorepos.
		buf := make([]byte, maxBytesPerFile)
		n, _ := file.Read(buf)
		file.Close()
		if n == 0 {
			continue
		}
		content := string(buf[:n])

		rel := relativeTo(f, r.root)
		// Walk lines so the location includes a line number, which
		// makes the mandate concrete enough for the AI to reference.
		scanner := bufio.NewScanner(strings.NewReader(content))
		scanner.Buffer(make([]byte, 0, 1<<16), 1<<20)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			for _, sig := range flowSignals {
				m := sig.pattern.FindStringSubmatch(line)
				if m == nil {
					continue
				}
				name := sig.name(m)
				key := sig.kind + "|" + name
				if seen[key] {
					continue
				}
				seen[key] = true
				flows = append(flows, Flow{
					Kind:     sig.kind,
					Name:     name,
					Location: fmt.Sprintf("%s:%d", rel, lineNo),
				})
				if len(flows) >= 40 {
					return flows
				}
			}
		}
	}
	return flows
}

// computeComplexity is a deliberately crude signal: small/medium/large
// rather than any meaningful cyclomatic measure. The mandate uses it
// to pick a scenario budget, nothing more.
func computeComplexity(fileCount int) string {
	switch {
	case fileCount < 50:
		return "simple"
	case fileCount < 500:
		return "moderate"
	default:
		return "complex"
	}
}

// extractDependencies pulls a flat, deduplicated list of top-level
// dependency identifiers from common manifest files. We avoid parsing
// version strings — the mandate only needs the names.
func extractDependencies(r *scanResult) []string {
	deps := map[string]struct{}{}
	for _, m := range r.manifestFiles {
		base := filepath.Base(m)
		data, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		text := string(data)
		switch base {
		case "go.mod":
			re := regexp.MustCompile(`(?m)^\s*([\w./\-]+/[\w./\-]+)\s+v[\w.\-]+`)
			for _, m := range re.FindAllStringSubmatch(text, -1) {
				deps[m[1]] = struct{}{}
			}
		case "package.json":
			re := regexp.MustCompile(`"([@\w./\-]+)"\s*:\s*"[\^~>=<\d. xX\-*]+"`)
			for _, m := range re.FindAllStringSubmatch(text, -1) {
				deps[m[1]] = struct{}{}
			}
		case "requirements.txt":
			for _, line := range strings.Split(text, "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				name := strings.FieldsFunc(line, func(r rune) bool {
					return r == '=' || r == '<' || r == '>' || r == '!' || r == ';' || r == ' '
				})
				if len(name) > 0 && name[0] != "" {
					deps[name[0]] = struct{}{}
				}
			}
		case "Cargo.toml":
			re := regexp.MustCompile(`(?m)^\s*([\w\-]+)\s*=\s*["{]`)
			for _, m := range re.FindAllStringSubmatch(text, -1) {
				if m[1] != "name" && m[1] != "version" && m[1] != "edition" &&
					m[1] != "authors" && m[1] != "description" {
					deps[m[1]] = struct{}{}
				}
			}
		}
	}
	out := make([]string, 0, len(deps))
	for k := range deps {
		out = append(out, k)
	}
	return capStrings(out, 30)
}

func relativeTo(p, root string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return rel
}

func capStrings(in []string, max int) []string {
	if len(in) <= max {
		return in
	}
	return in[:max]
}

// Read produces a ProjectContext for the given path. It is the single
// entry point external packages should call.
func Read(path string) (*ProjectContext, error) {
	r, err := scan(path)
	if err != nil {
		return nil, fmt.Errorf("scan project: %w", err)
	}

	lang := detectLanguage(r)
	pType := detectType(r, lang)

	ctx := &ProjectContext{
		Path:         r.root,
		Name:         projectName(r.root),
		Language:     lang,
		Type:         pType,
		EntryPoints:  detectEntryPoints(r, lang),
		KeyFlows:     detectFlows(r),
		Dependencies: extractDependencies(r),
		TestFiles:    capStrings(relativizeAll(r.testFiles, r.root), 50),
		FileCount:    len(r.files),
		Complexity:   computeComplexity(len(r.files)),
	}
	return ctx, nil
}

func relativizeAll(paths []string, root string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, relativeTo(p, root))
	}
	return out
}
