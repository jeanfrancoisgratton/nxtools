// nxtools
// Shared in-process fake Nexus server used by the tasks package's
// integration-style tests, modeled after the equivalent fixture in
// migrate/migrate_test.go.

package tasks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"nxtools/blobstores"
	"nxtools/env"
	"nxtools/shared"
)

type fakeTask struct {
	ID           string
	Name         string
	Type         string
	CurrentState string
	Properties   map[string]string
}

// fakeNexus models just enough of the Tasks and Blobstores REST APIs for the
// tasks package's functions to run end-to-end against.
type fakeNexus struct {
	mu         sync.Mutex
	baseURL    string
	tasks      []fakeTask
	blobstores []string
	nextID     int

	// failStatus, keyed by "METHOD path", forces that exact request to
	// return the given HTTP status instead of the normal handling. Used to
	// exercise failure/aggregation paths (e.g. one bad ID among several).
	failStatus map[string]int

	runCalls    []string
	stopCalls   []string
	deleteCalls []string
	createCalls []TaskTemplateXO
}

func newFakeNexus(t *testing.T) *fakeNexus {
	t.Helper()
	fn := &fakeNexus{failStatus: map[string]int{}}
	srv := httptest.NewServer(http.HandlerFunc(fn.handle))
	t.Cleanup(srv.Close)
	fn.baseURL = srv.URL
	return fn
}

func (fn *fakeNexus) addTask(id, name, taskType, state string, props map[string]string) {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	fn.tasks = append(fn.tasks, fakeTask{ID: id, Name: name, Type: taskType, CurrentState: state, Properties: props})
}

func (fn *fakeNexus) addBlobstore(name string) {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	fn.blobstores = append(fn.blobstores, name)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (fn *fakeNexus) toSummary(t fakeTask) TaskSummary {
	return TaskSummary{ID: t.ID, Name: t.Name, Type: t.Type, CurrentState: t.CurrentState, Properties: t.Properties}
}

func (fn *fakeNexus) handle(w http.ResponseWriter, r *http.Request) {
	fn.mu.Lock()
	defer fn.mu.Unlock()

	if status, ok := fn.failStatus[r.Method+" "+r.URL.Path]; ok {
		w.WriteHeader(status)
		return
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/service/rest/v1/tasks":
		typeFilter := r.URL.Query().Get("type")
		var items []TaskSummary
		for _, t := range fn.tasks {
			if typeFilter != "" && t.Type != typeFilter {
				continue
			}
			items = append(items, fn.toSummary(t))
		}
		writeJSON(w, http.StatusOK, ListTasksResponse{Items: items})

	case r.Method == http.MethodGet && r.URL.Path == "/service/rest/v1/blobstores":
		var blobs []blobstores.BlobStoreSummary
		for _, b := range fn.blobstores {
			blobs = append(blobs, blobstores.BlobStoreSummary{Name: b, Type: "File"})
		}
		writeJSON(w, http.StatusOK, blobs)

	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/run"):
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/service/rest/v1/tasks/"), "/run")
		fn.runCalls = append(fn.runCalls, id)
		w.WriteHeader(http.StatusNoContent)

	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/stop"):
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/service/rest/v1/tasks/"), "/stop")
		fn.stopCalls = append(fn.stopCalls, id)
		w.WriteHeader(http.StatusNoContent)

	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/service/rest/v1/tasks/"):
		id := strings.TrimPrefix(r.URL.Path, "/service/rest/v1/tasks/")
		fn.deleteCalls = append(fn.deleteCalls, id)
		var kept []fakeTask
		for _, t := range fn.tasks {
			if t.ID != id {
				kept = append(kept, t)
			}
		}
		fn.tasks = kept
		// The live server returns 204 here despite the swagger doc claiming
		// 200 (see delete.go) — model the real, observed behavior.
		w.WriteHeader(http.StatusNoContent)

	case r.Method == http.MethodPost && r.URL.Path == "/service/rest/v1/tasks":
		var payload TaskTemplateXO
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fn.createCalls = append(fn.createCalls, payload)
		fn.nextID++
		nt := fakeTask{
			ID:           "task-" + strconv.Itoa(fn.nextID),
			Name:         payload.Name,
			Type:         payload.Type,
			CurrentState: "WAITING",
			Properties:   payload.Properties,
		}
		fn.tasks = append(fn.tasks, nt)
		writeJSON(w, http.StatusCreated, fn.toSummary(nt))

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// configureTestEnv points nxtools' env-file-backed REST client at the fake
// server by writing a throwaway env file under a temp $HOME, mirroring
// migrate/migrate_test.go's fixture of the same name.
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

	origEnvfile := shared.Envfile
	origQuiet := shared.QuietOutput
	shared.Envfile = "defaultEnv.json"
	shared.QuietOutput = true
	t.Cleanup(func() {
		shared.Envfile = origEnvfile
		shared.QuietOutput = origQuiet
	})
}

// newTaskTestFixture spins up a fake Nexus server and points the tasks
// package's REST client at it.
func newTaskTestFixture(t *testing.T) *fakeNexus {
	t.Helper()
	fn := newFakeNexus(t)
	configureTestEnv(t, fn.baseURL)
	return fn
}
