// nxtools
// Integration-style tests for the REST client against an in-process HTTP server.

package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClientDo_AgainstTestServer(t *testing.T) {
	var gotPath, gotQuery, gotAuth, gotUA, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{Host: srv.URL, Username: "admin", Password: "pw"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	q := url.Values{}
	q.Set("repository", "myrepo")

	resp, cerr := c.Do(context.Background(), http.MethodPost, "/service/rest/v1/components", q, strings.NewReader("payload-bytes"), nil)
	if cerr != nil {
		t.Fatalf("Do: %v", cerr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
	if gotPath != "/service/rest/v1/components" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "repository=myrepo" {
		t.Errorf("query = %q", gotQuery)
	}
	if !strings.HasPrefix(gotAuth, "Basic ") {
		t.Errorf("auth = %q, want Basic", gotAuth)
	}
	if gotUA != "nxtools" {
		t.Errorf("user-agent = %q, want nxtools", gotUA)
	}
	if gotBody != "payload-bytes" {
		t.Errorf("body = %q", gotBody)
	}
}

func TestClientDo_BearerAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := NewClient(Config{Host: srv.URL, BearerToken: "tok123"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resp, cerr := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/status", nil, nil, nil)
	if cerr != nil {
		t.Fatalf("Do: %v", cerr)
	}
	defer resp.Body.Close()

	if gotAuth != "Bearer tok123" {
		t.Errorf("auth = %q, want 'Bearer tok123'", gotAuth)
	}
}
