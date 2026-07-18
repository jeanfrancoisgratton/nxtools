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
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
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
func ListAssets(repoName string, latestOnly bool, displayOutput bool) ([]shared.AssetSummary, *cerr.CustomError) {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return nil, &cerr.CustomError{Title: "Missing parameters", Message: "repository name is required"}
	}

	repo, err := repositories.GetRepositorySummary(repoName)
	if err != nil {
		return nil, err
	}
	repoFormat := repo.Format

	var items []shared.AssetSummary
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

func listAllAssets(repoName, repoFormat string) ([]shared.AssetSummary, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	var all []shared.AssetSummary
	var continuationToken string

	for {
		q := url.Values{}
		q.Set("repository", repoName)
		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/assets", q, nil, nil)
		if e2 != nil {
			return nil, e2
		}

		var payload shared.ListAssetResponse
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &cerr.CustomError{Title: "Unable to list assets", Message: "HTTP status code: " + resp.Status}
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&payload); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		resp.Body.Close()

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

func listLatestAssets(repoName, repoFormat string) ([]shared.AssetSummary, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var selected []shared.AssetSummary
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
			return nil, e2
		}

		var payload ListComponentResponse
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &cerr.CustomError{Title: "Unable to list latest assets", Message: "HTTP status code: " + resp.Status}
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&payload); err != nil {
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
		}
		defer resp.Body.Close()

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

func shouldIncludeAsset(repoFormat string, item shared.AssetSummary) bool {
	repoFormat = strings.ToLower(strings.TrimSpace(repoFormat))
	name := strings.ToLower(displayAssetName(item))

	switch repoFormat {
	case "apt":
		return strings.HasSuffix(name, ".deb") || strings.HasSuffix(name, ".udeb")
	case "alpine":
		return strings.HasSuffix(name, ".apk")
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

func filterAssetsByFormat(repoFormat string, items []shared.AssetSummary) []shared.AssetSummary {
	filtered := make([]shared.AssetSummary, 0, len(items))
	for _, item := range items {
		if shouldIncludeAsset(repoFormat, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func displayAssetName(item shared.AssetSummary) string {
	p := strings.TrimSpace(item.Path)
	if p == "" {
		return ""
	}
	return pathlib.Base(p)
}

func printAssets(repoName string, latestOnly bool, items []shared.AssetSummary) {
	msg := fmt.Sprintf("Number of assets in repository %s: %s", hftx.Blue(repoName), hftx.Green(fmt.Sprintf("%d", len(items))))
	if latestOnly {
		msg += " (latest version only)"
	}
	fmt.Println(msg)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	if !AlternateInfo {
		t.AppendHeader(table.Row{"Asset name", "Asset ID", "Size", "Last modified", "Download URL"})
	} else {
		t.AppendHeader(table.Row{"Asset name", "Size", "Last modified", "Uploader", "Uploader IP", "Download URL"})
	}

	for _, item := range items {
		if !AlternateInfo {
			t.AppendRow(table.Row{displayAssetName(item), item.ID, shared.FormatSize(item.FileSize),
				item.LastModified.Format("2006.01.02 15:04:05"), item.DownloadURL})
		} else {
			t.AppendRow(table.Row{displayAssetName(item), shared.FormatSize(item.FileSize),
				item.LastModified.Format("2006.01.02 15:04:05"), item.Uploader, item.UploaderIP, item.DownloadURL})
		}
	}

	t.SortBy([]table.SortBy{{Name: "Asset name", Mode: table.Asc}})
	t.SetStyle(table.StyleRounded)
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}
