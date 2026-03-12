// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/11
// Original filename: src/assets/list.go

package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	pathlib "path"
	"sort"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"nxtools/repositories"
	"nxtools/rest"
	"nxtools/shared"
)

// ListAssets lists all assets for a given repository.
//
// Standard mode uses:
//
//	GET /service/rest/v1/assets?repository=<name>
//
// Latest-only mode uses component search sorted by version, then keeps only
// the first version seen for a given logical package identity and flattens its
// package assets:
//
//	GET /service/rest/v1/search?repository=<name>&sort=version
func ListAssets(repoName string, latestOnly bool, displayOutput bool) ([]AssetSummary, *cerr.CustomError) {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return nil, &cerr.CustomError{Title: "Missing parameters", Message: "repository name is required"}
	}

	repoFormat, err := repositories.QueryRepoType(repoName)
	if err != nil {
		return nil, err
	}

	var items []AssetSummary
	if latestOnly {
		items, err = listLatestAssets(repoName, repoFormat)
	} else {
		items, err = listAllAssets(repoName, repoFormat)
	}
	if err != nil {
		return nil, err
	}

	sort.SliceStable(items, func(i, j int) bool {
		return displayAssetName(items[i]) < displayAssetName(items[j])
	})

	if !displayOutput {
		return items, nil
	}

	printAssets(repoName, latestOnly, items)
	return items, nil
}

func listAllAssets(repoName, repoFormat string) ([]AssetSummary, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	var all []AssetSummary
	var continuationToken string

	for {
		q := url.Values{}
		q.Set("repository", repoName)
		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/assets", q, nil, nil)
		if e2 != nil {
			return nil, &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
		}

		var payload ListAssetResponse
		decodeErr := decodeAssetResponse(resp, &payload)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}

		for _, item := range payload.Items {
			if shouldIncludeAsset(repoFormat, item) {
				all = append(all, item)
			}
		}

		if payload.ContinuationToken == nil || strings.TrimSpace(*payload.ContinuationToken) == "" {
			break
		}
		continuationToken = strings.TrimSpace(*payload.ContinuationToken)
	}

	return all, nil
}

func listLatestAssets(repoName, repoFormat string) ([]AssetSummary, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var selected []AssetSummary
	var continuationToken string

	for {
		q := url.Values{}
		q.Set("repository", repoName)
		q.Set("sort", "version")
		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/search", q, nil, nil)
		if e2 != nil {
			return nil, &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
		}

		var payload ListComponentResponse
		decodeErr := decodeComponentResponse(resp, &payload)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}

		for _, component := range payload.Items {
			filteredAssets := filterAssetsByFormat(repoFormat, component.Assets)
			if len(filteredAssets) == 0 {
				continue
			}

			key := buildComponentIdentityKey(component)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			selected = append(selected, filteredAssets...)
		}

		if payload.ContinuationToken == nil || strings.TrimSpace(*payload.ContinuationToken) == "" {
			break
		}
		continuationToken = strings.TrimSpace(*payload.ContinuationToken)
	}

	return selected, nil
}

func decodeAssetResponse(resp *http.Response, payload *ListAssetResponse) *cerr.CustomError {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to list assets", Message: "HTTP status code: " + resp.Status}
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(payload); err != nil {
		return &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
	}

	return nil
}

func decodeComponentResponse(resp *http.Response, payload *ListComponentResponse) *cerr.CustomError {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to list latest assets", Message: "HTTP status code: " + resp.Status}
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(payload); err != nil {
		return &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
	}

	return nil
}

func buildComponentIdentityKey(component ComponentSummary) string {
	format := strings.TrimSpace(component.Format)
	group := strings.TrimSpace(component.Group)
	name := strings.TrimSpace(component.Name)

	if group != "" || name != "" {
		return format + "|" + group + "|" + name
	}

	if len(component.Assets) > 0 {
		path := strings.TrimSpace(component.Assets[0].Path)
		version := strings.TrimSpace(component.Version)
		if version != "" {
			path = strings.Replace(path, "/"+version+"/", "/", 1)
			path = strings.Replace(path, "-"+version+".", ".", 1)
			path = strings.Replace(path, "_"+version+".", ".", 1)
		}
		if path != "" {
			return format + "|" + path
		}
	}

	return format + "|" + strings.TrimSpace(component.ID)
}

func shouldIncludeAsset(repoFormat string, item AssetSummary) bool {
	repoFormat = strings.ToLower(strings.TrimSpace(repoFormat))
	name := strings.ToLower(displayAssetName(item))

	switch repoFormat {
	case "apt":
		return strings.HasSuffix(name, ".deb") || strings.HasSuffix(name, ".udeb")
	case "yum":
		return strings.HasSuffix(name, ".rpm")
	case "helm":
		return strings.HasSuffix(name, ".tgz")
	case "docker":
		return true
	default:
		return true
	}
}

func filterAssetsByFormat(repoFormat string, items []AssetSummary) []AssetSummary {
	filtered := make([]AssetSummary, 0, len(items))
	for _, item := range items {
		if shouldIncludeAsset(repoFormat, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func displayAssetName(item AssetSummary) string {
	p := strings.TrimSpace(item.Path)
	if p == "" {
		return ""
	}
	return pathlib.Base(p)
}

func printAssets(repoName string, latestOnly bool, items []AssetSummary) {
	msg := fmt.Sprintf("Number of assets in repository %s: %s", hftx.Blue(repoName), hftx.Green(fmt.Sprintf("%d", len(items))))
	if latestOnly {
		msg += " (latest version only)"
	}
	fmt.Println(msg)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Path", "Size", "Last modified", "Download URL"})

	for _, item := range items {
		t.AppendRow(table.Row{
			displayAssetName(item),
			shared.FormatSize(item.FileSize),
			item.LastModified,
			item.DownloadURL,
		})
	}

	t.SortBy([]table.SortBy{{Name: "Path", Mode: table.Asc}})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}
