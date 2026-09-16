// nxtools
// Original filename: src/tasks/helpers_test.go

package tasks

import "testing"

func TestTaskNameByID(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("task-1", "Compact pypiLocal", "blobstore.compact", "WAITING", nil)

	name, ce := TaskNameByID("task-1")
	if ce != nil {
		t.Fatalf("unexpected error: %v", ce)
	}
	if name != "Compact pypiLocal" {
		t.Fatalf("expected %q, got %q", "Compact pypiLocal", name)
	}
}

func TestTaskNameByID_NotFound(t *testing.T) {
	newTaskTestFixture(t)

	_, ce := TaskNameByID("does-not-exist")
	if ce == nil {
		t.Fatal("expected error for unknown task ID, got nil")
	}
}
