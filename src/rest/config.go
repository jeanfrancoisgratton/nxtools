// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/02/24 14:18
// Original filename: src/rest/config.go

package rest

import (
	"os"
	"time"
)

// ConfigFromEnv builds a Config from environment variables.
//
// Required (unless you set Config.Host yourself):
//   - NEXUS_HOST="https://host:port" or "http://host:port"
//
// Optional auth:
//   - NEXUS_USER / NEXUS_PASS (basic auth)
//   - NEXUS_TOKEN (bearer token)
//
// Optional TLS:
//   - NEXUS_CA_CERT, NEXUS_CLIENT_CERT, NEXUS_CLIENT_KEY
//   - NEXUS_TLS_INSECURE=1 (skip cert verification; for testing only)
func ConfigFromEnv() Config {
	host := os.Getenv("NEXUS_HOST")

	cfg := Config{
		Host:               host,
		Username:           os.Getenv("NEXUS_USER"),
		Password:           os.Getenv("NEXUS_PASS"),
		BearerToken:        os.Getenv("NEXUS_TOKEN"),
		CACertPath:         NormalizePath(os.Getenv("NEXUS_CA_CERT")),
		CertPath:           NormalizePath(os.Getenv("NEXUS_CLIENT_CERT")),
		KeyPath:            NormalizePath(os.Getenv("NEXUS_CLIENT_KEY")),
		InsecureSkipVerify: isTruthy(os.Getenv("NEXUS_TLS_INSECURE")),
		FastFailTimeout:    time.Duration(FastFailTimeoutSeconds) * time.Second,
		SessionTimeout:     time.Duration(SessionTimeoutMinutes) * time.Minute,
	}

	// If NEXUS_HOST has no scheme, allow opting into https via NEXUS_TLS.
	if isTruthy(os.Getenv("NEXUS_TLS")) {
		cfg.UseTLS = true
	}

	if ua := os.Getenv("NEXUS_USER_AGENT"); ua != "" {
		cfg.UserAgent = ua
	}
	
	return cfg
}
