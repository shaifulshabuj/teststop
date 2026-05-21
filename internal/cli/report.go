package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shaifulshabuj/teststop/internal/memory"
	"github.com/shaifulshabuj/teststop/internal/reporter"

	"github.com/spf13/cobra"
)

var (
	reportPath   string
	reportFormat string
	reportFrom   string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Render a report from the most recent run.",
	Long: `Render the most recent JSON report (written by --report-dir on
teststop run) in text, JSON, or Markdown. If no --from is given,
teststop picks the newest *.json file in the project's .teststop
directory.`,
	RunE: runReport,
}

func init() {
	reportCmd.Flags().StringVar(&reportPath, "path", ".", "Path to the project")
	reportCmd.Flags().StringVar(&reportFormat, "format", "text", "Report format: text | json | markdown")
	reportCmd.Flags().StringVar(&reportFrom, "from", "", "Path to a specific JSON report file to render")
}

func runReport(cmd *cobra.Command, _ []string) error {
	abs, err := filepath.Abs(reportPath)
	if err != nil {
		return err
	}

	source := reportFrom
	if source == "" {
		source, err = pickLatestReport(filepath.Join(abs, memory.Dir))
		if err != nil {
			return err
		}
	}
	if source == "" {
		return fmt.Errorf("no report found; run `teststop run --report-dir %s` first", filepath.Join(abs, memory.Dir))
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read report: %w", err)
	}
	var run reporter.Run
	if err := json.Unmarshal(data, &run); err != nil {
		return fmt.Errorf("parse report: %w", err)
	}

	switch strings.ToLower(reportFormat) {
	case "json":
		return reporter.WriteJSON(cmd.OutOrStdout(), run)
	case "markdown", "md":
		return reporter.WriteMarkdown(cmd.OutOrStdout(), run)
	case "text", "":
		return reporter.WriteText(cmd.OutOrStdout(), run)
	default:
		return fmt.Errorf("unknown --format %q (want text|json|markdown)", reportFormat)
	}
}

func pickLatestReport(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	type item struct {
		path string
		name string
	}
	var items []item
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "teststop-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		items = append(items, item{filepath.Join(dir, name), name})
	}
	if len(items) == 0 {
		return "", nil
	}
	// Filenames embed a sortable timestamp; lexicographic sort works.
	sort.Slice(items, func(i, j int) bool {
		return items[i].name > items[j].name
	})
	return items[0].path, nil
}
