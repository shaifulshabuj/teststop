package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var reportFormat string

var reportCmd = &cobra.Command{
	Use:   "report [path]",
	Short: "Show the last run report",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		absPath, _ := filepath.Abs(path)
		reportsDir := filepath.Join(absPath, ".teststop", "reports")

		entries, err := os.ReadDir(reportsDir)
		if err != nil || len(entries) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No reports yet. Run `teststop run` first.")
			return nil
		}

		// Most recent report = last entry (files are named by timestamp, lexicographically sorted).
		latest := entries[len(entries)-1]
		content, err := os.ReadFile(filepath.Join(reportsDir, latest.Name()))
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), string(content))
		return nil
	},
}

func init() {
	reportCmd.Flags().StringVar(&reportFormat, "format", "md", "Output format: md | text")
}
