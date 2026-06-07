package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the teststop version and build info",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		w := cmd.OutOrStdout()
		fmt.Fprintf(w, "teststop %s\n", buildVersion)
		fmt.Fprintf(w, "  commit:  %s\n", buildCommit)
		fmt.Fprintf(w, "  built:   %s\n", buildDate)
		fmt.Fprintf(w, "  go:      %s\n", runtime.Version())
		fmt.Fprintf(w, "  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
