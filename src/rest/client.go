// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/02/24 14:18
// Original filename: src/rest/client.go

package rest

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// NewClient builds a Client from Config.
// If cfg.Host is empty, it falls back to NEXUS_HOST.
func NewClient(cfg Config) (*Client, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = strings.TrimSpace(os.Getenv("NEXUS_HOST"))
	}
	if host == "" {
		return nil, errors.New("nexus host is empty (set cfg.Host or NEXUS_HOST)")
	}

	// If no scheme is provided, assume http:// unless the caller opted into TLS.
	if !strings.Contains(host, "://") {
		if cfg.UseTLS {
			host = "https://" + host
		} else {
			host = "http://" + host
		}
	}

	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, errors.New("unsupported scheme in NEXUS_HOST (use http or https)")
	}
	if baseURL.Host == "" {
		return nil, errors.New("invalid nexus host (missing host:port)")
	}

	fastFail := cfg.FastFailTimeout
	if fastFail <= 0 {
		fastFail = time.Duration(FastFailTimeoutSeconds) * time.Second
		if fastFail <= 0 {
			fastFail = 30 * time.Second
		}
	}

	// Session timeout may be zero (explicitly disabled).
	sessionTimeout := cfg.SessionTimeout
	if sessionTimeout < 0 {
		sessionTimeout = 0
	}

	// TLS
	//
	// We build TLS config if:
	//   - the base URL is https, OR
	//   - the caller provided TLS-related settings (to support http->https redirects).
	var tlsCfg *tls.Config
	if strings.EqualFold(baseURL.Scheme, "https") || cfg.CACertPath != "" || cfg.CertPath != "" || cfg.KeyPath != "" || cfg.InsecureSkipVerify {
		cfg.CACertPath = NormalizePath(cfg.CACertPath)
		cfg.CertPath = NormalizePath(cfg.CertPath)
		cfg.KeyPath = NormalizePath(cfg.KeyPath)

		tlsCfg, err = buildTLSConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

	dialer := &net.Dialer{Timeout: fastFail, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		TLSClientConfig:       tlsCfg,
		TLSHandshakeTimeout:   fastFail,
		ResponseHeaderTimeout: fastFail,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	}

	// IMPORTANT: http.Client.Timeout is deliberately disabled.
	// We enforce timeouts per request with contexts so large uploads/downloads
	// are not killed mid-transfer.
	httpClient := &http.Client{Transport: transport, Timeout: 0}

	ua := strings.TrimSpace(cfg.UserAgent)
	if ua == "" {
		ua = "nxtools"
	}

	return &Client{
		httpClient:      httpClient,
		baseURL:         baseURL,
		fastFailTimeout: fastFail,
		sessionTimeout:  sessionTimeout,
		username:        cfg.Username,
		password:        cfg.Password,
		bearerToken:     cfg.BearerToken,
		userAgent:       ua,
	}, nil
}

func (c cancelOnClose) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return c.ReadCloser.Close()
}

// Do issues an HTTP request to the Nexus server.
//
// `path` is relative to cfg.Host and may include a base path if you run Nexus
// behind a reverse proxy (e.g. cfg.Host="https://host/nexus").
// Example path: "/service/rest/v1/repositories".
func (c *Client) Do(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	body io.Reader,
	headers http.Header,
) (*http.Response, error) {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	u := *c.baseURL
	u.Path = joinURLPath(c.baseURL.Path, path)
	u.RawQuery = ""
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	reqCtx := ctx
	if reqCtx == nil {
		reqCtx = context.Background()
	}

	var cancel context.CancelFunc
	if _, hasDeadline := reqCtx.Deadline(); !hasDeadline {
		if c.sessionTimeout > 0 {
			reqCtx, cancel = context.WithTimeout(reqCtx, c.sessionTimeout)
		}
	}

	req, err := http.NewRequestWithContext(reqCtx, method, u.String(), body)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	// Default headers
	if c.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	// Caller headers
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	// Optional auth (caller can override by setting Authorization explicitly).
	if req.Header.Get("Authorization") == "" {
		if strings.TrimSpace(c.bearerToken) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.bearerToken))
		} else if c.username != "" {
			req.SetBasicAuth(c.username, c.password)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	if cancel != nil {
		resp.Body = cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
	}
	return resp, nil
}

// BaseURL returns the resolved base URL the client uses.
func (c *Client) BaseURL() string {
	if c.baseURL == nil {
		return ""
	}
	return c.baseURL.String()
}
