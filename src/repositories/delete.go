// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/repositories/delete.go

package repositories

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/rest"
	"nxtools/shared"
)

// DeleteRepository deletes a repository by name.
//
// Endpoint (generic):
//
//	DELETE /service/rest/v1/repositories/{repositoryName}
func DeleteRepository(repoName string) *cerr.CustomError {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return &cerr.CustomError{Title: "Missing parameters", Message: "repository name is required"}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	path := "/service/rest/v1/repositories/" + url.PathEscape(repoName)
	resp, e2 := c.Do(context.Background(), http.MethodDelete, path, nil, nil, nil)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to delete repository", Message: "HTTP status code: " + resp.Status}
	}

	return nil
}
