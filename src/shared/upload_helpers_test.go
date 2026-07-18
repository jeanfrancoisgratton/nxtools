// nxtools
// Unit tests for the pure helpers in the shared package.

package shared

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatSize(t *testing.T) {
	cases := []struct {
		bytes    int64
		expected string
	}{
		{1_500_000, "1.500 MB"},
		{5_000_000, "5.000 MB"},
		{999_000_000, "999.000 MB"},
		{2_500_000_000, "2.500 GB"},
		{5_000_000_000, "5.000 GB"},
	}
	for _, c := range cases {
		if got := FormatSize(c.bytes); got != c.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", c.bytes, got, c.expected)
		}
	}
}

func TestNormalizeDirectory(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"", "/"},
		{".", "/"},
		{"   ", "/"},
		{"/", "/"},
		{"foo", "/foo"},
		{"/foo/", "/foo"},
		{"/foo/bar/", "/foo/bar"},
		{"foo\\bar", "/foo/bar"},
		{"  spaced  ", "/spaced"},
	}
	for _, c := range cases {
		if got := NormalizeDirectory(c.in); got != c.expected {
			t.Errorf("NormalizeDirectory(%q) = %q, want %q", c.in, got, c.expected)
		}
	}
}

func TestBuildRepositoryAssetPath(t *testing.T) {
	cases := []struct {
		name      string
		repo      string
		directory string
		filename  string
		expected  string
	}{
		{"root dir", "myrepo", "/", "file.rpm", "/repository/myrepo/file.rpm"},
		{"empty dir", "myrepo", "", "file.rpm", "/repository/myrepo/file.rpm"},
		{"nested dir", "myrepo", "/dir/sub", "file.rpm", "/repository/myrepo/dir/sub/file.rpm"},
		{"backslash dir", "myrepo", "dir\\sub", "file.rpm", "/repository/myrepo/dir/sub/file.rpm"},
		{"escaped repo", "my repo", "/", "a b.rpm", "/repository/my%20repo/a%20b.rpm"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildRepositoryAssetPath(c.repo, c.directory, c.filename); got != c.expected {
				t.Fatalf("buildRepositoryAssetPath = %q, want %q", got, c.expected)
			}
		})
	}
}

func TestCreateMultipartBody(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "pkg.apk")
	if err := os.WriteFile(fp, []byte("PACKAGE-CONTENT"), 0644); err != nil {
		t.Fatal(err)
	}

	body, contentType, cerr := createMultipartBody("alpine.asset", fp, map[string]string{"note": "hello"})
	if cerr != nil {
		t.Fatalf("unexpected error: %v", cerr)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		t.Fatalf("content type = %q", contentType)
	}

	raw, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	got := string(raw)
	for _, want := range []string{`name="note"`, "hello", `name="alpine.asset"`, "pkg.apk", "PACKAGE-CONTENT"} {
		if !strings.Contains(got, want) {
			t.Errorf("multipart body missing %q", want)
		}
	}
}

func TestCreateMultipartBody_Errors(t *testing.T) {
	if _, _, err := createMultipartBody("f", "   ", nil); err == nil {
		t.Error("expected error for empty file path")
	}
	if _, _, err := createMultipartBody("f", filepath.Join(t.TempDir(), "missing"), nil); err == nil {
		t.Error("expected error for missing file")
	}
}
