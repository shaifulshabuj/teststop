package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reportFormat string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show the last run report",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("teststop report: not yet implemented (Phase 7)")
		return nil
	},
}

func init() {
	reportCmd.Flags().StringVar(&reportFormat, "format", "text", "Output format: text | md")
}
