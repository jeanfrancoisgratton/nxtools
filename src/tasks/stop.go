// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/tasks/stop.go
// Original timestamp : 2026.09.16 02:45:06

package tasks

import cerr "github.com/jeanfrancoisgratton/customError/v3"

// StopTasks stops each running task in ids. It continues past individual
// failures so one bad ID doesn't block the rest, and returns an aggregate
// error naming every ID that failed.
//
// Endpoint (per ID):
//
//	POST /v1/tasks/{id}/stop
func StopTasks(ids []string) *cerr.CustomError {
	return applyToTasks(ids, "stop", "was successfully stopped")
}
