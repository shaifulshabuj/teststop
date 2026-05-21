package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/shaifulshabuj/teststop/internal/memory"
)

var (
	memoryReset bool
	memoryYes   bool
)

var memoryCmd = &cobra.Command{
	Use:   "memory [path]",
	Short: "Show accumulated testing memory for this project",
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

		if memoryReset {
			if !memoryYes {
				return fmt.Errorf("reset requires --yes to confirm (e.g. teststop memory --reset --yes)")
			}
			// Delete memory.json and retired.json.
			memFile := filepath.Join(absPath, ".teststop", "memory.json")
			retiredFile := filepath.Join(absPath, ".teststop", "retired.json")

			removed := 0
			for _, f := range []string{memFile, retiredFile} {
				if err := os.Remove(f); err != nil {
					if !os.IsNotExist(err) {
						return fmt.Errorf("removing %s: %w", f, err)
					}
				} else {
					removed++
				}
			}

			w := cmd.OutOrStdout()
			if removed == 0 {
				fmt.Fprintln(w, "No memory files found — nothing to reset.")
			} else {
				fmt.Fprintf(w, "Memory reset: removed %d file(s) from %s/.teststop/\n", removed, absPath)
			}
			return nil
		}

		// Show memory as pretty JSON.
		mem, err := memory.Load(absPath)
		if err != nil {
			return err
		}

		data, err := json.MarshalIndent(mem, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding memory: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	},
}

func init() {
	memoryCmd.Flags().BoolVar(&memoryReset, "reset", false, "Clear accumulated memory (requires --yes)")
	memoryCmd.Flags().BoolVar(&memoryYes, "yes", false, "Confirm reset without prompt")
}
