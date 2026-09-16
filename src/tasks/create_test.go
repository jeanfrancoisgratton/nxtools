// nxtools
// Integration-style tests for CreateBlobCompactTask against an in-process
// fake Nexus server.

package tasks

import (
	"testing"
)

func TestCreateBlobCompactTask_BlobStoreNotFound(t *testing.T) {
	newTaskTestFixture(t)

	_, err := CreateBlobCompactTask(BlobCompactTaskParams{
		TaskName:        "my-compact-task",
		BlobStoreName:   "doesNotExist",
		NotifyCondition: "FAILURE",
		Frequency:       FrequencyXO{Schedule: "manual"},
		Enabled:         true,
	})
	if err == nil {
		t.Fatal("expected an error for a nonexistent blob store")
	}
}

func TestCreateBlobCompactTask_RefusesDuplicateTarget(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addBlobstore("pypiLocal")
	fn.addTask("1", "some-other-name", "blobstore.compact", "WAITING", map[string]string{"blobstoreName": "pypiLocal"})

	_, err := CreateBlobCompactTask(BlobCompactTaskParams{
		TaskName:        "my-compact-task",
		BlobStoreName:   "pypiLocal",
		NotifyCondition: "FAILURE",
		Frequency:       FrequencyXO{Schedule: "manual"},
		Enabled:         true,
	})
	if err == nil {
		t.Fatal("expected an error when a compact task already targets this blob store")
	}
	if len(fn.createCalls) != 0 {
		t.Errorf("expected no POST /v1/tasks call once a duplicate is detected, got %d", len(fn.createCalls))
	}
}

func TestCreateBlobCompactTask_Success(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addBlobstore("pypiLocal")

	created, err := CreateBlobCompactTask(BlobCompactTaskParams{
		TaskName:        "my-compact-task",
		BlobStoreName:   "pypiLocal",
		OlderThanDays:   7,
		AlertEmail:      "ops@example.com",
		NotifyCondition: "SUCCESS_FAILURE",
		Frequency:       FrequencyXO{Schedule: "manual"},
		Enabled:         true,
	})
	if err != nil {
		t.Fatalf("CreateBlobCompactTask: %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected a created task with an ID, got %+v", created)
	}

	if len(fn.createCalls) != 1 {
		t.Fatalf("expected exactly one POST /v1/tasks call, got %d", len(fn.createCalls))
	}
	sent := fn.createCalls[0]
	if sent.Type != "blobstore.compact" {
		t.Errorf("Type = %q, want blobstore.compact", sent.Type)
	}
	if sent.Name != "my-compact-task" {
		t.Errorf("Name = %q, want my-compact-task", sent.Name)
	}
	if !sent.Enabled {
		t.Error("expected Enabled to be true")
	}
	if sent.AlertEmail != "ops@example.com" {
		t.Errorf("AlertEmail = %q, want ops@example.com", sent.AlertEmail)
	}
	if sent.NotificationCondition != "SUCCESS_FAILURE" {
		t.Errorf("NotificationCondition = %q, want SUCCESS_FAILURE", sent.NotificationCondition)
	}
	if sent.Properties["blobstoreName"] != "pypiLocal" {
		t.Errorf("properties.blobstoreName = %q, want pypiLocal", sent.Properties["blobstoreName"])
	}
	if sent.Properties["blobsOlderThan"] != "7" {
		t.Errorf("properties.blobsOlderThan = %q, want \"7\"", sent.Properties["blobsOlderThan"])
	}
}
