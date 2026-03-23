// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/14 19:01
// Original filename: src/assets/package_info.go

package assets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hfjson "github.com/jeanfrancoisgratton/helperFunctions/v5/prettyjson"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"nxtools/rest"
	"nxtools/shared"
)

// PackageInfo lists all indexed component entries for a given package name in a
// given repository.
//
// It uses the Search API instead of the Components API because the search
// endpoint is designed for filtering by repository/name/version attributes and
// returns the same component summary structure, including component assets.
func PackageInfo(repository, pkg string, displayOutput bool) ([]ComponentSummary, *cerr.CustomError) {
	pkg = strings.TrimSpace(pkg)
	repository = strings.TrimSpace(repository)

	if pkg == "" || repository == "" {
		return nil, &cerr.CustomError{
			Title:   "Missing parameters",
			Message: "package name and repository are required",
		}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	components, err := searchComponentsByPackage(c, pkg, repository)
	if err != nil {
		return nil, err
	}

	if !displayOutput {
		return components, nil
	}

	if JsonOutput {
		payload := ListComponentResponse{
			Items:             components,
			ContinuationToken: nil,
		}

		body, e2 := json.Marshal(payload)
		if e2 != nil {
			return nil, &cerr.CustomError{
				Title:   "Unable to encode JSON output",
				Message: e2.Error(),
			}
		}

		if e3 := hfjson.Print(body); e3 != nil {
			return nil, &cerr.CustomError{
				Title:   "Unable to parse server response",
				Message: e3.Error(),
			}
		}
		return nil, nil
	}

	printComponentSummary(components)
	return nil, nil
}

func searchComponentsByPackage(c *rest.Client, pkg, repository string) ([]ComponentSummary, *cerr.CustomError) {
	var all []ComponentSummary
	var continuationToken string

	for {
		q := url.Values{}
		q.Set("repository", repository)
		q.Set("name", pkg)
		q.Set("sort", "version")

		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		resp, err := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/search", q, nil, nil)
		if err != nil {
			return nil, err
		}

		var payload ListComponentResponse
		decodeErr := decodePackageSearchResponse(resp, &payload)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}

		for _, component := range payload.Items {
			if strings.EqualFold(strings.TrimSpace(component.Repository), repository) &&
				strings.EqualFold(strings.TrimSpace(component.Name), pkg) {
				all = append(all, component)
			}
		}

		if payload.ContinuationToken == nil || strings.TrimSpace(*payload.ContinuationToken) == "" {
			break
		}
		continuationToken = strings.TrimSpace(*payload.ContinuationToken)
	}

	return all, nil
}

func decodePackageSearchResponse(resp *http.Response, payload *ListComponentResponse) *cerr.CustomError {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to fetch package info", Message: "HTTP status code: " + resp.Status}
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(payload); err != nil {
		return &cerr.CustomError{Title: "Unable to parse server response", Message: err.Error()}
	}

	return nil
}

func printComponentSummary(components []ComponentSummary) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Package name", "Version", "Format", "Repository"})

	for _, item := range components {
		t.AppendRow(table.Row{item.Name, item.Version, item.Format, item.Repository})
	}

	t.SetStyle(table.StyleRounded)
	t.SortBy([]table.SortBy{{Name: "Version", Mode: table.Asc}})
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}
