package reader

import (
	"os"
	"path/filepath"
	"strings"
)

// languagePriority defines tie-break order when multiple languages have equal file count.
var languagePriority = []string{
	"Go", "Python", "TypeScript", "JavaScript", "Java", "Rust", "Ruby", "PHP", "C#", "Swift", "Kotlin", "C/C++",
}

// maxContentSize is the largest file we will read for pattern matching.
const maxContentSize = 100 * 1024 // 100 KB

// Detect enriches ctx with Language, Type, and EntryPoints by analyzing the scanned files.
func Detect(ctx *ProjectContext) error {
	ctx.Language = detectPrimaryLanguage(ctx.Languages)
	ctx.Type = detectType(ctx)
	return nil
}

// detectPrimaryLanguage returns the language with the most files.
// In case of a tie, uses languagePriority order.
func detectPrimaryLanguage(langs map[string]int) string {
	if len(langs) == 0 {
		return "unknown"
	}

	maxCount := 0
	for _, count := range langs {
		if count > maxCount {
			maxCount = count
		}
	}

	// Among languages with maxCount, pick highest priority
	for _, lang := range languagePriority {
		if langs[lang] == maxCount {
			return lang
		}
	}

	// Fallback: any language with maxCount
	for lang, count := range langs {
		if count == maxCount {
			return lang
		}
	}

	return "unknown"
}

// detectType determines the system type by checking for indicator files/patterns.
func detectType(ctx *ProjectContext) string {
	root := ctx.Path

	// mobile_app: AndroidManifest.xml, *.xcodeproj, pubspec.yaml
	if fileExists(filepath.Join(root, "AndroidManifest.xml")) ||
		fileExists(filepath.Join(root, "pubspec.yaml")) ||
		globExists(root, "*.xcodeproj") {
		return "mobile_app"
	}

	// data_pipeline: files with pipeline/etl/transform/ingest in name
	if hasFileWithKeyword(root, []string{"pipeline", "etl", "transform", "ingest"}) {
		return "data_pipeline"
	}

	// cli: cobra/cli main, cli.go, or python console_scripts
	if isCLIProject(ctx) {
		return "cli"
	}

	// web_app: index.html, react/vue/angular in package.json, templates/*.html
	if isWebApp(ctx) {
		return "web_app"
	}

	// api: handler/controller/routes files, or HTTP server patterns in source
	if isAPIProject(ctx) {
		return "api"
	}

	// library: no main function/entry point, or -lib/-sdk/-client suffix
	if isLibrary(ctx) {
		return "library"
	}

	return "unknown"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// globExists checks if any file matching pattern exists directly under root.
func globExists(root, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	return err == nil && len(matches) > 0
}

// hasFileWithKeyword returns true if any file under root has one of the keywords in its base name.
func hasFileWithKeyword(root string, keywords []string) bool {
	found := false
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(p)
			if p != root && (skipDirs[base] || strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		lower := strings.ToLower(filepath.Base(p))
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

// readFileIfSmall reads file content only if the file is <= maxContentSize bytes.
func readFileIfSmall(path string) []byte {
	info, err := os.Stat(path)
	if err != nil || info.Size() > maxContentSize {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// fileContentContains returns true if path exists, is small enough, and contains substr.
func fileContentContains(path, substr string) bool {
	data := readFileIfSmall(path)
	return data != nil && strings.Contains(string(data), substr)
}

// sourceContainsPattern walks source files and checks for a substring pattern.
func sourceContainsPattern(root string, extensions []string, pattern string) bool {
	extSet := make(map[string]bool, len(extensions))
	for _, e := range extensions {
		extSet[e] = true
	}

	found := false
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(p)
			if p != root && (skipDirs[base] || strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if !extSet[ext] {
			return nil
		}
		if strings.Contains(string(readFileIfSmall(p)), pattern) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func isCLIProject(ctx *ProjectContext) bool {
	root := ctx.Path

	// cobra/urfave/cli import in main.go — but not if the file also starts an HTTP server
	mainGo := filepath.Join(root, "main.go")
	if content := readFileIfSmall(mainGo); content != nil {
		s := string(content)
		hasCLI := strings.Contains(s, "cobra") || strings.Contains(s, "urfave/cli")
		hasHTTP := strings.Contains(s, "http.ListenAndServe") ||
			strings.Contains(s, "gin.Default()") ||
			strings.Contains(s, "echo.New()")
		if hasCLI && !hasHTTP {
			return true
		}
	}

	// cmd/**/main.go with cobra — same HTTP exclusion
	cmdDir := filepath.Join(root, "cmd")
	if dirExists(cmdDir) {
		hasCobra := false
		_ = filepath.Walk(cmdDir, func(p string, info os.FileInfo, err error) error {
			if err != nil || hasCobra {
				return nil
			}
			if !info.IsDir() && filepath.Base(p) == "main.go" {
				if content := readFileIfSmall(p); content != nil {
					s := string(content)
					hasCLI := strings.Contains(s, "cobra") || strings.Contains(s, "urfave/cli")
					hasHTTP := strings.Contains(s, "http.ListenAndServe") ||
						strings.Contains(s, "gin.Default()") ||
						strings.Contains(s, "echo.New()")
					if hasCLI && !hasHTTP {
						hasCobra = true
					}
				}
			}
			return nil
		})
		if hasCobra {
			return true
		}
	}

	// cli.go exists
	if fileExists(filepath.Join(root, "cli.go")) {
		return true
	}

	// Python setup.py with console_scripts
	setupPy := filepath.Join(root, "setup.py")
	if fileContentContains(setupPy, "console_scripts") {
		return true
	}

	return false
}

func isWebApp(ctx *ProjectContext) bool {
	root := ctx.Path

	// index.html at root
	if fileExists(filepath.Join(root, "index.html")) {
		return true
	}

	// package.json with react/vue/angular/next/nuxt
	pkgJSON := filepath.Join(root, "package.json")
	if content := readFileIfSmall(pkgJSON); content != nil {
		s := string(content)
		for _, fw := range []string{"react", "vue", "angular", "next", "nuxt"} {
			if strings.Contains(s, fw) {
				return true
			}
		}
	}

	// templates/ directory with .html files
	templatesDir := filepath.Join(root, "templates")
	if dirExists(templatesDir) {
		hasHTML := false
		_ = filepath.Walk(templatesDir, func(p string, info os.FileInfo, err error) error {
			if err != nil || hasHTML {
				return nil
			}
			if !info.IsDir() && strings.ToLower(filepath.Ext(p)) == ".html" {
				hasHTML = true
			}
			return nil
		})
		if hasHTML {
			return true
		}
	}

	return false
}

func isAPIProject(ctx *ProjectContext) bool {
	root := ctx.Path

	// Files named *_handler.go, *_controller.go, routes.go, router.go
	apiIndicators := []string{"_handler.go", "_controller.go", "routes.go", "router.go"}
	found := false
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(p)
			if p != root && (skipDirs[base] || strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		base := filepath.Base(p)
		for _, indicator := range apiIndicators {
			if strings.HasSuffix(base, indicator) || base == indicator {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	if found {
		return true
	}

	// HTTP server patterns in Go source
	goPatterns := []string{
		"http.ListenAndServe",
		"gin.Default()",
		"echo.New()",
	}
	for _, pat := range goPatterns {
		if sourceContainsPattern(root, []string{".go"}, pat) {
			return true
		}
	}

	// Python patterns
	pyPatterns := []string{"fastapi", "flask", "from flask", "from fastapi"}
	for _, pat := range pyPatterns {
		if sourceContainsPattern(root, []string{".py"}, pat) {
			return true
		}
	}

	// JS/TS patterns
	jsPatterns := []string{"express()"}
	for _, pat := range jsPatterns {
		if sourceContainsPattern(root, []string{".js", ".ts", ".mjs", ".cjs"}, pat) {
			return true
		}
	}

	return false
}

func isLibrary(ctx *ProjectContext) bool {
	name := strings.ToLower(ctx.Name)
	for _, suffix := range []string{"-lib", "-sdk", "-client"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	// No entry points detected
	if len(ctx.EntryPoints) == 0 {
		return true
	}

	return false
}
