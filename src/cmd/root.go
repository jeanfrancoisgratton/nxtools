// nxtools
// src/cmd/root.go

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "nxtools",
	Short:   "Nexus Repository Manager 3 CLI tool",
	Version: "1.00.00-0 (2026.01.28)",
	Long:    `This tools allows you to manage many actions on an NxRM server`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.DisableAutoGenTag = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(envCmd, repoCmd)
}
