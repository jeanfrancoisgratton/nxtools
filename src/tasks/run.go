// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/09/16 21:00
// Original filename: src/tasks/run.go

package tasks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

// RunTasks runs each task in ids now. It continues past individual failures
// so one bad ID doesn't block the rest, and returns an aggregate error naming
// every ID that failed.
//
// Endpoint (per ID):
//
//	POST /v1/tasks/{id}/run
func RunTasks(ids []string) *cerr.CustomError {
	return applyToTasks(ids, "run", "was successfully started")
}

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

func applyToTasks(ids []string, action, successMsg string) *cerr.CustomError {
	var failed []string

	for _, id := range ids {
		if ce := taskAction(id, action); ce != nil {
			if !shared.QuietOutput {
				fmt.Println(hftx.ErrorSign("Task " + id + ": " + ce.Error()))
			}
			failed = append(failed, id)
			continue
		}
		if !shared.QuietOutput {
			fmt.Println(hftx.EnabledSign("Task " + id + " " + successMsg))
		}
	}

	if len(failed) > 0 {
		return &cerr.CustomError{Title: "Unable to " + action + " some tasks", Message: "Failed IDs: " + strings.Join(failed, ", ")}
	}
	return nil
}

func taskAction(id, action string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	resp, e2 := c.Do(context.Background(), http.MethodPost, "/service/rest/v1/tasks/"+id+"/"+action, nil, nil, nil)
	if e2 != nil {
		return e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return &cerr.CustomError{Title: "Unable to " + action + " task", Message: "HTTP status code: " + resp.Status}
	}
	return nil
}
