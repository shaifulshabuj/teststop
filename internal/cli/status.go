package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shaifulshabuj/teststop/internal/memory"
)

var statusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Show the current confidence state of the project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		mem, err := memory.Load(absPath)
		if err != nil {
			return err
		}

		if len(mem.Areas) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No memory yet. Run `teststop run` first.")
			return nil
		}

		w := cmd.OutOrStdout()
		fmt.Fprintf(w, "teststop status — %s\n\n", absPath)
		fmt.Fprintf(w, "%-30s %10s %8s %8s %8s\n", "AREA", "CONFIDENCE", "TESTS", "PASSES", "STAGE")
		fmt.Fprintf(w, "%s\n", strings.Repeat("─", 70))

		for name, area := range mem.Areas {
			retired := ""
			if area.Retired {
				retired = " [retired]"
			}
			fmt.Fprintf(w, "%-30s %9.1f%% %8d %8d %8s%s\n",
				name,
				area.Confidence*100,
				area.TestCount,
				area.PassCount,
				area.MaturityStage,
				retired,
			)
		}
		return nil
	},
}
