// nxtools
// Unit tests for list.go's pure helpers, plus integration-style tests for
// fetchTasks/ListTasks against an in-process fake Nexus server.

package tasks

import (
	"strings"
	"testing"
)

func TestStateLabel(t *testing.T) {
	if !strings.Contains(stateLabel("RUNNING"), "RUNNING") {
		t.Error("expected RUNNING state to be preserved in the label")
	}
	if !strings.Contains(stateLabel("running"), "running") {
		t.Error("expected lowercase running to still be recognized")
	}
	if !strings.Contains(stateLabel("WAITING"), "WAITING") {
		t.Error("expected WAITING state to be preserved in the label")
	}
	if stateLabel("SOME_OTHER_STATE") != "SOME_OTHER_STATE" {
		t.Errorf("expected unrecognized states to pass through unchanged, got %q", stateLabel("SOME_OTHER_STATE"))
	}
}

func TestTruncateName(t *testing.T) {
	short := "short-name"
	if got := truncateName(short); got != short {
		t.Errorf("truncateName(%q) = %q, want unchanged", short, got)
	}

	long := strings.Repeat("x", maxNameLength+10)
	got := truncateName(long)
	if len(got) != maxNameLength {
		t.Errorf("truncateName should cap length at %d, got %d", maxNameLength, len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("truncated name should end with '...', got %q", got)
	}
}

func TestDeref(t *testing.T) {
	if got := deref(nil); got != "" {
		t.Errorf("deref(nil) = %q, want empty string", got)
	}
	s := "hello"
	if got := deref(&s); got != "hello" {
		t.Errorf("deref(&s) = %q, want %q", got, s)
	}
}

func TestFormatTimestamp(t *testing.T) {
	if got := formatTimestamp(nil); got != "" {
		t.Errorf("formatTimestamp(nil) = %q, want empty string", got)
	}

	valid := "2026-09-16T05:00:00.022+00:00"
	got := formatTimestamp(&valid)
	want := "2026-09-16 05:00:00"
	if got != want {
		t.Errorf("formatTimestamp(%q) = %q, want %q", valid, got, want)
	}

	invalid := "not-a-timestamp"
	if got := formatTimestamp(&invalid); got != invalid {
		t.Errorf("formatTimestamp should pass through unparseable input unchanged, got %q", got)
	}
}

func TestFetchTasks_TypeFilter(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("1", "_reindex_aptLocal", "repository.apt.rebuild.metadata", "WAITING", nil)
	fn.addTask("2", "compact-A", "blobstore.compact", "WAITING", map[string]string{"blobstoreName": "A"})
	fn.addTask("3", "compact-B", "blobstore.compact", "RUNNING", map[string]string{"blobstoreName": "B"})

	items, err := fetchTasks("blobstore.compact")
	if err != nil {
		t.Fatalf("fetchTasks: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 blobstore.compact tasks, got %d", len(items))
	}
	for _, it := range items {
		if it.Type != "blobstore.compact" {
			t.Errorf("unexpected task type leaked through filter: %q", it.Type)
		}
	}
}

func TestFetchAllTasks_NoFilter(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("1", "a", "typeA", "WAITING", nil)
	fn.addTask("2", "b", "typeB", "WAITING", nil)

	items, err := fetchAllTasks()
	if err != nil {
		t.Fatalf("fetchAllTasks: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(items))
	}
}

func TestListTasks_RunningOnly(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addTask("1", "idle-task", "typeA", "WAITING", nil)
	fn.addTask("2", "active-task", "typeB", "RUNNING", nil)

	all, err := ListTasks(false, false)
	if err != nil {
		t.Fatalf("ListTasks(all): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks total, got %d", len(all))
	}

	running, err := ListTasks(false, true)
	if err != nil {
		t.Fatalf("ListTasks(running only): %v", err)
	}
	if len(running) != 1 || running[0].ID != "2" {
		t.Fatalf("expected only the running task to survive the filter, got %+v", running)
	}
}
