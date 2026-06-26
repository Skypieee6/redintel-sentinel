// Package cli defines the Cobra command tree for the RedIntel Sentinel binary.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Skypieee6/redintel-sentinel/internal/version"
)

// cfgFile is an optional path to an explicit config file, set via --config.
var cfgFile string

// rootCmd is the base command invoked when the binary runs with no subcommand.
var rootCmd = &cobra.Command{
	Use:   "redintel",
	Short: "RedIntel Sentinel - enterprise Attack Surface Management platform",
	Long: `RedIntel Sentinel is an enterprise Attack Surface Management (ASM)
platform for authorized, defensive security assessments.

It helps organizations inventory, monitor and assess the systems they own or
are explicitly authorized to test.`,
	Version:       version.Get().String(),
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"path to config file (default: search ./, ./configs, /etc/redintel)")

	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newMigrateCmd())
	rootCmd.AddCommand(newVersionCmd())
}
