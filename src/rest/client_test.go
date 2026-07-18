// nxtools
// Unit tests for NewClient construction and URL building.

package rest

import (
	"strings"
	"testing"
)

func TestNewClient_EmptyHost(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestNewClient_SchemeDefaulting(t *testing.T) {
	c, err := NewClient(Config{Host: "localhost:8081"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.BaseURL(); got != "http://localhost:8081" {
		t.Fatalf("default scheme should be http, got %q", got)
	}

	c, err = NewClient(Config{Host: "localhost:8081", UseTLS: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.BaseURL(); got != "https://localhost:8081" {
		t.Fatalf("UseTLS should force https, got %q", got)
	}
}

func TestNewClient_ExplicitSchemePreserved(t *testing.T) {
	c, err := NewClient(Config{Host: "https://nexus.example.com:8443"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.BaseURL(); got != "https://nexus.example.com:8443" {
		t.Fatalf("BaseURL = %q", got)
	}
}

func TestNewClient_UnsupportedScheme(t *testing.T) {
	if _, err := NewClient(Config{Host: "ftp://nexus.example.com"}); err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}

func TestNewClient_DefaultUserAgent(t *testing.T) {
	c, err := NewClient(Config{Host: "http://localhost:8081"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.userAgent != "nxtools" {
		t.Fatalf("default user agent = %q, want nxtools", c.userAgent)
	}
}

func TestClient_DumpURL(t *testing.T) {
	// Plain host, no base path.
	c, err := NewClient(Config{Host: "http://localhost:8081"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.DumpURL("/service/rest/v1/repositories"); got != "http://localhost:8081/service/rest/v1/repositories" {
		t.Fatalf("DumpURL = %q", got)
	}

	// Host carrying a reverse-proxy base path.
	c, err = NewClient(Config{Host: "https://host.example.com/nexus"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.DumpURL("/service/rest/v1/status"); got != "https://host.example.com/nexus/service/rest/v1/status" {
		t.Fatalf("DumpURL with base path = %q", got)
	}
}

func TestGetWithOptions_EmptyTarget(t *testing.T) {
	c, err := NewClient(Config{Host: "http://localhost:8081"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, cerr := c.GetWithOptions(nil, "   ", nil); cerr == nil {
		t.Fatal("expected error for empty GET target")
	} else if !strings.Contains(cerr.Error(), "empty GET target") {
		t.Fatalf("unexpected error message: %v", cerr)
	}
}
