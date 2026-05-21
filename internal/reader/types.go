package reader

// Flow is a single user-reachable surface in the system under test:
// an HTTP route, an exported function, a CLI command, etc.
type Flow struct {
	Kind     string `json:"kind"`     // http | cli | export | event
	Name     string `json:"name"`     // GET /users/:id, --help, ExportedFunc
	Location string `json:"location"` // file:line reference (best effort)
}

// ProjectContext is the digest of a project that the mandate composer
// injects into the AI prompt. Keep it short and high-signal — the AI
// does not need a full code listing.
type ProjectContext struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Language     string   `json:"language"`
	Type         string   `json:"type"` // web_app | api | cli | library | service
	EntryPoints  []string `json:"entry_points"`
	KeyFlows     []Flow   `json:"key_flows"`
	Dependencies []string `json:"dependencies"`
	TestFiles    []string `json:"test_files"`
	FileCount    int      `json:"file_count"`
	Complexity   string   `json:"complexity"` // simple | moderate | complex
}
