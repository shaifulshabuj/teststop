package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shaifulshabuj/teststop/internal/memory"

	"github.com/spf13/cobra"
)

var (
	statusPath   string
	statusFormat string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current confidence state of the project.",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().StringVar(&statusPath, "path", ".", "Path to the project")
	statusCmd.Flags().StringVar(&statusFormat, "output", "text", "Output format: text | json")
}

func runStatus(cmd *cobra.Command, _ []string) error {
	abs, err := filepath.Abs(statusPath)
	if err != nil {
		return err
	}
	mem, err := memory.Load(abs)
	if err != nil {
		return err
	}

	switch strings.ToLower(statusFormat) {
	case "json":
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(mem)
	case "text", "":
		printStatus(cmd, abs, mem)
		return nil
	default:
		return fmt.Errorf("unknown --output %q (want text|json)", statusFormat)
	}
}

func printStatus(cmd *cobra.Command, root string, m *memory.Memory) {
	fmt.Fprintf(cmd.OutOrStdout(), "teststop status — %s\n\n", root)
	if m.TotalRuns == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No memory yet. Run `teststop run` to begin.")
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Maturity:   %s\n", m.MaturityStage)
	fmt.Fprintf(cmd.OutOrStdout(), "Confidence: %.2f\n", m.OverallConfidence)
	fmt.Fprintf(cmd.OutOrStdout(), "Total runs: %d\n", m.TotalRuns)
	fmt.Fprintf(cmd.OutOrStdout(), "Last run:   %s\n\n", m.LastRun.Format("2006-01-02 15:04:05 UTC"))

	type entry struct {
		key string
		a   memory.AreaConfidence
	}
	rows := make([]entry, 0, len(m.SystemAreas))
	for k, a := range m.SystemAreas {
		rows = append(rows, entry{k, a})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].a.Confidence != rows[j].a.Confidence {
			return rows[i].a.Confidence > rows[j].a.Confidence
		}
		return rows[i].key < rows[j].key
	})

	fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %-9s %s\n", "AREA", "STATUS", "CONF", "TESTS")
	for _, r := range rows {
		fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %-9.2f %d\n",
			truncate(r.key, 40), r.a.Status, r.a.Confidence, r.a.TestCount)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
