// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/cmd/repositories_commands.go

package cmd

import (
	"fmt"

	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"github.com/spf13/cobra"
	"nxtools/repositories"
	"nxtools/tasks"
)

var repoCmd = &cobra.Command{
	Use:     "repo",
	Aliases: []string{"repos", "repositories"},
	Short:   "Repository-related sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list | create | delete | type | reindex }")
	},
}

var repoListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Example: "nxtools repo list [-e defaultEnv.json] [--json]",
	Short:   "Lists all repositories visible to the configured user",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := repositories.ListRepositories(true); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var repoCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Example: "nxtools repo create [-e defaultEnv.json] --format FORMAT REPO_NAME BLOBSTORE_NAME",
	Short:   "Creates a repository (payload is recipe-specific)",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := repositories.CreateRepository(args[0], args[1]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var repoDeleteCmd = &cobra.Command{
	Use:     "delete REPO_NAME",
	Aliases: []string{"rm", "remove"},
	Example: "nxtools repo delete [-e defaultEnv.json] my-repo-name",
	Short:   "Deletes a repository by name",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := repositories.DeleteRepository(args[0]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var repoQueryTypeCmd = &cobra.Command{
	Use:     "type",
	Example: "nxtools repo type [-e defaultEnv.json] REPO_NAME",
	Short:   "Returns the type of the repository",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if repoformat, err := repositories.QueryRepoType(args[0]); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Repository format: " + hftx.Blue(repoformat))
		}
	},
}

var reindexRepoCmd = &cobra.Command{
	Use:     "reindex",
	Example: "nxtools repo reindex [-e defaultEnv.json] REPO_NAME",
	Short:   "Rebuilds the repository metadata",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := tasks.ReindexRepo(args[0]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	repoCmd.AddCommand(repoListCmd, repoCreateCmd, repoDeleteCmd, repoQueryTypeCmd, reindexRepoCmd)

	repoListCmd.Flags().BoolVar(&repositories.RepoListJSONOutput, "json", false, "Output repository information as JSON")

	repoCreateCmd.Flags().StringVarP(&repositories.RepoFormat, "format", "f", "", "Repository format/recipe family (e.g. yum, apt, maven, docker)")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoType, "type", "t", "hosted", "Repository type (hosted, proxy, group)")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoSigningFile, "keyfile", "k", "", "Private key location")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoSigningPassphrase, "passphrase", "p", "", "Private key passphrase")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoAptDistro, "distro", "d", "nexus", "Debian-like distribution")
	repoCreateCmd.Flags().StringVarP(&repositories.StorageWritePolicy, "writepolicy", "w", "ALLOW", "Blob storage policy")
	repoCreateCmd.Flags().BoolVarP(&repositories.StorageStrictContentValidation, "validation", "V", true, "Set this to disable strict content validation")
	_ = repoCreateCmd.MarkFlagRequired("format")
}
