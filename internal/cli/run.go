package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	runPath      string
	runDepth     string
	runOutput    string
	runThreshold int
	runNoColor   bool
	runQuiet     bool
)

var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Run adversarial user testing on a project",
	Long:  `Scan the project, compose a mandate, trigger AI to generate test scenarios, update confidence memory, and report results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Phase 7 will wire up the full pipeline here.
		fmt.Println("teststop run: not yet implemented (Phase 7)")
		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&runPath, "path", ".", "Path to the project to test")
	runCmd.Flags().StringVar(&runDepth, "depth", "normal", "Testing depth: light | normal | aggressive")
	runCmd.Flags().StringVar(&runOutput, "output", "text", "Output format: json | text | markdown")
	runCmd.Flags().IntVar(&runThreshold, "threshold", 80, "Confidence threshold (0-100)")
	runCmd.Flags().BoolVar(&runNoColor, "no-color", false, "Disable ANSI color output (for agents)")
	runCmd.Flags().BoolVar(&runQuiet, "quiet", false, "Minimal output")
}
