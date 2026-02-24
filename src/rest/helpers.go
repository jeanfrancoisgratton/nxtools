// nxtools
// Small helper functions for the REST client.

package rest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// buildTLSConfig constructs a *tls.Config from the given settings.
// Callers should only use it when the target scheme is https.
func buildTLSConfig(cfg Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	// Root CAs
	if cfg.CACertPath != "" {
		caPEM, err := os.ReadFile(cfg.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA cert %q: %w", cfg.CACertPath, err)
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to parse CA cert %q", cfg.CACertPath)
		}
		tlsConfig.RootCAs = pool
	} else {
		// Use system roots if available.
		sysPool, _ := x509.SystemCertPool()
		tlsConfig.RootCAs = sysPool
	}

	// Client certificate (mTLS)
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load client cert/key (%q, %q): %w", cfg.CertPath, cfg.KeyPath, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func joinURLPath(basePath, addPath string) string {
	if basePath == "" || basePath == "/" {
		return addPath
	}
	return strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(addPath, "/")
}

// DumpURL is a small helper for debugging.
func (c *Client) DumpURL(path string) string {
	u := *c.baseURL
	u.Path = joinURLPath(c.baseURL.Path, path)
	return u.String()
}

// NormalizePath expands ~ to $HOME and cleans the path.
func NormalizePath(p string) string {
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

func isTruthy(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "1" || s == "true" || s == "yes" || s == "y" || s == "on"
}
