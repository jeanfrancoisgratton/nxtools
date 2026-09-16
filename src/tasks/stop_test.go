// nxtools
// Integration-style tests for StopTasks against an in-process fake Nexus server.

package tasks

import (
	"net/http"
	"strings"
	"testing"
)

func TestStopTasks_Success(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("1", "task-one", "blobstore.compact", "RUNNING", nil)

	if err := StopTasks([]string{"1"}); err != nil {
		t.Fatalf("StopTasks: %v", err)
	}
	if len(fn.stopCalls) != 1 || fn.stopCalls[0] != "1" {
		t.Fatalf("expected a single /stop call for task 1, got %v", fn.stopCalls)
	}
}

func TestStopTasks_FailureIsReported(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("1", "task-one", "blobstore.compact", "WAITING", nil)
	fn.failStatus["POST /service/rest/v1/tasks/1/stop"] = http.StatusNotFound

	err := StopTasks([]string{"1"})
	if err == nil {
		t.Fatal("expected an error when /stop fails")
	}
	if !strings.Contains(err.Error(), "1") {
		t.Errorf("expected error to mention the failed ID, got: %v", err)
	}
}
