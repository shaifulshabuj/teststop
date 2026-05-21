package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var memoryReset bool

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Show accumulated testing memory for this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("teststop memory: not yet implemented (Phase 7)")
		return nil
	},
}

func init() {
	memoryCmd.Flags().BoolVar(&memoryReset, "reset", false, "Clear accumulated memory (with confirmation)")
}
