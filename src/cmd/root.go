// nxtools
// src/cmd/root.go

package cmd

import (
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"nxtools/shared"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "nxtools",
	Short:   "Nexus Repository Manager 3 CLI tool",
	Version: "0.20.00 (" + time.Now().Format("2006.01.02") + "), Go version = " + runtime.Version(),
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

	rootCmd.AddCommand(envCmd, repoCmd, blobCmd, uploadCmd, reindexRepoCmd)

	rootCmd.PersistentFlags().StringVarP(&shared.Envfile, "env", "e", "defaultEnv.json", "Environment file to load in from $HOME/.config/JFG/nxtools")
	rootCmd.PersistentFlags().BoolVarP(&shared.QuietOutput, "quiet", "q", false, "Output will be as quiet as possible")
}
