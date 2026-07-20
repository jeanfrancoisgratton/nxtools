// nxtools
// src/cmd/root.go

package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"nxtools/shared"

	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nxtools",
	Short: "Nexus Repository Manager 3 CLI tool",
	Long:  `This tools allows you to manage many actions on an NxRM server`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the software version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(hftx.White("nxtools v1.0.1 (2026.07.20), Go version = v" + strings.TrimPrefix(runtime.Version(), "go")))
	},
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

	rootCmd.AddCommand(versionCmd, envCmd, repoCmd, blobCmd, completionCmd, assetsCmd, assetsUploadCmd, assetsDownloadCmd, assetsComponentInfoCmd, repoReindexCmd, repoMigrateCmd)

	rootCmd.Flags().BoolP("version", "V", false, "Show version and exit")
	rootCmd.PersistentFlags().StringVarP(&shared.Envfile, "env", "e", "defaultEnv.json", "Environment file to load in from $HOME/.config/JFG/nxtools")
	rootCmd.PersistentFlags().BoolVarP(&shared.QuietOutput, "quiet", "q", false, "Output will be as quiet as possible")
}
