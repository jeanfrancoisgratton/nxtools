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

// buildVersion and buildDate are set via -ldflags -X at package-build time.
// Each __*/ builder reads its own already-authoritative version field
// (APKBUILD's pkgver, PKGBUILD's pkgver, control's Version minus the Debian
// revision, the spec's %{_version}); src/build.sh reads nxtools.json since it
// isn't tied to any one distro's packaging file. The fallbacks below are what
// you get from a plain `go build .` with no ldflags, e.g. local development.
var buildVersion = "dev"
var buildDate = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the software version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(hftx.White("nxtools v" + buildVersion + " (" + buildDate + "), Go version = v" + strings.TrimPrefix(runtime.Version(), "go") + " (" + runtime.GOARCH + ")"))
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

	rootCmd.AddCommand(versionCmd, envCmd, repoCmd, blobCmd, completionCmd, assetsCmd, assetsUploadCmd, assetsDownloadCmd, assetsComponentInfoCmd, repoMigrateCmd)

	rootCmd.Flags().BoolP("version", "V", false, "Show version and exit")
	rootCmd.PersistentFlags().StringVarP(&shared.Envfile, "env", "e", "defaultEnv.json", "Environment file to load in from $HOME/.config/JFG/nxtools")
	rootCmd.PersistentFlags().BoolVarP(&shared.QuietOutput, "quiet", "q", false, "Output will be as quiet as possible")
}
