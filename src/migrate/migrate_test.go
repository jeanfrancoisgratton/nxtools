// nxtools
// Integration-style tests for MigrateRepo, driven against an in-process fake
// Nexus server (httptest) that models just enough of the REST API for the
// migrate code path to run end-to-end. These exist to lock in the six
// data-loss/correctness bugs fixed in migrate.go on 2026-09-15 so they can't
// silently regress.

package migrate

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"nxtools/env"
	"nxtools/repositories"
	"nxtools/shared"
)

// ---- fake Nexus server -----------------------------------------------------

type fakeAsset struct {
	path    string
	content []byte
}

type fakeRepo struct {
	Name      string
	Format    string
	Type      string
	BlobStore string
	Members   []string
	Assets    []fakeAsset
}

type fakeNexus struct {
	mu      sync.Mutex
	repos   map[string]*fakeRepo
	baseURL string
}

func newFakeNexus(t *testing.T) *fakeNexus {
	t.Helper()
	fn := &fakeNexus{repos: map[string]*fakeRepo{}}
	srv := httptest.NewServer(http.HandlerFunc(fn.handle))
	t.Cleanup(srv.Close)
	fn.baseURL = srv.URL
	return fn
}

func (fn *fakeNexus) addRepo(r *fakeRepo) {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	fn.repos[r.Name] = r
}

func (fn *fakeNexus) getRepo(name string) *fakeRepo {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	return fn.repos[name]
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (fn *fakeNexus) handle(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/service/rest/v1/repositories" && r.Method == http.MethodGet:
		fn.listRepositories(w)
	case r.URL.Path == "/service/rest/v1/assets" && r.Method == http.MethodGet:
		fn.listAssets(w, r)
	case r.URL.Path == "/service/rest/v1/components" && r.Method == http.MethodPost:
		fn.uploadComponent(w, r)
	case r.URL.Path == "/download" && r.Method == http.MethodGet:
		fn.download(w, r)
	case strings.HasPrefix(r.URL.Path, "/service/rest/v1/repositories/"):
		fn.repositoriesSubpath(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (fn *fakeNexus) listRepositories(w http.ResponseWriter) {
	fn.mu.Lock()
	out := make([]repositories.RepositorySummary, 0, len(fn.repos))
	for _, rp := range fn.repos {
		out = append(out, repositories.RepositorySummary{Name: rp.Name, Format: rp.Format, Type: rp.Type})
	}
	fn.mu.Unlock()
	writeJSON(w, http.StatusOK, out)
}

func (fn *fakeNexus) repositoriesSubpath(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/service/rest/v1/repositories/")
	segs := strings.Split(rest, "/")

	switch len(segs) {
	case 1:
		name, _ := url.PathUnescape(segs[0])
		switch r.Method {
		case http.MethodGet:
			fn.getRepositorySummary(w, name)
		case http.MethodDelete:
			fn.deleteRepository(w, name)
		default:
			http.NotFound(w, r)
		}
	case 2:
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		fn.createRepository(w, r, segs[0], segs[1])
	case 3:
		format, kind, name := segs[0], segs[1], segs[2]
		name, _ = url.PathUnescape(name)
		switch {
		case kind == "hosted" && r.Method == http.MethodGet:
			fn.getHostedDetail(w, format, name)
		case kind == "group" && r.Method == http.MethodGet:
			fn.getGroupDetail(w, name)
		case kind == "group" && r.Method == http.MethodPut:
			fn.updateGroupDetail(w, r, name)
		default:
			http.NotFound(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

func (fn *fakeNexus) getRepositorySummary(w http.ResponseWriter, name string) {
	rp := fn.getRepo(name)
	if rp == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, repositories.RepositorySummary{
		Name: rp.Name, Format: rp.Format, Type: rp.Type,
		Storage: repositories.StorageAttributesStruct{BlobStoreName: rp.BlobStore},
	})
}

func (fn *fakeNexus) deleteRepository(w http.ResponseWriter, name string) {
	fn.mu.Lock()
	delete(fn.repos, name)
	fn.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (fn *fakeNexus) createRepository(w http.ResponseWriter, r *http.Request, format, kind string) {
	var body repositories.HostedRepoCommonAttributesStruct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fn.addRepo(&fakeRepo{Name: body.Name, Format: format, Type: kind, BlobStore: body.Storage.BlobStoreName})
	w.WriteHeader(http.StatusCreated)
}

func (fn *fakeNexus) getHostedDetail(w http.ResponseWriter, _ string, name string) {
	rp := fn.getRepo(name)
	if rp == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, repositories.HostedRepoCommonAttributesStruct{
		Name: rp.Name, Format: rp.Format, Type: rp.Type,
		Storage: repositories.StorageAttributesStruct{BlobStoreName: rp.BlobStore},
	})
}

func (fn *fakeNexus) getGroupDetail(w http.ResponseWriter, name string) {
	rp := fn.getRepo(name)
	if rp == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	fn.mu.Lock()
	members := append([]string(nil), rp.Members...)
	fn.mu.Unlock()
	writeJSON(w, http.StatusOK, repositories.GroupedRepoCommonAttributesStruct{
		Name: rp.Name, Online: true,
		Storage: repositories.StorageAttributesStruct{BlobStoreName: rp.BlobStore},
		Group:   repositories.GroupAttributesStruct{MemberNames: members},
	})
}

func (fn *fakeNexus) updateGroupDetail(w http.ResponseWriter, r *http.Request, name string) {
	var body repositories.GroupedRepoCommonAttributesStruct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fn.mu.Lock()
	rp := fn.repos[name]
	if rp == nil {
		fn.mu.Unlock()
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rp.Members = body.Group.MemberNames
	fn.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (fn *fakeNexus) listAssets(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repository")
	rp := fn.getRepo(repoName)

	var items []shared.AssetSummary
	if rp != nil {
		fn.mu.Lock()
		for _, a := range rp.Assets {
			q := url.Values{}
			q.Set("repo", repoName)
			q.Set("path", a.path)
			items = append(items, shared.AssetSummary{
				Path:        a.path,
				DownloadURL: fn.baseURL + "/download?" + q.Encode(),
				Format:      rp.Format,
				Repository:  repoName,
				FileSize:    int64(len(a.content)),
			})
		}
		fn.mu.Unlock()
	}
	writeJSON(w, http.StatusOK, shared.ListAssetResponse{Items: items, ContinuationToken: nil})
}

func (fn *fakeNexus) download(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	assetPath := r.URL.Query().Get("path")

	rp := fn.getRepo(repoName)
	if rp == nil {
		http.NotFound(w, r)
		return
	}

	fn.mu.Lock()
	defer fn.mu.Unlock()
	for _, a := range rp.Assets {
		if a.path == assetPath {
			_, _ = w.Write(a.content)
			return
		}
	}
	http.NotFound(w, r)
}

// uploadComponent handles the raw-format multipart upload
// (POST /service/rest/v1/components?repository=X, fields raw.directory /
// raw.asset1.filename / raw.asset1) issued by assets.uploadRaw.
func (fn *fakeNexus) uploadComponent(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repository")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	directory := strings.Trim(r.FormValue("raw.directory"), "/")
	filename := r.FormValue("raw.asset1.filename")

	file, _, err := r.FormFile("raw.asset1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	assetPath := filename
	if directory != "" {
		assetPath = directory + "/" + filename
	}

	fn.mu.Lock()
	rp := fn.repos[repoName]
	if rp == nil {
		fn.mu.Unlock()
		http.Error(w, "repo not found", http.StatusNotFound)
		return
	}
	rp.Assets = append(rp.Assets, fakeAsset{path: assetPath, content: content})
	fn.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

// ---- test fixtures ----------------------------------------------------------

// resetRepoGlobals snapshots the package-level mutable state that
// MigrateRepo/CreateRepository/etc rely on and restores it after the test, so
// tests don't leak configuration into one another.
func resetRepoGlobals(t *testing.T) {
	t.Helper()
	origFormat := repositories.RepoFormat
	origType := repositories.RepoType
	origKeep := repositories.KeepSource
	origWritePolicy := repositories.StorageWritePolicy
	origQuiet := shared.QuietOutput
	origEnvfile := shared.Envfile

	t.Cleanup(func() {
		repositories.RepoFormat = origFormat
		repositories.RepoType = origType
		repositories.KeepSource = origKeep
		repositories.StorageWritePolicy = origWritePolicy
		shared.QuietOutput = origQuiet
		shared.Envfile = origEnvfile
	})
}

// configureTestEnv points nxtools' env-file-backed REST client at the fake
// server by writing a throwaway env file under a temp $HOME.
func configureTestEnv(t *testing.T, serverURL string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "JFG", "nxtools")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	cfg := env.EnvironmentStruct{NexusServerUrl: serverURL, Username: "admin", Password: "admin123"}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal env config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "defaultEnv.json"), b, 0600); err != nil {
		t.Fatalf("write env config: %v", err)
	}

	shared.Envfile = "defaultEnv.json"
}

func newMigrateTestFixture(t *testing.T) *fakeNexus {
	t.Helper()
	resetRepoGlobals(t)
	fn := newFakeNexus(t)
	configureTestEnv(t, fn.baseURL)
	return fn
}

// ---- tests -------------------------------------------------------------

func TestMigrateRepo_SourceNotFound(t *testing.T) {
	newMigrateTestFixture(t)

	ce := MigrateRepo("does-not-exist", "target")
	if ce == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(ce.Title, "does not exist") {
		t.Errorf("Title = %q, want it to mention the source repo not existing", ce.Title)
	}
}

func TestMigrateRepo_NonHostedSourceRejected(t *testing.T) {
	fn := newMigrateTestFixture(t)
	fn.addRepo(&fakeRepo{Name: "src-proxy", Format: "raw", Type: "proxy"})

	ce := MigrateRepo("src-proxy", "target")
	if ce == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(ce.Message, "hosted") {
		t.Errorf("Message = %q, want it to mention hosted-only restriction", ce.Message)
	}
}

// TestMigrateRepo_NonHostedExistingTargetRejected locks in bug #6: an
// existing target repo of a non-hosted type (proxy/group) must be rejected
// up front instead of being silently written into.
func TestMigrateRepo_NonHostedExistingTargetRejected(t *testing.T) {
	fn := newMigrateTestFixture(t)
	fn.addRepo(&fakeRepo{Name: "src-raw", Format: "raw", Type: "hosted", BlobStore: "blob1",
		Assets: []fakeAsset{{path: "file1.txt", content: []byte("hello")}}})
	fn.addRepo(&fakeRepo{Name: "dst-group", Format: "raw", Type: "group"})

	ce := MigrateRepo("src-raw", "dst-group")
	if ce == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(ce.Message, "hosted") {
		t.Errorf("Message = %q, want it to mention that only hosted targets are allowed", ce.Message)
	}

	// Nothing should have been touched.
	if rp := fn.getRepo("src-raw"); rp == nil || len(rp.Assets) != 1 {
		t.Errorf("source repo should be untouched, got %+v", rp)
	}
}

func TestMigrateRepo_NoAssetsInSource(t *testing.T) {
	fn := newMigrateTestFixture(t)
	fn.addRepo(&fakeRepo{Name: "src-empty", Format: "raw", Type: "hosted", BlobStore: "blob1"})

	ce := MigrateRepo("src-empty", "dst-empty")
	if ce == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(ce.Title, "No assets found") {
		t.Errorf("Title = %q, want it to mention no assets found", ce.Title)
	}
}

// TestMigrateRepo_CreatesTargetAndPreservesDirectories locks in bugs #1
// (multi-asset migrations losing data), #2 (auto-created target repo failing
// because RepoFormat wasn't set), and #5 (raw migrations flattening
// directory structure).
func TestMigrateRepo_CreatesTargetAndPreservesDirectories(t *testing.T) {
	fn := newMigrateTestFixture(t)
	fn.addRepo(&fakeRepo{
		Name: "src-raw", Format: "raw", Type: "hosted", BlobStore: "blob1",
		Assets: []fakeAsset{
			{path: "dirA/file1.txt", content: []byte("content-one")},
			{path: "dirB/sub/file2.txt", content: []byte("content-two")},
		},
	})

	ce := MigrateRepo("src-raw", "dst-raw")
	if ce != nil {
		t.Fatalf("MigrateRepo returned an error: %s: %s", ce.Title, ce.Message)
	}

	dst := fn.getRepo("dst-raw")
	if dst == nil {
		t.Fatal("target repo was not created")
	}
	if dst.Format != "raw" {
		t.Errorf("target repo format = %q, want %q (bug #2: RepoFormat must be set before CreateRepository)", dst.Format, "raw")
	}
	if dst.BlobStore != "blob1" {
		t.Errorf("target repo blob store = %q, want %q (config should be copied from source)", dst.BlobStore, "blob1")
	}
	if len(dst.Assets) != 2 {
		t.Fatalf("target repo has %d assets, want 2 (bug #1: no asset should be lost)", len(dst.Assets))
	}

	got := map[string]string{}
	for _, a := range dst.Assets {
		got[a.path] = string(a.content)
	}
	want := map[string]string{
		"dirA/file1.txt":     "content-one",
		"dirB/sub/file2.txt": "content-two",
	}
	for path, content := range want {
		if got[path] != content {
			t.Errorf("asset %q = %q, want %q (bug #5: raw migrations must preserve directory structure)", path, got[path], content)
		}
	}

	if fn.getRepo("src-raw") != nil {
		t.Error("source repo should have been deleted after a successful migration (KeepSource defaults to false)")
	}
}

func TestMigrateRepo_KeepSourceOption(t *testing.T) {
	fn := newMigrateTestFixture(t)
	fn.addRepo(&fakeRepo{
		Name: "src-raw", Format: "raw", Type: "hosted", BlobStore: "blob1",
		Assets: []fakeAsset{{path: "file1.txt", content: []byte("hello")}},
	})

	repositories.KeepSource = true

	ce := MigrateRepo("src-raw", "dst-raw")
	if ce != nil {
		t.Fatalf("MigrateRepo returned an error: %s: %s", ce.Title, ce.Message)
	}

	if fn.getRepo("src-raw") == nil {
		t.Error("source repo should still exist when KeepSource is set")
	}
}

// TestMigrateRepo_GroupMembershipRepointed locks in bug #3: group repos that
// referenced the source repo must be repointed at the new repo, not left as
// a no-op.
func TestMigrateRepo_GroupMembershipRepointed(t *testing.T) {
	fn := newMigrateTestFixture(t)
	fn.addRepo(&fakeRepo{
		Name: "src-raw", Format: "raw", Type: "hosted", BlobStore: "blob1",
		Assets: []fakeAsset{{path: "file1.txt", content: []byte("hello")}},
	})
	fn.addRepo(&fakeRepo{
		Name: "grp-raw", Format: "raw", Type: "group",
		Members: []string{"src-raw", "other-raw"},
	})

	ce := MigrateRepo("src-raw", "dst-raw")
	if ce != nil {
		t.Fatalf("MigrateRepo returned an error: %s: %s", ce.Title, ce.Message)
	}

	grp := fn.getRepo("grp-raw")
	if grp == nil {
		t.Fatal("group repo vanished")
	}

	memberSet := map[string]bool{}
	for _, m := range grp.Members {
		memberSet[m] = true
	}
	if memberSet["src-raw"] {
		t.Error("group still lists the source repo as a member after migration")
	}
	if !memberSet["dst-raw"] {
		t.Error("group does not list the new repo as a member after migration")
	}
	if !memberSet["other-raw"] {
		t.Error("unrelated group member was dropped")
	}
}
