// nxtools
// src/cmd/root.go

package cmd

import (
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "nxtools",
	Short:   "Nexus Repository Manager 3 CLI tool",
	Version: "0.10.00_POC (" + time.Now().Format("2006.01.02") + "), Go version = " + runtime.Version(),
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

	rootCmd.AddCommand(envCmd, repoCmd, blobCmd)

	rootCmd.PersistentFlags().StringVarP(&Envfile, "env", "e", "defaultEnv.json", "Environment file to load (from $HOME/.config/JFG/nxtools)")
}
