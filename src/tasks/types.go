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
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Message       *string `json:"message"`
	CurrentState  string  `json:"currentState"`
	LastRunResult *string `json:"lastRunResult"`
	NextRun       *string `json:"nextRun"`
	LastRun       *string `json:"lastRun"`
	Schedule      *string `json:"schedule"`
}
