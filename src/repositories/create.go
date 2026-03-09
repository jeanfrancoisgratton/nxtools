// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/repositories/create.go

package repositories

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/rest"
)

// CreateRepository creates a repository using the Repositories API.
//
// Endpoint:
//
//	POST /service/rest/v1/repositories/{format}/{type}
//
// You are expected to supply the full JSON payload for that recipe.
// Easiest workflow: export the payload from Nexus' embedded Swagger UI
// (Settings -> System -> API) for the desired recipe.
func CreateRepository(envFile, format, repoType, jsonFile string) *cerr.CustomError {
	format = strings.TrimSpace(format)
	repoType = strings.TrimSpace(repoType)
	jsonFile = strings.TrimSpace(jsonFile)

	if format == "" || repoType == "" {
		return &cerr.CustomError{Title: "Missing parameters", Message: "format and type are required"}
	}
	if jsonFile == "" {
		return &cerr.CustomError{Title: "Missing parameters", Message: "json payload file is required"}
	}

	var payload []byte
	var e1 error
	if jsonFile == "-" {
		payload, e1 = io.ReadAll(os.Stdin)
	} else {
		payload, e1 = os.ReadFile(filepath.Clean(jsonFile))
	}
	if e1 != nil {
		return &cerr.CustomError{Title: "Unable to read JSON payload", Message: e1.Error()}
	}
	if len(bytes.TrimSpace(payload)) == 0 {
		return &cerr.CustomError{Title: "Invalid JSON payload", Message: "payload is empty"}
	}

	c, err := rest.NewClientFromEnvFile(envFile)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/service/rest/v1/repositories/%s/%s", format, repoType)
	headers := http.Header{}
	headers.Set("Accept", "application/json")
	headers.Set("Content-Type", "application/json")

	resp, e2 := c.Do(context.Background(), http.MethodPost, path, nil, bytes.NewReader(payload), headers)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to create repository", Message: "HTTP status code: " + resp.Status}
	}

	// Nexus often returns 201 + an empty body for create.
	return nil
}
