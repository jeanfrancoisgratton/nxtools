// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/11
// Original filename: src/cmd/assets_commands.go

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"nxtools/assets"
	"nxtools/tasks"
)

var assetsCmd = &cobra.Command{
	Use:     "assets",
	Aliases: []string{"asset"},
	Short:   "Asset-related sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list | info | upload }")
	},
}

var assetsListCmd = &cobra.Command{
	Use:     "list REPO_NAME",
	Aliases: []string{"ls"},
	Example: "nxtools assets list [-e defaultEnv.json] my-repository [--latest]",
	Args:    cobra.ExactArgs(1),
	Short:   "Lists all assets from a repository",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := assets.ListAssets(args[0], assets.LatestAssetsOnly, true); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var assetInfoCmd = &cobra.Command{
	Use:     "info ASSET_ID",
	Example: "nxtools assets info [-e defaultEnv.json] ASSET_ID",
	Args:    cobra.ExactArgs(1),
	Short:   "Provides information on an asset",
	Run: func(cmd *cobra.Command, args []string) {
		if err := assets.AssetInformation(args[0]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var assetsUploadCmd = &cobra.Command{
	Use:     "upload REPO_NAME FILE_NAME",
	Aliases: []string{"push"},
	Example: "nxtools assets upload [-e defaultEnv.json] [-r] REPO_NAME /path/to/file",
	Args:    cobra.ExactArgs(2),
	Short:   "Uploads a file to a supported hosted repository",
	Run: func(cmd *cobra.Command, args []string) {
		if err := assets.UploadAsset(args[0], args[1], assets.UploadDirectory); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if assets.ReindexRepo {
			if err := tasks.ReindexRepo(args[0]); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	},
}

var assetsDownloadCmd = &cobra.Command{
	Use:     "download URL DEST_FILE",
	Aliases: []string{"get"},
	Example: "nxtools assets download [-e defaultEnv.json] ASSET_URL /PATH/TO/FILE",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Downloads a file from a supported hosted repository",
	Run: func(cmd *cobra.Command, args []string) {
		destfile := ""
		if len(args) == 2 {
			destfile = args[1]
		}
		if err := assets.DownloadAsset(args[0], destfile); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var assetsDeleteCmd = &cobra.Command{
	Use:     "delete ASSET_ID1 [ASSET_ID2 ...]",
	Aliases: []string{"rm"},
	Example: "nxtools assets delete [-e defaultEnv.json] ASSET_ID [ASSET_ID2 ...]",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Deletes one or more assets from a repository",
	Run: func(cmd *cobra.Command, args []string) {
		if err := assets.DeleteAssets(args); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var assetsComponentInfoCmd = &cobra.Command{
	Use:     "pkginfo REPO_NAME PKG_NAME",
	Example: "nxtools assets pkginfo [-e defaultEnv.json] REPO_NAME PKG_NAME",
	Args:    cobra.ExactArgs(2),
	Short:   "Fetches information about a package from a given repository",
	Long: `This subcommand relies on the component API endpoint and thus needs a package name (not an asset ID),
and the name of the repository where the package is housed.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := assets.PackageInfo(args[0], args[1], true); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	assetsCmd.AddCommand(assetsListCmd, assetInfoCmd, assetsUploadCmd, assetsDownloadCmd, assetsDeleteCmd, assetsComponentInfoCmd)

	assetsListCmd.Flags().BoolVarP(&assets.LatestAssetsOnly, "latest", "l", false, "Only list the latest version of each logical asset/component")
	assetsListCmd.Flags().BoolVarP(&assets.AlternateInfo, "alternate", "a", false, "Show alternate asset information")
	assetsUploadCmd.Flags().StringVarP(&assets.UploadDirectory, "directory", "d", "/", "Target directory inside the repository (defaults inferred for raw and yum; for alpine, <version>/<repository>, e.g. edge/main; defaults to edge/main if omitted)")
	assetsUploadCmd.Flags().BoolVarP(&assets.ReindexRepo, "reindex", "r", false, "Reindex repo after the upload")

	assetInfoCmd.Flags().BoolVarP(&assets.JsonOutput, "json", "j", false, "Output asset information as JSON")
	assetsComponentInfoCmd.Flags().BoolVarP(&assets.JsonOutput, "json", "j", false, "Output asset information as JSON")
}
