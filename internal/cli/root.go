package cli

import "github.com/spf13/cobra"

const Version = "0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:           "teststop",
	Short:         "Trigger AI to test software the way a real adversarial user would break it.",
	Long:          longDescription,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

const longDescription = `teststop is an agent-native CLI that gives AI the right mandate
to test any codebase as a real adversarial user — not as a developer.

Zero configuration. Universal. Self-reducing. Designed to disappear.

The mandate is the heart of teststop. Read it with:
    teststop mandate --show
`

// Execute is the entry point invoked from cmd/teststop/main.go.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd, statusCmd, memoryCmd, reportCmd, mandateCmd)
}
