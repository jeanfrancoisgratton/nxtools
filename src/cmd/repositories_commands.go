// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/cmd/repositories_commands.go

package cmd

import (
	"fmt"
	"nxtools/assets"
	"os"
	"strings"

	"nxtools/repositories"
	"nxtools/tasks"

	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"github.com/spf13/cobra"
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

var repoReindexCmd = &cobra.Command{
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

var repoCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Example: "nxtools repo create [-e defaultEnv.json] --format FORMAT REPO_NAME BLOBSTORE_NAME",
	Short:   "Creates a repository (payload is recipe-specific)",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// some sanity checks before going ahead
		writepol := strings.ToLower(repositories.StorageWritePolicy)
		if writepol != "allow" && writepol != "deny" && writepol != "allow_once" {
			hftx.ErrorSign("Supported policies are ALLOW, ALLOW_ONCE and DENY; you selected " + repositories.StorageWritePolicy)
			os.Exit(1)
		} else {
			repositories.StorageWritePolicy = strings.ToUpper(writepol)
		}
		if strings.ToLower(repositories.RepoFormat) == "apt" && repositories.RepoSigningFile == "" {
			hftx.ErrorSign("You need to provide a PGP private key file (flag -k) when using the APT format")
			os.Exit(1)
		}

		// ok, let's go
		if err := repositories.CreateRepository(args[0], args[1]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var repoListSupportedCmd = &cobra.Command{
	Use:     "supported",
	Example: "nxtools repo supported [-e defaultEnv.json]",
	Short:   "Lists all repositories formats, by type, and show their support status",
	Run: func(cmd *cobra.Command, args []string) {
		repositories.ListSupportedFormats()
	},
}

var repoMigrateCmd = &cobra.Command{
	Use:     "migrate",
	Example: "nxtools repo migrate OLD_REPO NEW_REPO [-e defaultEnv.json] [-k]",
	Short:   "Migrate OLD_REPO's contents to NEW_REPO",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := assets.MigrateRepo(args[0], args[1]); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	repoCmd.AddCommand(repoListCmd, repoCreateCmd, repoDeleteCmd, repoQueryTypeCmd, repoReindexCmd, repoListSupportedCmd, repoMigrateCmd)

	repoListCmd.Flags().BoolVar(&repositories.RepoListJSONOutput, "json", false, "Output repository information as JSON")
	repoMigrateCmd.Flags().BoolVarP(&repositories.KeepSource, "keep", "k", false, "Keep source repository assets")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoFormat, "format", "f", "", "Repository format/recipe family (e.g. yum, apt, maven, docker)")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoType, "type", "t", "hosted", "Repository type (hosted, proxy, group)")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoSigningFile, "keyfile", "k", "", "Private key location")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoSigningPassphrase, "passphrase", "p", "", "Private key passphrase")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoAptDistro, "distro", "d", "nexus", "Debian-like distribution")
	repoCreateCmd.Flags().StringVarP(&repositories.StorageWritePolicy, "writepolicy", "w", "ALLOW", "Blob storage policy: ALLOW, ALLOW_ONCE, DENY")
	repoCreateCmd.Flags().BoolVarP(&repositories.StorageStrictContentValidation, "strict", "s", true, "Set this to disable strict content validation")
	repoCreateCmd.Flags().StringVarP(&repositories.MavenVersionPolicy, "versionpolicy", "v", "RELEASE", "Maven version policy")
	repoCreateCmd.Flags().StringVarP(&repositories.MavenLayoutPolicy, "layoutpolicy", "l", "STRICT", "Maven layout policy")
	repoCreateCmd.Flags().StringVarP(&repositories.RepoContentDisposition, "contentdisposition", "c", "INLINE", "Maven content disposition")
	repoCreateCmd.Flags().UintVarP(&repositories.YumRepodataDepth, "repodepth", "r", 0, "Yum repository data depth")
	repoCreateCmd.Flags().StringVarP(&repositories.YumDeployPolicy, "deploypolicy", "D", "PERMISSIVE", "Maven content disposition")
	repoCreateCmd.Flags().BoolVarP(&repositories.DockerV1Enabled, "v1api", "1", false, "Allow v1 API calls")
	repoCreateCmd.Flags().BoolVarP(&repositories.DockerForceBasicAuth, "basic", "b", false, "Allow basic http(s) auth")
	repoCreateCmd.Flags().UintVar(&repositories.DockerHttpPort, "http", 0, "HTTPS port the registry listens on")
	repoCreateCmd.Flags().UintVar(&repositories.DockerHttpsPort, "https", 0, "HTTP port the registry listens on")
	repoCreateCmd.Flags().StringVarP(&repositories.DockerSubdomain, "subdomain", "S", "", "Subdomain to route registry to")
	repoCreateCmd.Flags().BoolVarP(&repositories.DockerPathEnabled, "path", "P", false, "Path enabled")

	// Mark "a" and "b" as mutually exclusive
	repoCreateCmd.MarkFlagsMutuallyExclusive("http", "https")
	_ = repoCreateCmd.MarkFlagRequired("format")
}
