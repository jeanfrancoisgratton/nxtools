// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 20:06
// Original filename: src/cmd/blob_commands.go

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"nxtools/blobstores"
)

var blobCmd = &cobra.Command{
	Use:     "blob",
	Aliases: []string{"blobs", "blobstore"},
	Short:   "Blobstore-related sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list | create | delete }")
	},
}

var blobListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Example: "nxtools blob list FLAGS",
	Short:   "Lists all blobs visible to the configured user",
	Run: func(cmd *cobra.Command, args []string) {
		if err := blobstores.ListBlobs(); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm", "delete", "del"},
	Example: "nxtools blob rm FLAGS blobstore",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Removes a blobstore from the server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := blobstores.RemoveBlob(args); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	blobCmd.AddCommand(blobListCmd, removeCmd)
}
