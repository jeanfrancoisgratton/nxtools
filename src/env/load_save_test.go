// nxtools
// Unit tests for environment file load/save and password handling.

package env

import (
	"os"
	"path/filepath"
	"testing"

	hf "github.com/jeanfrancoisgratton/helperFunctions/v5"
)

func TestSafeDecodePassword(t *testing.T) {
	if got := safeDecodePassword(""); got != "" {
		t.Errorf("safeDecodePassword(\"\") = %q, want empty", got)
	}

	// Non-ciphertext input must fall back to the original string rather than panic.
	if got := safeDecodePassword("just-plain-text"); got != "just-plain-text" {
		t.Errorf("safeDecodePassword(plain) = %q, want passthrough", got)
	}

	// A value produced by EncodeString should decode back to the original.
	enc := hf.EncodeString("s3cr3t", "")
	if got := safeDecodePassword(enc); got != "s3cr3t" {
		t.Errorf("safeDecodePassword(encoded) = %q, want s3cr3t", got)
	}
}

func TestSaveLoadEnvironmentRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	if err := os.MkdirAll(filepath.Join(tmp, ".config", "JFG", "nxtools"), 0755); err != nil {
		t.Fatal(err)
	}

	prev := EnvConfigFile
	EnvConfigFile = "unit-test-env"
	t.Cleanup(func() { EnvConfigFile = prev })

	in := EnvironmentStruct{
		NexusServerUrl: "https://nexus.example.com:8443",
		Username:       "admin",
		Password:       hf.EncodeString("s3cr3t", ""), // stored encoded, as AddEnvFile does
		Comments:       "unit test",
	}
	if err := in.SaveEnvironmentFile("unit-test-env.json"); err != nil {
		t.Fatalf("save: %v", err)
	}

	out, err := LoadEnvironmentFile()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out.NexusServerUrl != in.NexusServerUrl {
		t.Errorf("NexusServerUrl = %q", out.NexusServerUrl)
	}
	if out.Username != "admin" {
		t.Errorf("Username = %q", out.Username)
	}
	if out.Comments != "unit test" {
		t.Errorf("Comments = %q", out.Comments)
	}
	// Load transparently decodes the stored password.
	if out.Password != "s3cr3t" {
		t.Errorf("Password = %q, want decoded s3cr3t", out.Password)
	}
}

func TestLoadEnvironmentFile_Missing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	prev := EnvConfigFile
	EnvConfigFile = "does-not-exist"
	t.Cleanup(func() { EnvConfigFile = prev })

	if _, err := LoadEnvironmentFile(); err == nil {
		t.Fatal("expected error loading a missing env file")
	}
}
