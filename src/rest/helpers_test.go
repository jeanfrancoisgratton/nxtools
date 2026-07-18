// nxtools
// Unit tests for the pure helpers in the rest package.

package rest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJoinURLPath(t *testing.T) {
	cases := []struct {
		name     string
		base     string
		add      string
		expected string
	}{
		{"empty base", "", "/service/rest", "/service/rest"},
		{"root base", "/", "/service/rest", "/service/rest"},
		{"base with trailing slash", "/nexus/", "/service", "/nexus/service"},
		{"base without trailing slash", "/nexus", "service", "/nexus/service"},
		{"both slashes trimmed", "/nexus/", "/service/", "/nexus/service/"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := joinURLPath(c.base, c.add); got != c.expected {
				t.Fatalf("joinURLPath(%q, %q) = %q, want %q", c.base, c.add, got, c.expected)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	if got := NormalizePath(""); got != "" {
		t.Fatalf("NormalizePath(\"\") = %q, want empty", got)
	}
	if got := NormalizePath("/etc/ssl/ca.pem"); got != "/etc/ssl/ca.pem" {
		t.Fatalf("NormalizePath absolute path changed: %q", got)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home dir available: %v", err)
	}
	want := filepath.Join(home, "certs/ca.pem")
	if got := NormalizePath("~/certs/ca.pem"); got != want {
		t.Fatalf("NormalizePath(~) = %q, want %q", got, want)
	}
}

func TestIsTruthy(t *testing.T) {
	truthy := []string{"1", "true", "TRUE", "yes", "Y", "on", "  On  "}
	for _, s := range truthy {
		if !isTruthy(s) {
			t.Errorf("isTruthy(%q) = false, want true", s)
		}
	}
	falsy := []string{"", "0", "false", "no", "n", "off", "nope", "2"}
	for _, s := range falsy {
		if isTruthy(s) {
			t.Errorf("isTruthy(%q) = true, want false", s)
		}
	}
}

func TestBuildTLSConfig_MissingCACert(t *testing.T) {
	_, err := buildTLSConfig(Config{CACertPath: filepath.Join(t.TempDir(), "does-not-exist.pem")})
	if err == nil {
		t.Fatal("expected error for missing CA cert, got nil")
	}
	if !strings.Contains(err.Error(), "unable to read CA cert") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildTLSConfig_Defaults(t *testing.T) {
	tlsCfg, err := buildTLSConfig(Config{InsecureSkipVerify: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tlsCfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify not propagated")
	}
	if tlsCfg.MinVersion == 0 {
		t.Error("MinVersion should be set")
	}
}
