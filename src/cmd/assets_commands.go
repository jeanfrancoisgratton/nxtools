// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/11
// Original filename: src/cmd/assets_commands.go

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"nxtools/assets"
)

var assetsCmd = &cobra.Command{
	Use:     "assets",
	Aliases: []string{"asset"},
	Short:   "Asset-related sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list }")
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

func init() {
	assetsCmd.AddCommand(assetsListCmd)

	assetsListCmd.Flags().BoolVar(&assets.LatestAssetsOnly, "latest", false, "Only list the latest version of each logical asset/component")
}
