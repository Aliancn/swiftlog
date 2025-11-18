package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "swiftlog",
	Short: "SwiftLog - Script log collection and analysis platform",
	Long: `SwiftLog is a lightweight, high-performance platform for collecting,
storing, and analyzing logs from script executions.

Use 'swiftlog run' to execute and log a command, or pipe output directly
to SwiftLog for real-time log collection.`,
	Version: version,
}

// Execute executes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("server", "s", "", "Server address (overrides config)")
	rootCmd.PersistentFlags().StringP("token", "t", "", "API token (overrides config)")
}
