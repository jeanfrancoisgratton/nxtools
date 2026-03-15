// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/12 18:24
// Original filename: src/repositories/details.go

package repositories

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/rest"
	"nxtools/shared"
)

func GetRepositorySummary(repoName string) (*RepositorySummary, *cerr.CustomError) {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return nil, &cerr.CustomError{Title: "Missing parameters", Message: "repository name is required"}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	path := "/service/rest/v1/repositories/" + url.PathEscape(repoName)
	resp, e2 := c.Do(context.Background(), http.MethodGet, path, nil, nil, nil)
	if e2 != nil {
		return nil, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		msg := "HTTP status code: " + resp.Status
		if len(strings.TrimSpace(string(body))) > 0 {
			msg += "; response body: " + strings.TrimSpace(string(body))
		}
		return nil, &cerr.CustomError{Title: "Unable to fetch repository details", Message: msg}
	}

	var repo RepositorySummary
	if e3 := json.NewDecoder(resp.Body).Decode(&repo); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	return &repo, nil
}

func GetYumHostedRepository(repoName string) (*YumHostedRepository, *cerr.CustomError) {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return nil, &cerr.CustomError{Title: "Missing parameters", Message: "repository name is required"}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	path := "/service/rest/v1/repositories/yum/hosted/" + url.PathEscape(repoName)
	resp, e2 := c.Do(context.Background(), http.MethodGet, path, nil, nil, nil)
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		msg := "HTTP status code: " + resp.Status
		if len(strings.TrimSpace(string(body))) > 0 {
			msg += "; response body: " + strings.TrimSpace(string(body))
		}
		return nil, &cerr.CustomError{Title: "Unable to fetch yum repository details", Message: msg}
	}

	var repo YumHostedRepository
	if e3 := json.NewDecoder(resp.Body).Decode(&repo); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	return &repo, nil
}

func EnsureUploadableRepository(repo *RepositorySummary) *cerr.CustomError {
	if repo == nil {
		return &cerr.CustomError{Title: "Repository lookup failed", Message: "repository details are missing"}
	}
	if !strings.EqualFold(strings.TrimSpace(repo.Type), "hosted") {
		return &cerr.CustomError{
			Title:   "Unsupported repository type",
			Message: "Repository " + repo.Name + " is of type " + repo.Type + "; uploads are only supported (for now) on hosted repositories",
		}
	}

	return nil
}
