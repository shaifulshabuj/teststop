package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	internalmandatepkg "github.com/shaifulshabuj/teststop/internal/mandate"
	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reader"
	basemandatepkg "github.com/shaifulshabuj/teststop/mandate"
)

var (
	mandateShow  bool
	mandateDepth string
)

var mandateCmd = &cobra.Command{
	Use:   "mandate [path]",
	Short: "Show the mandate (instruction) sent to the AI",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !mandateShow {
			return fmt.Errorf("use --show to print the mandate")
		}

		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		absPath, _ := filepath.Abs(path)

		// Try to compose with real project context; fall back to base mandate on error.
		ctx, err := reader.ScanProject(absPath)
		if err != nil {
			// Fallback: show the embedded base mandate without substitutions.
			fmt.Fprintln(cmd.OutOrStdout(), basemandatepkg.BaseMandateContent)
			return nil
		}

		mem, err := memory.Load(absPath)
		if err != nil || mem == nil {
			mem = memory.NewMemory()
		}

		composed, err := internalmandatepkg.Compose(ctx, mem, internalmandatepkg.Options{Depth: mandateDepth})
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), composed)
		return nil
	},
}

func init() {
	mandateCmd.Flags().BoolVar(&mandateShow, "show", false, "Print the full mandate text")
	mandateCmd.Flags().StringVar(&mandateDepth, "depth", "normal", "Testing depth: light | normal | aggressive")
}
