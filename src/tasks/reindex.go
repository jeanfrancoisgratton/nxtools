// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/10 01:05
// Original filename: src/tasks/reindex.go

package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/repositories"
	"nxtools/rest"
	"nxtools/shared"
)

// Reindex tasks should follow this naming convention:
// NAME = "_reindex_$REPONAME"
//
// ... there are reasons:
// - the name of the repo is needed to look up the repo format.
// - once the repo format is set, we can build the query parameter to fetch the list of repos

func ReindexRepo(reponame string) *cerr.CustomError {
	var id string
	var ce *cerr.CustomError

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	if id, ce = findTaskByRepoName(reponame); ce != nil {
		return ce
	}

	resp, e2 := c.Do(context.Background(), http.MethodPost, "service/rest/v1/tasks/"+id+"/run", nil, nil, nil)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return &cerr.CustomError{Title: "Unable to list blobs", Message: "HTTP status code: " + resp.Status}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Repository " + reponame + " was successfully reindexed"))
	}
	return nil
}

func findTaskByRepoName(reponame string) (string, *cerr.CustomError) {
	var taskType, repoFormat, taskID string
	var err *cerr.CustomError
	if repoFormat, err = repositories.QueryRepoType(reponame); err != nil {
		return "", err
	}

	switch repoFormat {
	case "yum":
		taskType = "repository.yum.rebuild.metadata"
	case "apt":
		taskType = "repository.apt.rebuild.metadata"
	default:
		return "", &cerr.CustomError{Title: "Unknown repository format", Message: "Repo format " + repoFormat + " is either unsupported or invalid"}
	}
	taskID, err = lookupTaskID(taskType, reponame)
	return taskID, err
}

// lookupTaskID : now that we have the task type, we can find the task ID
// If multiple IDs match the tasktype + reponame, we return an error as we should have one and only one such task

func lookupTaskID(taskType, reponame string) (string, *cerr.CustomError) {
	var ids []string

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("type", taskType)

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/tasks", q, nil, nil)
	if e2 != nil {
		return "", e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", &cerr.CustomError{Title: "Unable to list tasks", Message: "HTTP status code: " + resp.Status}
	}

	var taskResp ListTasksResponse
	dec := json.NewDecoder(resp.Body)
	if e3 := dec.Decode(&taskResp); e3 != nil {
		return "", &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	for _, task := range taskResp.Items {
		if task.Type == taskType && task.Name == "_reindex_"+reponame {
			ids = append(ids, task.ID)
		}
	}
	if len(ids) == 0 {
		return "", &cerr.CustomError{Title: "Task not found", Message: "No " + taskType + " task associated with the repository " + reponame + " was found"}
	}
	if len(ids) > 1 {
		return "", &cerr.CustomError{Title: "Server misconfiguration", Message: "Multiple " + taskType + " tasks associated with the repository " + reponame + " were found"}
	}
	return ids[0], nil
}
