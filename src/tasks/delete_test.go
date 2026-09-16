// nxtools
// Integration-style tests for DeleteTasks against an in-process fake Nexus
// server. TestDeleteTasks_Success in particular locks in a real bug: the
// live Nexus server returns 204 No Content on a successful delete, despite
// the swagger doc claiming 200 — deleteTask must accept the real response,
// not the documented one.

package tasks

import (
	"net/http"
	"strings"
	"testing"
)

func TestDeleteTasks_Success(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("1", "task-one", "blobstore.compact", "WAITING", nil)

	if err := DeleteTasks([]string{"1"}); err != nil {
		t.Fatalf("DeleteTasks: %v", err)
	}
	if len(fn.deleteCalls) != 1 || fn.deleteCalls[0] != "1" {
		t.Fatalf("expected a single DELETE call for task 1, got %v", fn.deleteCalls)
	}

	remaining, err := fetchAllTasks()
	if err != nil {
		t.Fatalf("fetchAllTasks: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected the task to be gone after delete, got %+v", remaining)
	}
}

func TestDeleteTasks_MultipleIDsPartialFailure(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("good", "task-good", "blobstore.compact", "WAITING", nil)
	fn.addTask("bad", "task-bad", "blobstore.compact", "WAITING", nil)
	fn.failStatus["DELETE /service/rest/v1/tasks/bad"] = http.StatusNotFound

	err := DeleteTasks([]string{"good", "bad"})
	if err == nil {
		t.Fatal("expected an aggregate error naming the failed ID")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("expected error to mention the failed ID, got: %v", err)
	}

	remaining, ce := fetchAllTasks()
	if ce != nil {
		t.Fatalf("fetchAllTasks: %v", ce)
	}
	if len(remaining) != 1 || remaining[0].ID != "bad" {
		t.Fatalf("expected only the failed task to remain, got %+v", remaining)
	}
}
