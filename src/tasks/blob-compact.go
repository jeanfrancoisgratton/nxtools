// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/tasks/blob-compact.go

package tasks

import (
	"strings"

	"nxtools/blobstores"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
)

// CompactBlobStore runs the existing "blobstore.compact" task targeting
// blobstoreName. It hard-fails if the blob store doesn't exist, if no compact
// task targets it, or if that task is already running.
//
// Endpoint (via RunTasks):
//
//	POST /v1/tasks/{id}/run
func CompactBlobStore(blobstoreName string) *cerr.CustomError {
	if ce := verifyBlobStoreExists(blobstoreName); ce != nil {
		return ce
	}

	task, ce := findCompactTaskByBlobstore(blobstoreName)
	if ce != nil {
		return ce
	}
	if task == nil {
		return &cerr.CustomError{Title: "Task not found",
			Message: "No blobstore.compact task targets blob store " + blobstoreName +
				"; create one first with `task create blob-compact`"}
	}
	if strings.EqualFold(task.CurrentState, "RUNNING") {
		return &cerr.CustomError{Title: "Task already running",
			Message: "Task " + task.Name + " (" + task.ID + ") is currently running"}
	}

	return RunTasks([]string{task.ID})
}

// verifyBlobStoreExists checks that blobstoreName is a real blob store on the
// server. A blob store has no format, so a single list-and-match works
// regardless of its type (file, s3, azure, group...).
func verifyBlobStoreExists(blobstoreName string) *cerr.CustomError {
	blobs, err := blobstores.ListBlobs(false)
	if err != nil {
		return err
	}
	for _, b := range blobs {
		if b.Name == blobstoreName {
			return nil
		}
	}
	return &cerr.CustomError{Title: "Blob store not found", Message: "No blob store named " + blobstoreName + " exists"}
}

// findCompactTaskByBlobstore returns the single "blobstore.compact" task whose
// properties.blobstoreName matches blobstoreName. It matches by actual target,
// not by task name, since a task's display name is arbitrary and there's no
// naming convention nxtools can rely on for tasks it didn't create itself.
// Returns (nil, nil) if no such task exists.
func findCompactTaskByBlobstore(blobstoreName string) (*TaskSummary, *cerr.CustomError) {
	items, err := fetchTasks("blobstore.compact")
	if err != nil {
		return nil, err
	}

	var matches []TaskSummary
	for _, t := range items {
		if t.Properties["blobstoreName"] == blobstoreName {
			matches = append(matches, t)
		}
	}

	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return &matches[0], nil
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name + " (" + m.ID + ")"
		}
		return nil, &cerr.CustomError{Title: "Server misconfiguration",
			Message: "Multiple blobstore.compact tasks target " + blobstoreName + ": " + strings.Join(names, ", ")}
	}
}
