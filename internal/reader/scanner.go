package reader

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// scanResult is the raw filesystem signal we extract once and then
// hand off to the detector and analyzer.
type scanResult struct {
	root          string
	files         []string
	extCounts     map[string]int
	manifestFiles []string
	testFiles     []string
}

// ignoredDirs are directories that contribute noise rather than signal
// when reading a project: dependency caches, build artifacts, VCS
// metadata, and editor state. Detection works on source, not vendored
// trees.
var ignoredDirs = map[string]struct{}{
	".git":         {},
	".hg":          {},
	".svn":         {},
	".idea":        {},
	".vscode":      {},
	".teststop":    {},
	"node_modules": {},
	"vendor":       {},
	"dist":         {},
	"build":        {},
	"target":       {},
	"out":          {},
	".next":        {},
	".nuxt":        {},
	"__pycache__":  {},
	".venv":        {},
	"venv":         {},
	"env":          {},
	".tox":         {},
	".mypy_cache":  {},
	".pytest_cache": {},
	"coverage":     {},
	".cache":       {},
}

// scan walks the project root once and returns the raw signal we
// later interpret. It deliberately does not parse source files — that
// is the analyzer's job and is best-effort.
func scan(root string) (*scanResult, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	r := &scanResult{
		root:      abs,
		extCounts: make(map[string]int),
	}

	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Skip unreadable subtrees rather than abort the whole scan.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if path == abs {
				return nil
			}
			name := d.Name()
			if _, skip := ignoredDirs[name]; skip {
				return fs.SkipDir
			}
			if strings.HasPrefix(name, ".") && name != "." {
				// Hidden dirs (other than the root) are skipped by default.
				return fs.SkipDir
			}
			return nil
		}

		name := d.Name()
		r.files = append(r.files, path)
		ext := strings.ToLower(filepath.Ext(name))
		if ext != "" {
			r.extCounts[ext]++
		}
		if isManifest(name) {
			r.manifestFiles = append(r.manifestFiles, path)
		}
		if isTestFile(name) {
			r.testFiles = append(r.testFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func isManifest(name string) bool {
	switch name {
	case "go.mod", "package.json", "Cargo.toml", "pyproject.toml",
		"requirements.txt", "Pipfile", "pom.xml", "build.gradle",
		"build.gradle.kts", "Gemfile", "composer.json", "mix.exs",
		"pubspec.yaml", "deno.json", "bun.lockb", "Dockerfile",
		"docker-compose.yml", "docker-compose.yaml":
		return true
	}
	return false
}

func isTestFile(name string) bool {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, "_test.go"):
		return true
	case strings.HasSuffix(lower, ".test.js"), strings.HasSuffix(lower, ".test.ts"),
		strings.HasSuffix(lower, ".test.jsx"), strings.HasSuffix(lower, ".test.tsx"):
		return true
	case strings.HasSuffix(lower, ".spec.js"), strings.HasSuffix(lower, ".spec.ts"),
		strings.HasSuffix(lower, ".spec.jsx"), strings.HasSuffix(lower, ".spec.tsx"):
		return true
	case strings.HasPrefix(lower, "test_") && strings.HasSuffix(lower, ".py"):
		return true
	case strings.HasSuffix(lower, "_test.py"):
		return true
	case strings.HasSuffix(lower, "_spec.rb"), strings.HasSuffix(lower, "_test.rb"):
		return true
	}
	return false
}
