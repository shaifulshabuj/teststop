package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current confidence state of the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("teststop status: not yet implemented (Phase 7)")
		return nil
	},
}
