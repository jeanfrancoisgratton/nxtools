// nxtools
// Integration-style tests for RunTasks against an in-process fake Nexus server.

package tasks

import (
	"net/http"
	"strings"
	"testing"
)

func TestRunTasks_Success(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("1", "task-one", "blobstore.compact", "WAITING", nil)
	fn.addTask("2", "task-two", "blobstore.compact", "WAITING", nil)

	if err := RunTasks([]string{"1", "2"}); err != nil {
		t.Fatalf("RunTasks: %v", err)
	}
	if len(fn.runCalls) != 2 {
		t.Fatalf("expected 2 /run calls, got %d: %v", len(fn.runCalls), fn.runCalls)
	}
}

func TestRunTasks_PartialFailureIsAggregatedNotFatal(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("good", "task-good", "blobstore.compact", "WAITING", nil)
	fn.addTask("bad", "task-bad", "blobstore.compact", "WAITING", nil)
	fn.failStatus["POST /service/rest/v1/tasks/bad/run"] = http.StatusInternalServerError

	err := RunTasks([]string{"good", "bad"})
	if err == nil {
		t.Fatal("expected an aggregate error naming the failed ID")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("expected error to mention the failed ID, got: %v", err)
	}

	// The good ID must still have been attempted despite the bad one failing.
	found := false
	for _, id := range fn.runCalls {
		if id == "good" {
			found = true
		}
	}
	if !found {
		t.Error("expected the good task to still be run despite the other one failing")
	}
}
