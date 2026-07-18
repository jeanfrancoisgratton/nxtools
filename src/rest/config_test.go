// nxtools
// Unit tests for ConfigFromEnv.

package rest

import "testing"

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("NEXUS_HOST", "https://nexus.example.com:8443")
	t.Setenv("NEXUS_USER", "admin")
	t.Setenv("NEXUS_PASS", "s3cr3t")
	t.Setenv("NEXUS_TOKEN", "tok")
	t.Setenv("NEXUS_TLS_INSECURE", "1")
	t.Setenv("NEXUS_TLS", "yes")
	t.Setenv("NEXUS_USER_AGENT", "nxtools-test")

	cfg := ConfigFromEnv()

	if cfg.Host != "https://nexus.example.com:8443" {
		t.Errorf("Host = %q", cfg.Host)
	}
	if cfg.Username != "admin" || cfg.Password != "s3cr3t" {
		t.Errorf("credentials = %q / %q", cfg.Username, cfg.Password)
	}
	if cfg.BearerToken != "tok" {
		t.Errorf("BearerToken = %q", cfg.BearerToken)
	}
	if !cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true when NEXUS_TLS_INSECURE=1")
	}
	if !cfg.UseTLS {
		t.Error("UseTLS should be true when NEXUS_TLS is truthy")
	}
	if cfg.UserAgent != "nxtools-test" {
		t.Errorf("UserAgent = %q", cfg.UserAgent)
	}
}

func TestConfigFromEnv_Defaults(t *testing.T) {
	// Ensure optional toggles default to off/empty when unset.
	t.Setenv("NEXUS_HOST", "http://localhost:8081")
	t.Setenv("NEXUS_TLS_INSECURE", "")
	t.Setenv("NEXUS_TLS", "")
	t.Setenv("NEXUS_USER_AGENT", "")

	cfg := ConfigFromEnv()
	if cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should default to false")
	}
	if cfg.UseTLS {
		t.Error("UseTLS should default to false")
	}
	if cfg.UserAgent != "" {
		t.Errorf("UserAgent should be empty by default, got %q", cfg.UserAgent)
	}
}
