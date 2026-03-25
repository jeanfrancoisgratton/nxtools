// nxtools
// src/cmd/root.go

package cmd

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"nxtools/shared"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "nxtools",
	Short:   "Nexus Repository Manager 3 CLI tool",
	Version: "0.75.01 (" + time.Now().Format("2006.01.02") + "), Go version = v" + strings.TrimPrefix(runtime.Version(), "go"),
	Long:    `This tools allows you to manage many actions on an NxRM server`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// remove the -v flag from rootCmd so I can use it elsewhere
	if f := rootCmd.Flags().Lookup("version"); f != nil {
		f.Shorthand = ""
	}
	rootCmd.DisableAutoGenTag = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(envCmd, repoCmd, blobCmd, assetsCmd, assetsUploadCmd, assetsDownloadCmd, assetsComponentInfoCmd, reindexRepoCmd)

	rootCmd.PersistentFlags().StringVarP(&shared.Envfile, "env", "e", "defaultEnv.json", "Environment file to load in from $HOME/.config/JFG/nxtools")
	rootCmd.PersistentFlags().BoolVarP(&shared.QuietOutput, "quiet", "q", false, "Output will be as quiet as possible")
}
