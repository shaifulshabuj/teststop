package reader

// ProjectContext holds everything teststop learned about a project by static analysis.
type ProjectContext struct {
	Name         string         // project/directory name
	Path         string         // absolute path scanned
	Language     string         // primary language: "Go", "Python", "JavaScript", "TypeScript", "Ruby", "Java", "Rust", "PHP", "C#", "unknown"
	Type         string         // system type: "api", "cli", "web_app", "library", "mobile_app", "data_pipeline", "unknown"
	EntryPoints  []string       // files that serve as entry points (main.go, index.js, app.py, etc.)
	Flows        []Flow         // key user/system flows detected
	Dependencies []string       // notable external dependencies detected
	FileCount    int            // total files scanned
	Languages    map[string]int // file count per language (for multi-language projects)
}

// Flow represents a key user or system flow detected in the codebase.
type Flow struct {
	Name        string // short name (e.g., "POST /api/users", "login", "process-csv")
	Description string // one-line description
	Area        string // which system area (e.g., "auth", "api/users", "cli")
}
