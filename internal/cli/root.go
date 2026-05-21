package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "teststop",
	Short: "Trigger AI to test your software like a real adversarial user",
	Long: `teststop is a CLI tool with one job:
Trigger AI to test any software system the way a real adversarial user would break it.

It is NOT a test runner. It is a TRIGGER — a thin CLI that gives AI the right
mandate, then gets out of the way.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(mandateCmd)
}
