package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// skipDirs are directories that should never be scanned.
var skipDirs = map[string]bool{
	".git":        true,
	"node_modules": true,
	"vendor":      true,
	".teststop":   true,
	"dist":        true,
	"build":       true,
	"__pycache__": true,
	".venv":       true,
	"target":      true,
	"bin":         true,
	"obj":         true,
	".next":       true,
}

// extToLanguage maps file extensions to language names.
var extToLanguage = map[string]string{
	".go":   "Go",
	".py":   "Python",
	".js":   "JavaScript",
	".mjs":  "JavaScript",
	".cjs":  "JavaScript",
	".ts":   "TypeScript",
	".tsx":  "TypeScript",
	".rb":   "Ruby",
	".java": "Java",
	".rs":   "Rust",
	".php":  "PHP",
	".cs":   "C#",
	".swift": "Swift",
	".kt":   "Kotlin",
	".kts":  "Kotlin",
	".cpp":  "C/C++",
	".cc":   "C/C++",
	".cxx":  "C/C++",
	".c":    "C/C++",
}

// Scan walks path and returns a ProjectContext populated with basic file info.
// Call Detect and Analyze after Scan to enrich the context.
func Scan(path string) (ProjectContext, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ProjectContext{}, fmt.Errorf("resolve path %q: %w", path, err)
	}

	ctx := ProjectContext{
		Path:      absPath,
		Name:      filepath.Base(absPath),
		Languages: make(map[string]int),
	}

	entryPointNames := map[string]bool{
		"main.go":         true,
		"main.py":         true,
		"__main__.py":     true,
		"app.py":          true,
		"server.py":       true,
		"index.js":        true,
		"index.ts":        true,
		"server.js":       true,
		"app.js":          true,
		"main.rs":         true,
		"Program.cs":      true,
		"Application.java": true,
	}

	err = filepath.Walk(absPath, func(p string, info os.FileInfo, werr error) error {
		if werr != nil {
			// Skip files/dirs we cannot access
			return nil
		}

		if info.IsDir() {
			base := filepath.Base(p)
			// Always skip the root itself
			if p == absPath {
				return nil
			}
			// Skip hidden dirs (but allow root-level hidden files, not dirs)
			if strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			if skipDirs[base] {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		ctx.FileCount++

		ext := strings.ToLower(filepath.Ext(p))
		if lang, ok := extToLanguage[ext]; ok {
			ctx.Languages[lang]++
		}

		// Collect entry points (up to 20)
		if len(ctx.EntryPoints) < 20 {
			base := filepath.Base(p)
			if entryPointNames[base] {
				ctx.EntryPoints = append(ctx.EntryPoints, p)
			}
		}

		return nil
	})
	if err != nil {
		return ctx, fmt.Errorf("walk %q: %w", absPath, err)
	}

	return ctx, nil
}

// ScanProject runs Scan + Detect + Analyze and returns a fully populated ProjectContext.
func ScanProject(path string) (ProjectContext, error) {
	ctx, err := Scan(path)
	if err != nil {
		return ctx, fmt.Errorf("scan: %w", err)
	}
	if err := Detect(&ctx); err != nil {
		return ctx, fmt.Errorf("detect: %w", err)
	}
	if err := Analyze(&ctx); err != nil {
		return ctx, fmt.Errorf("analyze: %w", err)
	}
	return ctx, nil
}
