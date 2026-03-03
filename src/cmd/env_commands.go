// certificateManager
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/cmd/env_commands.go
// Original timestamp: 2023/09/15 13:40

package cmd

import (
	"fmt"
	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"github.com/spf13/cobra"
	"nxtools/env"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Environment sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list | add | remove }")
	},
}

var envListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Example: "nxtools env list [directory]",
	Short:   "Lists all environment files",
	Run: func(cmd *cobra.Command, args []string) {
		argument := ""
		if len(args) > 0 {
			argument = args[0]
		}
		if err := env.ListEnvironments(argument); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var envRmCmd = &cobra.Command{
	Use:     "rm",
	Aliases: []string{"remove"},
	Example: "nxtools env remove { FILE[.json] | defaultEnv.json }",
	Short:   "Removes the environment FILE",
	Run: func(cmd *cobra.Command, args []string) {
		fname := ""
		if len(args) == 0 {
			fname = "defaultEnv.json"
		} else {
			fname = args[0]
		}
		var err *cerr.CustomError
		if err = env.RemoveEnvFile(fname); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var envAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"create"},
	Example: "nxtools env add [FILE[.json]]",
	Short:   "Adds the environment FILE",
	Long: `The extension (.json) is implied and will be added if missing. Moreover, not specifying a filename
Will create a defaultEnv.json file, which is the application's default file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fname := ""
		if len(args) == 0 {
			fname = "defaultEnv.json"
		} else {
			fname = args[0]
		}
		if err := env.AddEnvFile(fname); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var envInfoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"explain"},
	Example: "nxtools env info FILE1[.json] FILE2[.json]... FILEn[.json]",
	Short:   "Prints the environment FILE[12n] information",
	Long:    `You can list as many environment files as you wish, here`,
	Run: func(cmd *cobra.Command, args []string) {
		envfiles := []string{"defaultEnv.json"}
		if len(args) != 0 {
			envfiles = args
		}
		var err *cerr.CustomError
		if err = env.ExplainEnvFile(envfiles); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	envCmd.AddCommand(envListCmd, envRmCmd, envAddCmd, envInfoCmd)
}
