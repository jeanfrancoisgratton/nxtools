// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/cmd/repo_commands.go

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"nxtools/repositories"
)

var repoFormat string
var repoType string
var repoJSONFile string

var repoCmd = &cobra.Command{
	Use:     "repo",
	Aliases: []string{"repos", "repositories"},
	Short:   "Repository-related sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list | create | delete }")
	},
}

var repoListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Example: "nxtools repo list -e defaultEnv.json",
	Short:   "Lists all repositories visible to the configured user",
	Run: func(cmd *cobra.Command, args []string) {
		if err := repositories.ListRepositories(Envfile); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var repoCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Example: "nxtools repo create -e defaultEnv.json --format yum --type hosted --json payload.json",
	Short:   "Creates a repository (payload is recipe-specific)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := repositories.CreateRepository(Envfile, repoFormat, repoType, repoJSONFile); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var repoDeleteCmd = &cobra.Command{
	Use:     "delete REPO_NAME",
	Aliases: []string{"rm", "remove"},
	Example: "nxtools repo delete -e defaultEnv.json my-repo-name",
	Short:   "Deletes a repository by name",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := repositories.DeleteRepository(Envfile, args[0]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	repoCreateCmd.Flags().StringVar(&repoFormat, "format", "", "Repository format/recipe family (e.g. yum, apt, maven, docker)")
	repoCreateCmd.Flags().StringVar(&repoType, "type", "", "Repository type (hosted, proxy, group)")
	repoCreateCmd.Flags().StringVar(&repoJSONFile, "json", "", "JSON payload file for the recipe (use '-' to read from stdin)")
	_ = repoCreateCmd.MarkFlagRequired("format")
	_ = repoCreateCmd.MarkFlagRequired("type")
	_ = repoCreateCmd.MarkFlagRequired("json")

	repoCmd.AddCommand(repoListCmd, repoCreateCmd, repoDeleteCmd)
}
