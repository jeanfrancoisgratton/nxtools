// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/tasks/helpers.go

package tasks

import (
	"context"
	"encoding/json"
	"net/http"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/rest"
	"nxtools/shared"
)

// TaskNameByID looks up a single task by ID and returns its Name.
//
// Endpoint:
//
//	GET /v1/tasks/{id}
func TaskNameByID(id string) (string, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return "", err
	}

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/tasks/"+id, nil, nil, nil)
	if e2 != nil {
		return "", e2
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", &cerr.CustomError{Title: "Task not found", Message: "No task with ID " + id + " exists"}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &cerr.CustomError{Title: "Unable to get task", Message: "HTTP status code: " + resp.Status}
	}

	var task TaskSummary
	if e3 := json.NewDecoder(resp.Body).Decode(&task); e3 != nil {
		return "", &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}
	return task.Name, nil
}
