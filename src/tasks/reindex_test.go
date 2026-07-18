// nxtools
// Unit tests for reindex task resolution that don't require a live server.

package tasks

import (
	"strings"
	"testing"
)

func TestFormatsWithoutReindexTask(t *testing.T) {
	if !formatsWithoutReindexTask["alpine"] {
		t.Error("alpine should be marked as having no reindex task")
	}
	for _, f := range []string{"apt", "yum", "docker", "maven"} {
		if formatsWithoutReindexTask[f] {
			t.Errorf("%q should not be in formatsWithoutReindexTask", f)
		}
	}
}

func TestFindTaskByRepoName_UnsupportedFormat(t *testing.T) {
	// An unknown format short-circuits before any network call.
	_, err := findTaskByRepoName("bogusformat", "somerepo")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "bogusformat") {
		t.Fatalf("error should mention the format, got: %v", err)
	}
}
