// nxtools
// Integration-style tests for CompactBlobStore and its helpers against an
// in-process fake Nexus server.

package tasks

import (
	"strings"
	"testing"
)

func TestVerifyBlobStoreExists(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addBlobstore("pypiLocal")

	if err := verifyBlobStoreExists("pypiLocal"); err != nil {
		t.Errorf("expected pypiLocal to be found, got error: %v", err)
	}
	if err := verifyBlobStoreExists("doesNotExist"); err == nil {
		t.Error("expected an error for a nonexistent blob store")
	}
}

func TestFindCompactTaskByBlobstore(t *testing.T) {
	t.Run("none found", func(t *testing.T) {
		newTaskTestFixture(t)
		task, err := findCompactTaskByBlobstore("pypiLocal")
		if err != nil {
			t.Fatalf("expected no error when no task exists, got: %v", err)
		}
		if task != nil {
			t.Fatalf("expected nil task, got %+v", task)
		}
	})

	t.Run("single match, found by target not name", func(t *testing.T) {
		fn := newTaskTestFixture(t)
		fn.addTask("1", "some-arbitrary-name-set-in-the-webUI", "blobstore.compact", "WAITING",
			map[string]string{"blobstoreName": "pypiLocal"})

		task, err := findCompactTaskByBlobstore("pypiLocal")
		if err != nil {
			t.Fatalf("findCompactTaskByBlobstore: %v", err)
		}
		if task == nil || task.ID != "1" {
			t.Fatalf("expected to find task 1 by its target property, got %+v", task)
		}
	})

	t.Run("multiple matches is a hard error", func(t *testing.T) {
		fn := newTaskTestFixture(t)
		fn.addTask("1", "compact-1", "blobstore.compact", "WAITING", map[string]string{"blobstoreName": "pypiLocal"})
		fn.addTask("2", "compact-2", "blobstore.compact", "WAITING", map[string]string{"blobstoreName": "pypiLocal"})

		_, err := findCompactTaskByBlobstore("pypiLocal")
		if err == nil {
			t.Fatal("expected an error when multiple compact tasks target the same blob store")
		}
	})
}

func TestCompactBlobStore_BlobStoreNotFound(t *testing.T) {
	newTaskTestFixture(t)

	err := CompactBlobStore("doesNotExist")
	if err == nil {
		t.Fatal("expected an error for a nonexistent blob store")
	}
	if !strings.Contains(err.Error(), "doesNotExist") {
		t.Errorf("expected error to mention the blob store name, got: %v", err)
	}
}

func TestCompactBlobStore_NoTaskConfigured(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addBlobstore("pypiLocal")

	err := CompactBlobStore("pypiLocal")
	if err == nil {
		t.Fatal("expected an error when no compact task targets the blob store")
	}
}

func TestCompactBlobStore_AlreadyRunningHardFails(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addBlobstore("pypiLocal")
	fn.addTask("1", "compact-pypi", "blobstore.compact", "RUNNING", map[string]string{"blobstoreName": "pypiLocal"})

	err := CompactBlobStore("pypiLocal")
	if err == nil {
		t.Fatal("expected an error when the task is already running")
	}
	if len(fn.runCalls) != 0 {
		t.Errorf("expected /run to never be called for an already-running task, got calls: %v", fn.runCalls)
	}
}

func TestCompactBlobStore_Success(t *testing.T) {
	fn := newTaskTestFixture(t)
	fn.addBlobstore("pypiLocal")
	fn.addTask("42", "compact-pypi", "blobstore.compact", "WAITING", map[string]string{"blobstoreName": "pypiLocal"})

	if err := CompactBlobStore("pypiLocal"); err != nil {
		t.Fatalf("CompactBlobStore: %v", err)
	}
	if len(fn.runCalls) != 1 || fn.runCalls[0] != "42" {
		t.Fatalf("expected task 42 to be run, got calls: %v", fn.runCalls)
	}
}
