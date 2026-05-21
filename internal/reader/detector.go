package reader

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// extLanguage maps a file extension (with leading dot) to a coarse
// language label. The label is used downstream by the mandate so the
// AI can tune its scenarios; it does not have to be perfectly precise.
var extLanguage = map[string]string{
	".go":    "Go",
	".js":    "JavaScript",
	".mjs":   "JavaScript",
	".cjs":   "JavaScript",
	".jsx":   "JavaScript",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".py":    "Python",
	".rb":    "Ruby",
	".rs":    "Rust",
	".java":  "Java",
	".kt":    "Kotlin",
	".kts":   "Kotlin",
	".swift": "Swift",
	".php":   "PHP",
	".cs":    "C#",
	".fs":    "F#",
	".scala": "Scala",
	".ex":    "Elixir",
	".exs":   "Elixir",
	".dart":  "Dart",
	".clj":   "Clojure",
	".elm":   "Elm",
	".hs":    "Haskell",
	".ml":    "OCaml",
	".c":     "C",
	".h":     "C",
	".cpp":   "C++",
	".cc":    "C++",
	".hpp":   "C++",
	".m":     "Objective-C",
	".mm":    "Objective-C++",
	".sh":    "Shell",
	".bash":  "Shell",
	".zsh":   "Shell",
	".lua":   "Lua",
}

// manifestLanguage gives a strong language signal from a manifest file
// when extension counts are noisy or ambiguous.
var manifestLanguage = map[string]string{
	"go.mod":           "Go",
	"package.json":     "JavaScript",
	"tsconfig.json":    "TypeScript",
	"Cargo.toml":       "Rust",
	"pyproject.toml":   "Python",
	"requirements.txt": "Python",
	"Pipfile":          "Python",
	"pom.xml":          "Java",
	"build.gradle":     "Java",
	"build.gradle.kts": "Kotlin",
	"Gemfile":          "Ruby",
	"composer.json":    "PHP",
	"mix.exs":          "Elixir",
	"pubspec.yaml":     "Dart",
	"deno.json":        "TypeScript",
}

// detectLanguage picks the dominant language from extension counts,
// preferring manifest signals when they disagree with raw counts.
//
// The TypeScript / JavaScript distinction is a known special case:
// a TS project commonly contains many .js artifacts in build output
// (already filtered by ignoredDirs) and many .ts source files. We
// pick whichever extension family has more source files.
func detectLanguage(r *scanResult) string {
	manifestVote := ""
	for _, m := range r.manifestFiles {
		base := filepath.Base(m)
		if lang, ok := manifestLanguage[base]; ok {
			// tsconfig.json + package.json → prefer TypeScript
			if lang == "TypeScript" {
				return "TypeScript"
			}
			if manifestVote == "" {
				manifestVote = lang
			}
		}
	}

	// Aggregate counts per language for source-of-truth ranking.
	langCounts := map[string]int{}
	for ext, n := range r.extCounts {
		if lang, ok := extLanguage[ext]; ok {
			langCounts[lang] += n
		}
	}

	if len(langCounts) == 0 {
		if manifestVote != "" {
			return manifestVote
		}
		return "unknown"
	}

	// Pick the language with the most source files; tiebreak alphabetically
	// for determinism.
	type kv struct {
		lang string
		n    int
	}
	ranked := make([]kv, 0, len(langCounts))
	for k, v := range langCounts {
		ranked = append(ranked, kv{k, v})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].n != ranked[j].n {
			return ranked[i].n > ranked[j].n
		}
		return ranked[i].lang < ranked[j].lang
	})

	top := ranked[0].lang

	// Manifest signal beats a close extension race so that, e.g., a Go
	// project with many JS docs/assets is still reported as Go.
	if manifestVote != "" && manifestVote != top {
		// If the top extension count is not at least 2x the manifest's
		// own language, respect the manifest.
		manifestN := langCounts[manifestVote]
		if ranked[0].n < manifestN*2 {
			return manifestVote
		}
	}
	return top
}

// detectType infers the project archetype using filesystem cues only.
// This is intentionally cheap and approximate; the mandate accepts a
// rough signal and the AI fills in gaps.
func detectType(r *scanResult, language string) string {
	hasMainGo := false
	hasCmdDir := false
	hasMainPy := false
	hasIndexHTML := false
	hasFrontendDeps := false
	hasServerDeps := false
	hasManagePy := false
	hasLibrarySignals := false
	hasCLISignals := false

	for _, f := range r.files {
		base := filepath.Base(f)
		lower := strings.ToLower(base)
		rel := strings.TrimPrefix(f, r.root+string(os.PathSeparator))
		switch base {
		case "main.go":
			hasMainGo = true
		case "manage.py":
			hasManagePy = true
		case "main.py", "app.py", "__main__.py":
			hasMainPy = true
		}
		if lower == "index.html" {
			hasIndexHTML = true
		}
		if strings.HasPrefix(rel, "cmd"+string(os.PathSeparator)) {
			hasCmdDir = true
		}
	}

	// Sniff a handful of manifest contents for stronger signal.
	for _, m := range r.manifestFiles {
		data, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		content := strings.ToLower(string(data))
		base := filepath.Base(m)
		switch base {
		case "package.json":
			if strings.Contains(content, "\"react\"") ||
				strings.Contains(content, "\"vue\"") ||
				strings.Contains(content, "\"svelte\"") ||
				strings.Contains(content, "\"next\"") ||
				strings.Contains(content, "\"nuxt\"") ||
				strings.Contains(content, "\"vite\"") {
				hasFrontendDeps = true
			}
			if strings.Contains(content, "\"express\"") ||
				strings.Contains(content, "\"fastify\"") ||
				strings.Contains(content, "\"koa\"") ||
				strings.Contains(content, "\"hapi\"") ||
				strings.Contains(content, "\"@nestjs/core\"") {
				hasServerDeps = true
			}
			if strings.Contains(content, "\"main\":") && !strings.Contains(content, "\"bin\":") &&
				!hasFrontendDeps && !hasServerDeps {
				hasLibrarySignals = true
			}
			if strings.Contains(content, "\"bin\":") || strings.Contains(content, "\"oclif\"") ||
				strings.Contains(content, "\"commander\"") || strings.Contains(content, "\"yargs\"") {
				hasCLISignals = true
			}
		case "pyproject.toml", "requirements.txt", "Pipfile":
			if strings.Contains(content, "django") || strings.Contains(content, "flask") ||
				strings.Contains(content, "fastapi") || strings.Contains(content, "starlette") {
				hasServerDeps = true
			}
			if strings.Contains(content, "click") || strings.Contains(content, "typer") ||
				strings.Contains(content, "argparse") {
				hasCLISignals = true
			}
		case "go.mod":
			if strings.Contains(content, "cobra") || strings.Contains(content, "urfave/cli") ||
				strings.Contains(content, "kong") {
				hasCLISignals = true
			}
			if strings.Contains(content, "gin-gonic") || strings.Contains(content, "echo") ||
				strings.Contains(content, "fiber") || strings.Contains(content, "chi") ||
				strings.Contains(content, "gorilla/mux") {
				hasServerDeps = true
			}
		}
	}

	switch {
	case hasIndexHTML && hasFrontendDeps && !hasServerDeps:
		return "web_app"
	case hasFrontendDeps && hasServerDeps:
		return "web_app"
	case hasServerDeps || hasManagePy:
		return "api"
	case hasCLISignals && (hasMainGo || hasMainPy || hasCmdDir):
		return "cli"
	case hasMainGo || hasMainPy || hasCmdDir:
		// Has an entrypoint but no obvious web/CLI signal → call it a service.
		return "service"
	case hasLibrarySignals:
		return "library"
	}

	// Fall back to a language-flavored guess.
	switch language {
	case "Go", "Rust":
		return "service"
	case "JavaScript", "TypeScript":
		return "library"
	case "Python":
		return "library"
	}
	return "service"
}

// projectName uses the directory name as the project name. The mandate
// can be re-composed with an overridden name from a flag if desired.
func projectName(root string) string {
	return filepath.Base(root)
}
