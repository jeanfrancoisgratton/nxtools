// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/10 01:05
// Original filename: src/tasks/types.go

package tasks

// maxNameLength caps the Name column width so the table stays readable in a
// standard 80/120-column terminal; longer names are truncated with "...".
const maxNameLength = 70

type ListTasksResponse struct {
	Items             []TaskSummary `json:"items"`
	ContinuationToken *string       `json:"continuationToken"`
}

type TaskSummary struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Message       *string           `json:"message"`
	CurrentState  string            `json:"currentState"`
	LastRunResult *string           `json:"lastRunResult"`
	NextRun       *string           `json:"nextRun"`
	LastRun       *string           `json:"lastRun"`
	Schedule      *string           `json:"schedule"`
	Properties    map[string]string `json:"properties"`
}

// FrequencyXO describes how often a task runs, as accepted by POST/PUT /v1/tasks.
// Schedule is one of: "manual", "once", "hourly", "daily", "weekly", "monthly", "cron".
type FrequencyXO struct {
	Schedule       string  `json:"schedule"`
	StartDate      *int64  `json:"startDate,omitempty"`      // unix ms; unused for "manual"
	TimeZoneOffset *string `json:"timeZoneOffset,omitempty"` // e.g. "-05:00"
	RecurringDays  []int   `json:"recurringDays,omitempty"`  // weekly: 1-7 (1=Sunday); monthly: 1-31, 999=last day
	CronExpression *string `json:"cronExpression,omitempty"` // "cron" schedule only
}

// TaskTemplateXO is the payload for POST /v1/tasks (create) and PUT /v1/tasks/{id} (update).
type TaskTemplateXO struct {
	Type                  string            `json:"type"`
	Name                  string            `json:"name"`
	Enabled               bool              `json:"enabled"`
	AlertEmail            string            `json:"alertEmail,omitempty"`
	NotificationCondition string            `json:"notificationCondition"` // "FAILURE" | "SUCCESS_FAILURE"
	Frequency             FrequencyXO       `json:"frequency"`
	Properties            map[string]string `json:"properties,omitempty"`
}
