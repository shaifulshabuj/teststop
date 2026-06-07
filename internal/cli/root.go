package cli

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Build metadata, populated by Execute from the values main injects at build time.
var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

// Command group IDs used to organize `teststop --help`.
const (
	groupCore = "core"
	groupMeta = "meta"
)

var rootCmd = &cobra.Command{
	Use:   "teststop",
	Short: "Trigger AI to test your software like a real adversarial user",
	Long: `teststop is a CLI tool with one job:
Trigger AI to test any software system the way a real adversarial user would break it.

It is NOT a test runner. It is a TRIGGER — a thin CLI that gives AI the right
mandate, then gets out of the way.`,
	Example: `  teststop run                                  Test the current directory
  teststop run --path ./src --depth aggressive  Deeper testing of a subdirectory
  teststop run --target http://localhost:8080   Execute scenarios against a running app
  teststop status                               Show the confidence state
  teststop report --format md                   Show the last run report as Markdown
  teststop mandate --show                        Print the exact mandate sent to the AI
  teststop version                              Print version and build info`,
}

// Execute runs the root command. version, commit, and date are the build
// metadata injected by main (see cmd/teststop/main.go).
func Execute(version, commit, date string) {
	buildVersion = resolveVersion(version)
	buildCommit = commit
	buildDate = date

	// Setting Version makes cobra wire up `--version` (and `-v`, since it is free).
	rootCmd.Version = buildVersion
	rootCmd.SetVersionTemplate("teststop {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// resolveVersion prefers the ldflags-injected version (release builds) and falls
// back to the module version recorded in the binary by `go install`.
func resolveVersion(injected string) string {
	if injected != "" && injected != "dev" {
		return injected
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return injected
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{ID: groupCore, Title: "Core Commands:"},
		&cobra.Group{ID: groupMeta, Title: "Meta Commands:"},
	)

	runCmd.GroupID = groupCore
	statusCmd.GroupID = groupCore
	memoryCmd.GroupID = groupCore
	reportCmd.GroupID = groupCore
	mandateCmd.GroupID = groupCore
	versionCmd.GroupID = groupMeta

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(mandateCmd)
	rootCmd.AddCommand(versionCmd)

	// Put the built-in help and completion commands in the Meta group too.
	rootCmd.SetHelpCommandGroupID(groupMeta)
	rootCmd.SetCompletionCommandGroupID(groupMeta)
}
