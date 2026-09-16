// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/09/16 20:00
// Original filename: src/cmd/tasks_commands.go

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"nxtools/tasks"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Task-related sub-command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { list | run | stop | create }")
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

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new scheduled task",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid subcommands are: { blob-compact }")
	},
}

var (
	blobCompactOlderThan int
	blobCompactEmail     string
	blobCompactNotify    string
	blobCompactSchedule  string
	blobCompactStartDate string
	blobCompactTZ        string
	blobCompactDays      []int
	blobCompactCron      string
	blobCompactDisabled  bool
)

var taskCreateBlobCompactCmd = &cobra.Command{
	Use:     "blob-compact TASK_NAME BLOBSTORE_NAME",
	Example: "nxtools task create blob-compact [-e defaultEnv.json] MyCompactTask pypiLocal --schedule weekly --start-date 2026-09-20T02:00 --days 1,4",
	Short:   "Creates a \"Compact blob store\" task",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if blobCompactOlderThan < 0 {
			fmt.Println(hftx.ErrorSign("--older cannot be negative"))
			os.Exit(1)
		}

		freq, ce := buildFrequency(blobCompactSchedule, blobCompactStartDate, blobCompactTZ, blobCompactDays, blobCompactCron)
		if ce != nil {
			fmt.Println(ce.Error())
			os.Exit(1)
		}

		notifyCond, ce := normalizeNotifyCondition(blobCompactNotify)
		if ce != nil {
			fmt.Println(ce.Error())
			os.Exit(1)
		}
		if blobCompactEmail == "" && cmd.Flags().Changed("notify") {
			fmt.Println(hftx.WarningSign("--notify has no effect without --email"))
		}

		params := tasks.BlobCompactTaskParams{
			TaskName:        args[0],
			BlobStoreName:   args[1],
			OlderThanDays:   blobCompactOlderThan,
			AlertEmail:      blobCompactEmail,
			NotifyCondition: notifyCond,
			Frequency:       freq,
			Enabled:         !blobCompactDisabled,
		}
		if _, err := tasks.CreateBlobCompactTask(params); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var validSchedules = map[string]bool{
	"manual": true, "once": true, "hourly": true, "daily": true,
	"weekly": true, "monthly": true, "cron": true,
}

// buildFrequency validates the schedule-related flags and turns them into a
// tasks.FrequencyXO. Nexus's schedule shapes each require a different subset
// of fields (e.g. "cron" needs a cron expression, "weekly" needs recurring
// days), so validation happens here rather than leaving it to the server.
func buildFrequency(schedule, startDate, tz string, days []int, cron string) (tasks.FrequencyXO, *cerr.CustomError) {
	schedule = strings.ToLower(strings.TrimSpace(schedule))
	if !validSchedules[schedule] {
		return tasks.FrequencyXO{}, &cerr.CustomError{Title: "Invalid --schedule value",
			Message: schedule + " is not supported; use one of: manual, once, hourly, daily, weekly, monthly, cron"}
	}

	freq := tasks.FrequencyXO{Schedule: schedule}
	if schedule == "manual" {
		if startDate != "" || tz != "" || len(days) > 0 || cron != "" {
			fmt.Println(hftx.WarningSign("--start-date, --tz, --days and --cron are ignored for the manual schedule"))
		}
		return freq, nil
	}

	if schedule == "cron" {
		if cron == "" {
			return tasks.FrequencyXO{}, &cerr.CustomError{Title: "Missing --cron", Message: "The cron schedule requires --cron"}
		}
		freq.CronExpression = &cron
	} else {
		if cron != "" {
			fmt.Println(hftx.WarningSign("--cron is ignored for the " + schedule + " schedule"))
		}
		if startDate == "" {
			return tasks.FrequencyXO{}, &cerr.CustomError{Title: "Missing --start-date",
				Message: "The " + schedule + " schedule requires --start-date"}
		}
		ms, ce := parseStartDate(startDate)
		if ce != nil {
			return tasks.FrequencyXO{}, ce
		}
		freq.StartDate = &ms
	}

	if schedule == "weekly" || schedule == "monthly" {
		if len(days) == 0 {
			return tasks.FrequencyXO{}, &cerr.CustomError{Title: "Missing --days",
				Message: "The " + schedule + " schedule requires --days"}
		}
		maxDay := 7
		if schedule == "monthly" {
			maxDay = 31
		}
		for _, d := range days {
			if schedule == "monthly" && d == 999 {
				continue
			}
			if d < 1 || d > maxDay {
				return tasks.FrequencyXO{}, &cerr.CustomError{Title: "Invalid --days value",
					Message: fmt.Sprintf("%d is out of range for a %s schedule", d, schedule)}
			}
		}
		freq.RecurringDays = days
	} else if len(days) > 0 {
		fmt.Println(hftx.WarningSign("--days is ignored for the " + schedule + " schedule"))
	}

	offset := tz
	if offset == "" {
		offset = time.Now().Format("-07:00")
	}
	freq.TimeZoneOffset = &offset

	return freq, nil
}

// parseStartDate accepts a handful of reasonably human-friendly layouts and
// converts them to unix milliseconds, the unit Nexus's task-creation API uses.
func parseStartDate(s string) (int64, *cerr.CustomError) {
	layouts := []string{"2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.UnixMilli(), nil
		}
	}
	return 0, &cerr.CustomError{Title: "Invalid --start-date",
		Message: s + " doesn't match any supported format (2006-01-02, \"2006-01-02 15:04\", 2006-01-02T15:04)"}
}

// normalizeNotifyCondition maps the CLI's --notify vocabulary to the values
// Nexus's API actually accepts.
func normalizeNotifyCondition(s string) (string, *cerr.CustomError) {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "")) {
	case "failure":
		return "FAILURE", nil
	case "success+failure":
		return "SUCCESS_FAILURE", nil
	default:
		return "", &cerr.CustomError{Title: "Invalid --notify value", Message: s + " must be one of: failure, success+failure"}
	}
}

func init() {
	taskCmd.AddCommand(taskListCmd, taskRunCmd, taskStopCmd, taskCreateCmd)
	taskCreateCmd.AddCommand(taskCreateBlobCompactCmd)

	taskListCmd.Flags().BoolVarP(&taskListRunningOnly, "running", "r", false, "Only list tasks that are currently running")

	taskCreateBlobCompactCmd.Flags().IntVarP(&blobCompactOlderThan, "older", "o", 0, "Only compact blobs older than this many days")
	taskCreateBlobCompactCmd.Flags().StringVarP(&blobCompactEmail, "email", "m", "", "E-mail address for task notifications")
	taskCreateBlobCompactCmd.Flags().StringVarP(&blobCompactNotify, "notify", "n", "failure", "When to notify: failure, success+failure (no effect without --email)")
	taskCreateBlobCompactCmd.Flags().StringVarP(&blobCompactSchedule, "schedule", "s", "manual", "Schedule type: manual, once, hourly, daily, weekly, monthly, cron")
	taskCreateBlobCompactCmd.Flags().StringVar(&blobCompactStartDate, "start-date", "", "Start date/time (2006-01-02, \"2006-01-02 15:04\" or 2006-01-02T15:04); required unless --schedule=manual")
	taskCreateBlobCompactCmd.Flags().StringVar(&blobCompactTZ, "tz", "", "Timezone offset, e.g. -05:00 (defaults to the local machine's current offset)")
	taskCreateBlobCompactCmd.Flags().IntSliceVar(&blobCompactDays, "days", nil, "Recurring days: 1-7 for weekly (1=Sunday), 1-31 or 999 (last day) for monthly")
	taskCreateBlobCompactCmd.Flags().StringVar(&blobCompactCron, "cron", "", "Cron expression; required when --schedule=cron")
	taskCreateBlobCompactCmd.Flags().BoolVar(&blobCompactDisabled, "disabled", false, "Create the task in a disabled state")
}
