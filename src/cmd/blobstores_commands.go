// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 20:06
// Original filename: src/cmd/blobstores_commands.go

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"nxtools/blobstores"
)

var blobCmd = &cobra.Command{
	Use:     "blob",
	Aliases: []string{"blobs", "blobstore", "blobstores"},
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
		if _, err := blobstores.ListBlobs(true); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var blobRemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm", "delete", "del"},
	Example: "nxtools blob rm FLAGS blobstore",
	Args:    cobra.ExactArgs(1),
	Short:   "Removes a blobstore from the server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := blobstores.RemoveBlob(args[0]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var blobAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"create"},
	Example: "nxtools blob add FLAGS blobstore",
	Args:    cobra.ExactArgs(1),
	Short:   "Creates a blobstore from the server",
	Run: func(cmd *cobra.Command, args []string) {
		if blobstores.SoftQuotaEnabled {
			blobstores.SoftQuotaSummary = blobstores.SoftQuotaStruct{Type: blobstores.SoftQuotaType, Limit: blobstores.SoftQuotaLimit}
		}
		if err := blobstores.CreateBlob(args[0]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	blobCmd.AddCommand(blobListCmd, blobRemoveCmd, blobAddCmd)

	blobAddCmd.Flags().StringVar(&blobstores.Blobtype, "type", "file", "Blob type (file, gcp, amazon, azure, group)")
	blobAddCmd.Flags().StringVar(&blobstores.FileBlobPath, "path", "", "File blob path")
	blobAddCmd.Flags().BoolVar(&blobstores.SoftQuotaEnabled, "softquota", false, "Soft quota enabled or not")
	blobAddCmd.Flags().StringVar(&blobstores.SoftQuotaType, "sqtype", "spaceUsedQuota", "Softquota type ('spaceRemainingQuota' or 'spaceUsedQuota'")
	blobAddCmd.Flags().Int64Var(&blobstores.SoftQuotaLimit, "sqlimit", 0, "Soft quota limit")
	_ = blobAddCmd.MarkFlagRequired("type")
}
