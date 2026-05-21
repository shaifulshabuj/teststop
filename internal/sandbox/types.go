package sandbox

import "os"

// Mode controls how the sandbox runner operates.
type Mode int

const (
	ModeAuto     Mode = iota // use container if available, else direct
	ModeRequired             // error if container not available
	ModeDisabled             // always run directly
)

// ModeFromEnv reads TESTSTOP_SANDBOX env var.
// "required" → ModeRequired, "none" → ModeDisabled, anything else (incl "auto") → ModeAuto
func ModeFromEnv() Mode {
	switch os.Getenv("TESTSTOP_SANDBOX") {
	case "required":
		return ModeRequired
	case "none":
		return ModeDisabled
	default:
		return ModeAuto
	}
}

// RunConfig holds configuration for a sandboxed run.
type RunConfig struct {
	Image  string   // container image (default: DefaultImage)
	Mounts []string // "--volume src:dst:ro" entries
	Env    []string // "KEY=VALUE" entries to forward into container
}

// DefaultImage is the published teststop agent image.
const DefaultImage = "ghcr.io/shaifulshabuj/teststop-agent:latest"

// Result holds the output of a sandboxed command.
type Result struct {
	Stdout []byte
	Stderr []byte
	Err    error
}
