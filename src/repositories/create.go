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
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

// CreateRepository creates a repository using the Repositories API.
//
// Endpoint:
//
//	POST /service/rest/v1/repositories/{format}/{type}

func CreateRepository_old(format, repoType, jsonFile string) *cerr.CustomError {
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

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/service/rest/v1/repositories/%s/%s", format, repoType)
	headers := http.Header{}
	headers.Set("Accept", "application/json")
	headers.Set("Content-Type", "application/json")

	resp, e2 := c.Do(context.Background(), http.MethodPost, path, nil, bytes.NewReader(payload), headers)
	if e2 != nil {
		return e2
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to create repository", Message: "HTTP status code: " + resp.Status}
	}

	// Nexus often returns 201 + an empty body for create.
	return nil
}

func CreateRepository(reponame, blobname string) *cerr.CustomError {
	a := strings.ToLower(StorageWritePolicy)

	// ensure we have a sane storage write policy
	if a != "allow" && a != "allow_once" && a != "deny" {
		StorageWritePolicy = "ALLOW"
	}

	switch strings.ToLower(RepoFormat) {
	case "apt":
		if payload, e1 := createApt(reponame, blobname); e1 != nil {
			return e1
		} else {
			return sendPayload(reponame, blobname, payload)
		}
	case "yum":
		if payload, e1 := createYum(reponame, blobname); e1 != nil {
			return e1
		} else {
			return sendPayload(reponame, blobname, payload)
		}
	case "docker":
		if payload, e1 := createDocker(reponame, blobname); e1 != nil {
			return e1
		} else {
			return sendPayload(reponame, blobname, payload)
		}
	case "maven":
		if payload, e1 := createMaven(reponame, blobname); e1 != nil {
			return e1
		} else {
			return sendPayload(reponame, blobname, payload)
		}
	case "raw", "helm", "cargo", "npm", "nuget", "pypi":
		if payload, e1 := createGeneric(reponame, blobname); e1 != nil {
			return e1
		} else {
			return sendPayload(reponame, blobname, payload)
		}
	default:
		return &cerr.CustomError{Title: "Failed to create repository", Message: RepoFormat + " is not a supported format"}
	}
}

func sendPayload(repo, blob string, payload []byte) *cerr.CustomError {
	if len(bytes.TrimSpace(payload)) == 0 {
		return &cerr.CustomError{Title: "Invalid JSON payload", Message: "payload is empty"}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	path := "/service/rest/v1/repositories/" + RepoFormat + "/" + RepoType
	headers := http.Header{}
	headers.Set("Accept", "application/json")
	headers.Set("Content-Type", "application/json")

	resp, e2 := c.Do(context.Background(), http.MethodPost, path, nil, bytes.NewReader(payload), headers)
	if e2 != nil {
		return e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return &cerr.CustomError{
			Title:   "Unable to create repository",
			Message: fmt.Sprintf("HTTP %s: %s", resp.Status, string(body)),
		}
	}

	if !shared.QuietOutput {
		fmt.Printf("%s %s using blob store %s\n", hftx.EnabledSign("Succesfully created repository"),
			hftx.Green(repo), hftx.Green(blob))
	}
	return nil
}
