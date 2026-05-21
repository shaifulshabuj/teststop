package cli

import (
	"fmt"

	"github.com/shaifulshabuj/teststop/mandate"
	"github.com/spf13/cobra"
)

var mandateShow bool

var mandateCmd = &cobra.Command{
	Use:   "mandate",
	Short: "Print the adversarial user mandate (the instruction given to AI).",
	Long: `The mandate is the intellectual heart of teststop. It is the exact
instruction given to the AI that makes it test like a real adversarial user
instead of a developer. It is open, auditable, and improvable.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !mandateShow {
			return cmd.Help()
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), mandate.Base)
		return err
	},
}

func init() {
	mandateCmd.Flags().BoolVar(&mandateShow, "show", false, "Print the base mandate to stdout")
}
