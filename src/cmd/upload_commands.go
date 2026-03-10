// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/09
// Original filename: src/cmd/upload_commands.go

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"nxtools/repositories"
)

var uploadDirectory string

var uploadCmd = &cobra.Command{
	Use:     "upload REPO_NAME FILE_NAME",
	Aliases: []string{"push"},
	Example: "nxtools upload -e defaultEnv.json my-repository /path/to/file.rpm",
	Args:    cobra.ExactArgs(2),
	Short:   "Uploads a file to a supported hosted repository",
	Run: func(cmd *cobra.Command, args []string) {
		if err := repositories.UploadFile(args[0], args[1], uploadDirectory); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	uploadCmd.Flags().StringVar(&uploadDirectory, "directory", "", "Target directory inside the repository (optional; defaults are inferred for raw and yum)")
}
