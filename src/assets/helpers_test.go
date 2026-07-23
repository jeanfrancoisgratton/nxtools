// nxtools
// Unit tests for asset upload helpers and RPM architecture inference.

package assets

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeUploadFormat(t *testing.T) {
	cases := map[string]string{
		"apk":     "alpine",
		"APK":     "alpine",
		"  apk  ": "alpine",
		"ruby":    "rubygems",
		"gem":     "rubygems",
		"gems":    "rubygems",
		"rubygem": "rubygems",
		"python":  "pypi",
		"golang":  "go",
		"git-lfs": "gitlfs",
		"gitlfs":  "gitlfs",
		"conan":   "conan",
		"Apt":     "apt",
		"  yum ":  "yum",
		"some_thing": "something",
		"raw":     "raw",
	}
	for in, want := range cases {
		if got := normalizeUploadFormat(in); got != want {
			t.Errorf("normalizeUploadFormat(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUploadedTarget(t *testing.T) {
	cases := []struct {
		repo, dir, file, want string
	}{
		{"repo", "/", "f.apk", "repo/f.apk"},
		{"repo", "", "f.apk", "repo/f.apk"},
		{"repo", "/dir", "f.apk", "repo/dir/f.apk"},
		{"repo", "dir", "f.apk", "repo/dir/f.apk"},
		{"repo", "/dir/sub/", "f.apk", "repo/dir/sub/f.apk"},
	}
	for _, c := range cases {
		if got := uploadedTarget(c.repo, c.dir, c.file); got != c.want {
			t.Errorf("uploadedTarget(%q,%q,%q) = %q, want %q", c.repo, c.dir, c.file, got, c.want)
		}
	}
}

func TestInferRawDirectory(t *testing.T) {
	if got := inferRawDirectory(""); got != "/" {
		t.Errorf("inferRawDirectory(\"\") = %q, want /", got)
	}
	if got := inferRawDirectory("sub"); got != "/sub" {
		t.Errorf("inferRawDirectory(\"sub\") = %q, want /sub", got)
	}
}

func TestNormalizeRPMArchitectureSegment(t *testing.T) {
	cases := map[string]string{
		"src":    "SRPMS",
		"nosrc":  "SRPMS",
		"X86_64": "x86_64",
		"noarch": "noarch",
		"":       "",
		"  ":     "",
	}
	for in, want := range cases {
		if got := normalizeRPMArchitectureSegment(in); got != want {
			t.Errorf("normalizeRPMArchitectureSegment(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestReadRPMArchitectureFromFilename(t *testing.T) {
	cases := map[string]string{
		"foo-1.0-1.x86_64.rpm": "x86_64",
		"foo-1.0-1.aarch64.rpm": "aarch64",
		"foo-1.0-1.noarch.rpm": "noarch",
		"bar.src.rpm":          "SRPMS",
		"bar.nosrc.rpm":        "SRPMS",
		"baz.weirdarch.rpm":    "", // not an accepted arch
		"notanrpm.txt":         "", // not an rpm at all
	}
	for in, want := range cases {
		if got := readRPMArchitectureFromFilename(in); got != want {
			t.Errorf("readRPMArchitectureFromFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAlignRPMSection(t *testing.T) {
	cases := map[int64]int64{0: 0, 1: 8, 7: 8, 8: 8, 9: 16, 16: 16, 17: 24}
	for in, want := range cases {
		if got := alignRPMSection(in); got != want {
			t.Errorf("alignRPMSection(%d) = %d, want %d", in, got, want)
		}
	}
}

// rpmSection assembles one RPM header/signature section as parsed by
// readRPMHeaderSection: a 16-byte record (magic, reserved, index count, store
// size) followed by the index entries and the string store.
func rpmSection(entries []rpmHeaderIndex, store []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x8e, 0xad, 0xe8, 0x01}) // header magic
	buf.Write([]byte{0, 0, 0, 0})             // reserved
	_ = binary.Write(buf, binary.BigEndian, uint32(len(entries)))
	_ = binary.Write(buf, binary.BigEndian, uint32(len(store)))
	for _, e := range entries {
		_ = binary.Write(buf, binary.BigEndian, e.Tag)
		_ = binary.Write(buf, binary.BigEndian, e.Type)
		_ = binary.Write(buf, binary.BigEndian, e.Offset)
		_ = binary.Write(buf, binary.BigEndian, e.Count)
	}
	buf.Write(store)
	return buf.Bytes()
}

// writeSyntheticRPM produces a minimal but structurally valid RPM whose main
// header carries an arch tag with the given value.
func writeSyntheticRPM(t *testing.T, arch string) string {
	t.Helper()
	lead := make([]byte, rpmLeadSize)
	sig := rpmSection(nil, nil) // empty signature section, 16 bytes, 8-aligned
	store := append([]byte(arch), 0)
	main := rpmSection([]rpmHeaderIndex{
		{Tag: rpmTagArch, Type: rpmTypeString, Offset: 0, Count: 1},
	}, store)

	var rpm []byte
	rpm = append(rpm, lead...)
	rpm = append(rpm, sig...)
	rpm = append(rpm, main...)

	fp := filepath.Join(t.TempDir(), "pkg.rpm")
	if err := os.WriteFile(fp, rpm, 0644); err != nil {
		t.Fatal(err)
	}
	return fp
}

func TestReadRPMArchitectureFromHeader(t *testing.T) {
	fp := writeSyntheticRPM(t, "x86_64")
	got, err := readRPMArchitectureFromHeader(fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "x86_64" {
		t.Fatalf("arch from header = %q, want x86_64", got)
	}
}

func TestInferRPMArchitecture(t *testing.T) {
	// Header wins even when the filename has no arch segment.
	fp := writeSyntheticRPM(t, "aarch64")
	if got := inferRPMArchitecture(fp); got != "aarch64" {
		t.Fatalf("inferRPMArchitecture (header) = %q, want aarch64", got)
	}

	// A non-RPM file falls back to the default packages prefix.
	plain := filepath.Join(t.TempDir(), "plain.bin")
	if err := os.WriteFile(plain, []byte("not an rpm"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := inferRPMArchitecture(plain); got != rpmAutoDirPrefix {
		t.Fatalf("inferRPMArchitecture (fallback) = %q, want %q", got, rpmAutoDirPrefix)
	}
}
