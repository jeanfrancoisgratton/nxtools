// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/repositories/list.go

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
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
func ListRepositories(displayOutput bool) ([]RepositoryListEntry, *cerr.CustomError) {
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

	entries := make([]RepositoryListEntry, 0, len(repos))
	for _, repo := range repos {
		entry := RepositoryListEntry{
			Name:   repo.Name,
			Format: repo.Format,
			Type:   repo.Type,
			URL:    repo.URL,
		}

		if RepoListExtraOutput && strings.EqualFold(strings.TrimSpace(repo.Type), "hosted") {
			extra, e4 := getHostedRepositoryExtra(c, repo)
			if e4 != nil {
				return nil, e4
			}
			entry.Extra = extra
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	if !displayOutput {
		return entries, nil
	}

	if RepoListJSONOutput {
		body, e5 := json.Marshal(entries)
		if e5 != nil {
			return nil, &cerr.CustomError{Title: "Unable to encode repository list", Message: e5.Error()}
		}
		if e6 := hfjson.Print(body); e6 != nil {
			return nil, &cerr.CustomError{Title: "Unable to format repository list", Message: e6.Error()}
		}
		return entries, nil
	}

	fmt.Printf("Number of repositories: %s\n", hftx.Green(fmt.Sprintf("%d", len(entries))))
	printRepositoryTable(entries)

	return entries, nil
}

func getHostedRepositoryExtra(c *rest.Client, repo RepositorySummary) (*HostedRepositoryExtra, *cerr.CustomError) {
	path := fmt.Sprintf(
		"/service/rest/v1/repositories/%s/hosted/%s",
		url.PathEscape(strings.TrimSpace(repo.Format)),
		url.PathEscape(strings.TrimSpace(repo.Name)),
	)

	resp, err := c.Do(context.Background(), http.MethodGet, path, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &cerr.CustomError{
			Title:   "Unable to fetch repository details",
			Message: "HTTP status code: " + resp.Status + " for repository " + repo.Name,
		}
	}

	body, e5 := io.ReadAll(resp.Body)
	if e5 != nil {
		return nil, &cerr.CustomError{Title: "Unable to read server response", Message: e5.Error()}
	}

	return decodeHostedRepositoryExtra(strings.TrimSpace(repo.Format), body)
}

func decodeHostedRepositoryExtra(repoFormat string, body []byte) (*HostedRepositoryExtra, *cerr.CustomError) {
	switch strings.ToLower(strings.TrimSpace(repoFormat)) {
	case "apt":
		var settings AptRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "yum":
		var settings YumRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "docker":
		var settings DockerRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "maven", "maven2":
		var settings MavenRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "raw":
		var settings RawRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "helm":
		var settings HelmRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "cargo":
		var settings CargoRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "npm":
		var settings NpmRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "nuget":
		var settings NugetRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	case "pypi":
		var settings PypiRepoSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings.HostedRepoCommonSettings), nil

	default:
		var settings HostedRepoCommonSettings
		if err := json.Unmarshal(body, &settings); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		return hostedExtraFromCommonSettings(settings), nil
	}
}

func hostedExtraFromCommonSettings(settings HostedRepoCommonSettings) *HostedRepositoryExtra {
	return &HostedRepositoryExtra{
		BlobStoreName:               settings.Storage.BlobStoreName,
		StrictContentTypeValidation: settings.Storage.StrictContentTypeValidation,
		WritePolicy:                 settings.Storage.WritePolicy,
		HasCleanupPolicies:          len(settings.Cleanup.PolicyNames) > 0,
		ProprietaryComponents:       settings.Component.ProprietaryComponents,
	}
}

func printRepositoryTable(entries []RepositoryListEntry) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if RepoListExtraOutput {
		t.AppendHeader(table.Row{
			"Name",
			"Format",
			"Type",
			"Blobstore",
			"Content validation",
			"Write policy",
			"Proprietary components",
			"URL",
		})

		for _, entry := range entries {
			blobStoreName := ""
			strictContentValidation := ""
			writePolicy := ""
			proprietaryComponents := ""

			if entry.Extra != nil {
				blobStoreName = entry.Extra.BlobStoreName
				if entry.Extra.StrictContentTypeValidation {
					strictContentValidation = hftx.EnabledSign("")
				} else {
					strictContentValidation = hftx.ErrorSign("")
				}
				writePolicy = strings.ToLower(entry.Extra.WritePolicy)
				if entry.Extra.ProprietaryComponents {
					proprietaryComponents = hftx.EnabledSign("")
				} else {
					proprietaryComponents = hftx.ErrorSign("")
				}
			}

			t.AppendRow(table.Row{
				hftx.Green(entry.Name),
				hftx.Green(entry.Format),
				hftx.Green(entry.Type),
				hftx.Green(blobStoreName),
				hftx.Green(strictContentValidation),
				hftx.Green(writePolicy),
				hftx.Green(proprietaryComponents),
				hftx.Green(entry.URL),
			})
		}
	} else {
		t.AppendHeader(table.Row{"Name", "Format", "Type", "URL"})

		for _, entry := range entries {
			t.AppendRow(table.Row{
				hftx.Green(entry.Name),
				hftx.Green(entry.Format),
				hftx.Green(entry.Type),
				hftx.Green(entry.URL),
			})
		}
	}

	t.SortBy([]table.SortBy{{Name: "Name", Mode: table.Asc}})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}
