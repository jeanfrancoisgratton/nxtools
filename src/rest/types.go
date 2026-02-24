// nxtools
// REST client types for Nexus Repository Manager 3 (NxRM).

package rest

import (
	"net/http"
	"net/url"
	"time"
)

// Global timeout defaults.
//
// fast-fail: protects connect / TLS handshake / header waits (seconds)
// session-timeout: protects long-running operations that should complete (minutes)
var FastFailTimeoutSeconds = 30
var SessionTimeoutMinutes = 30

// Client wraps an http.Client and knows how to talk to a Nexus server over HTTP(S).
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL

	fastFailTimeout time.Duration
	sessionTimeout  time.Duration

	username    string
	password    string
	bearerToken string
	userAgent   string
}

// Config holds the connection parameters for the REST client.
type Config struct {
	// Host is the Nexus base URL, e.g. "https://nexus.example.com:8443".
	//
	// If you pass "host:port" without a scheme, http:// is assumed unless UseTLS=true.
	Host   string
	UseTLS bool

	// Auth is optional. If both are set, BearerToken takes precedence.
	Username    string
	Password    string
	BearerToken string

	// TLS is only used when scheme is https.
	CACertPath         string
	CertPath           string
	KeyPath            string
	InsecureSkipVerify bool

	UserAgent string // optional, defaults to "nxtools"

	// FastFailTimeout is used to quickly fail when the server is unreachable or
	// stalls before responding (dial/TLS handshake/headers).
	// If <= 0, a default is applied.
	FastFailTimeout time.Duration

	// SessionTimeout is applied as a *per-request* timeout for operations that
	// are expected to complete.
	// If 0, session timeouts are disabled (caller can still cancel via context).
	SessionTimeout time.Duration
}
