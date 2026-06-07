package main

import "github.com/shaifulshabuj/teststop/internal/cli"

// Build metadata, injected at release time via:
//
//	-ldflags "-X main.version=... -X main.commit=... -X main.date=..."
//
// For `go install` builds these stay at their defaults and the version is
// recovered from the module build info instead.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.Execute(version, commit, date)
}
