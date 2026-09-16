// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/tasks/delete.go

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

// DeleteTasks deletes each task in ids. It continues past individual failures
// so one bad ID doesn't block the rest, and returns an aggregate error naming
// every ID that failed.
//
// This doesn't share applyToTasks/taskAction (used by RunTasks/StopTasks)
// because deletion is a different verb shape: DELETE instead of POST, against
// /v1/tasks/{id} directly rather than /v1/tasks/{id}/{action}, and Nexus
// reports success as 200 here instead of 204.
//
// Endpoint (per ID):
//
//	DELETE /v1/tasks/{id}
func DeleteTasks(ids []string) *cerr.CustomError {
	var failed []string

	for _, id := range ids {
		if ce := deleteTask(id); ce != nil {
			if !shared.QuietOutput {
				fmt.Println(hftx.ErrorSign("Task " + id + ": " + ce.Error()))
			}
			failed = append(failed, id)
			continue
		}
		if !shared.QuietOutput {
			fmt.Println(hftx.EnabledSign("Task " + id + " was successfully deleted"))
		}
	}

	if len(failed) > 0 {
		return &cerr.CustomError{Title: "Unable to delete some tasks", Message: "Failed IDs: " + strings.Join(failed, ", ")}
	}
	return nil
}

func deleteTask(id string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	resp, e2 := c.Do(context.Background(), http.MethodDelete, "/service/rest/v1/tasks/"+id, nil, nil, nil)
	if e2 != nil {
		return e2
	}
	defer resp.Body.Close()

	// The swagger doc claims 200, but the live server actually returns 204
	// No Content on success; accept any 2xx rather than trust the doc.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to delete task", Message: "HTTP status code: " + resp.Status}
	}
	return nil
}
