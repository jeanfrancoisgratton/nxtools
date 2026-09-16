// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/09/16 20:00
// Original filename: src/cmd/tasks_commands.go

package cmd

import (
	"fmt"
	"os"

	"nxtools/tasks"

	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Task-related sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list | show | run | stop }")
	},
}

var taskListRunningOnly bool

var taskListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Example: "nxtools task list [-e defaultEnv.json] [--running]",
	Short:   "Lists all tasks defined on the server",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := tasks.ListTasks(true, taskListRunningOnly); err != nil {
			fmt.Println(err.Error())
		}
	},
}

var taskRunCmd = &cobra.Command{
	Use:     "run TASK_ID [TASK_ID2 ...]",
	Aliases: []string{"start"},
	Example: "nxtools task run [-e defaultEnv.json] TASK_ID [TASK_ID2 ...]",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Runs one or more tasks now",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tasks.RunTasks(args); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var taskStopCmd = &cobra.Command{
	Use:     "stop TASK_ID [TASK_ID2 ...]",
	Aliases: []string{"kill"},
	Example: "nxtools task stop [-e defaultEnv.json] TASK_ID [TASK_ID2 ...]",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Stops one or more running tasks",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tasks.StopTasks(args); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	taskCmd.AddCommand(taskListCmd, taskRunCmd, taskStopCmd)

	taskListCmd.Flags().BoolVarP(&taskListRunningOnly, "running", "r", false, "Only list tasks that are currently running")
}
