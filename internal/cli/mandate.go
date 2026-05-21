package cli

import (
	"fmt"

	"github.com/shaifulshabuj/teststop/mandate"
	"github.com/spf13/cobra"
)

var mandateShow bool

var mandateCmd = &cobra.Command{
	Use:   "mandate",
	Short: "Show the mandate (instruction) sent to the AI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !mandateShow {
			return fmt.Errorf("use --show to print the mandate")
		}
		fmt.Println(mandate.Base)
		return nil
	},
}

func init() {
	mandateCmd.Flags().BoolVar(&mandateShow, "show", false, "Print the full mandate text")
}
