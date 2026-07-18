// nxtools
// Unit tests for repository detail helpers that don't require a live server.

package repositories

import "testing"

func TestEnsureUploadableRepository(t *testing.T) {
	if err := EnsureUploadableRepository(nil); err == nil {
		t.Error("nil repository should be rejected")
	}

	if err := EnsureUploadableRepository(&RepositorySummary{Name: "r", Type: "hosted"}); err != nil {
		t.Errorf("hosted repo should be uploadable: %v", err)
	}

	// Case-insensitive type match.
	if err := EnsureUploadableRepository(&RepositorySummary{Name: "r", Type: "HOSTED"}); err != nil {
		t.Errorf("HOSTED (upper) repo should be uploadable: %v", err)
	}

	for _, typ := range []string{"proxy", "group", ""} {
		if err := EnsureUploadableRepository(&RepositorySummary{Name: "r", Type: typ}); err == nil {
			t.Errorf("type %q should not be uploadable", typ)
		}
	}
}
