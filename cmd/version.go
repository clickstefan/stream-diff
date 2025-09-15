package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build information (set by build flags)
var (
	version   = "dev"
	commit    = "unknown"
	date      = "unknown"
	builtBy   = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display detailed version information including build details,
Go version, and system information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Stream-Diff Data Comparator\n")
		fmt.Printf("  Version:    %s\n", version)
		fmt.Printf("  Commit:     %s\n", commit)
		fmt.Printf("  Built:      %s\n", date)
		fmt.Printf("  Built by:   %s\n", builtBy)
		fmt.Printf("  Go version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}