// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/repositories/list.go

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hfjson "github.com/jeanfrancoisgratton/helperFunctions/v5/prettyjson"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"nxtools/rest"
	"nxtools/shared"
)

// ListRepositories prints all repositories the configured user can browse.
//
// Endpoint:
//
//	GET /service/rest/v1/repositories
func ListRepositories(displayOutput bool) ([]RepositorySummary, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/repositories", nil, nil, nil)
	if e2 != nil {
		return nil, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &cerr.CustomError{Title: "Unable to list repositories", Message: "HTTP status code: " + resp.Status}
	}

	var repos []RepositorySummary
	dec := json.NewDecoder(resp.Body)
	if e3 := dec.Decode(&repos); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	if repos, err = getStorageSpecs(c, repos); err != nil {
		return nil, err
	}

	if !displayOutput {
		return repos, nil
	}

	if RepoListJSONOutput {
		body, e4 := json.Marshal(repos)
		if e4 != nil {
			return nil, &cerr.CustomError{Title: "Unable to encode repository list", Message: e4.Error()}
		}
		if e5 := hfjson.Print(body); e5 != nil {
			return nil, &cerr.CustomError{Title: "Unable to format repository list", Message: e5.Error()}
		}
		return repos, nil
	}

	fmt.Printf("Number of repositories: %s\n", hftx.Green(fmt.Sprintf("%d", len(repos))))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Format", "Type", "Blob store", "Strict validation", "Write policy", "URL"})

	for _, r := range repos {
		t.AppendRow(table.Row{
			hftx.Green(r.Name),
			hftx.Green(r.Format),
			hftx.Green(r.Type),
			hftx.Green(r.Storage.BlobStoreName),
			hftx.Green(fmt.Sprintf("%t", r.Storage.StrictContentTypeValidation)),
			hftx.Green(strings.ToLower(r.Storage.WritePolicy)),
			hftx.Green(r.URL),
		})
	}

	t.SortBy([]table.SortBy{{Name: "Name", Mode: table.Asc}})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()

	return repos, nil
}

func getStorageSpecs(c *rest.Client, repos []RepositorySummary) ([]RepositorySummary, *cerr.CustomError) {
	for i := range repos {
		if strings.ToLower(repos[i].Type) != "hosted" {
			continue
		}

		path := fmt.Sprintf(
			"/service/rest/v1/repositories/%s/hosted/%s",
			url.PathEscape(repos[i].Format),
			url.PathEscape(repos[i].Name),
		)

		resp, err := c.Do(context.Background(), http.MethodGet, path, nil, nil, nil)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return nil, &cerr.CustomError{Title: "Unable to fetch repository details", Message: "HTTP status code: " + resp.Status}
		}

		switch strings.ToLower(repos[i].Format) {
		case "apt":
			var repoSettings AptRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "yum":
			var repoSettings YumRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "docker":
			var repoSettings DockerRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "maven", "maven2":
			var repoSettings MavenRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "raw":
			var repoSettings RawRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "helm":
			var repoSettings HelmRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "cargo":
			var repoSettings CargoRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "npm":
			var repoSettings NpmRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "nuget":
			var repoSettings NugetRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		case "pypi":
			var repoSettings PypiRepoSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		default:
			var repoSettings HostedRepoCommonSettings
			if err := json.NewDecoder(resp.Body).Decode(&repoSettings); err != nil {
				_ = resp.Body.Close()
				return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
			}
			repos[i].Storage = repoSettings.Storage
		}

		_ = resp.Body.Close()
	}

	return repos, nil
}
