package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shaifulshabuj/teststop/internal/memory"

	"github.com/spf13/cobra"
)

var (
	memoryPath   string
	memoryFormat string
	memoryReset  bool
)

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Inspect or reset what teststop has learned about this project.",
	RunE:  runMemory,
}

func init() {
	memoryCmd.Flags().StringVar(&memoryPath, "path", ".", "Path to the project")
	memoryCmd.Flags().StringVar(&memoryFormat, "output", "text", "Output format: text | json")
	memoryCmd.Flags().BoolVar(&memoryReset, "reset", false, "Clear accumulated memory and start fresh")
}

func runMemory(cmd *cobra.Command, _ []string) error {
	abs, err := filepath.Abs(memoryPath)
	if err != nil {
		return err
	}
	if memoryReset {
		if err := memory.Reset(abs); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "teststop: memory cleared at %s\n", filepath.Join(abs, memory.Dir))
		return nil
	}
	mem, err := memory.Load(abs)
	if err != nil {
		return err
	}
	switch strings.ToLower(memoryFormat) {
	case "json":
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(mem)
	case "text", "":
		printMemory(cmd, mem)
		return nil
	default:
		return fmt.Errorf("unknown --output %q (want text|json)", memoryFormat)
	}
}

func printMemory(cmd *cobra.Command, m *memory.Memory) {
	if m.TotalRuns == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No memory yet. Run `teststop run` to begin.")
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(),
		"Areas: %d total · stable %d · volatile %d · retired %d\n",
		len(m.SystemAreas), len(m.StableAreas()), len(m.VolatileAreas()), len(m.Retired))
	for k, a := range m.SystemAreas {
		fmt.Fprintf(cmd.OutOrStdout(),
			"- %s [%s] conf=%.2f tests=%d (pass=%d fail=%d)\n",
			k, a.Status, a.Confidence, a.TestCount, a.PassCount, a.FailCount)
	}
}
