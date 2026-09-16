// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/tasks/create.go

package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

// BlobCompactTaskParams collects everything needed to provision a
// "blobstore.compact" task.
type BlobCompactTaskParams struct {
	TaskName        string
	BlobStoreName   string
	OlderThanDays   int
	AlertEmail      string
	NotifyCondition string // "FAILURE" | "SUCCESS_FAILURE"
	Frequency       FrequencyXO
	Enabled         bool
}

// CreateBlobCompactTask provisions a "Compact blob store" task against a
// single blob store. Nexus binds a task to exactly one target at creation
// time, so this refuses to create a second compact task for a blob store that
// already has one (regardless of what the existing task happens to be named).
//
// Endpoint:
//
//	POST /v1/tasks
func CreateBlobCompactTask(p BlobCompactTaskParams) (*TaskSummary, *cerr.CustomError) {
	if ce := verifyBlobStoreExists(p.BlobStoreName); ce != nil {
		return nil, ce
	}

	existing, ce := findCompactTaskByBlobstore(p.BlobStoreName)
	if ce != nil {
		return nil, ce
	}
	if existing != nil {
		return nil, &cerr.CustomError{Title: "Task already exists",
			Message: "Blob store " + p.BlobStoreName + " already has a compact task: " + existing.Name + " (" + existing.ID + ")"}
	}

	payload := TaskTemplateXO{
		Type:                  "blobstore.compact",
		Name:                  p.TaskName,
		Enabled:               p.Enabled,
		AlertEmail:            p.AlertEmail,
		NotificationCondition: p.NotifyCondition,
		Frequency:             p.Frequency,
		Properties: map[string]string{
			"blobstoreName":  p.BlobStoreName,
			"blobsOlderThan": strconv.Itoa(p.OlderThanDays),
		},
	}

	body, e1 := json.Marshal(payload)
	if e1 != nil {
		return nil, &cerr.CustomError{Title: "Unable to build request payload", Message: e1.Error()}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	resp, e2 := c.Do(context.Background(), http.MethodPost, "/service/rest/v1/tasks", nil, bytes.NewReader(body), headers)
	if e2 != nil {
		return nil, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return nil, &cerr.CustomError{Title: "Unable to create task", Message: fmt.Sprintf("HTTP %s: %s", resp.Status, string(b))}
	}

	var created TaskSummary
	if e3 := json.NewDecoder(resp.Body).Decode(&created); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Task " + created.Name + " (" + created.ID + ") was successfully created"))
	}
	return &created, nil
}
