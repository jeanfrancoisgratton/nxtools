// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/repositories/grouprepos.go

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

// groupAPIFormat maps a repository format to the path segment used by the
// group repository endpoints. Maven is exposed as "maven2" by the API.
func groupAPIFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "maven" {
		return "maven2"
	}
	return format
}

// GetGroupRepository fetches the full configuration of a single group-type
// repository, so callers can inspect or amend its member list.
//
// Endpoint:
//
//	GET /service/rest/v1/repositories/{format}/group/{repositoryName}
func GetGroupRepository(name, format string) (GroupedRepoCommonAttributesStruct, *cerr.CustomError) {
	var group GroupedRepoCommonAttributesStruct

	name = strings.TrimSpace(name)
	if name == "" || strings.TrimSpace(format) == "" {
		return group, &cerr.CustomError{Title: "Missing parameters", Message: "repository name and format are required"}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return group, err
	}

	path := "/service/rest/v1/repositories/" + groupAPIFormat(format) + "/group/" + url.PathEscape(name)
	resp, e2 := c.Do(context.Background(), http.MethodGet, path, nil, nil, nil)
	if e2 != nil {
		return group, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return group, &cerr.CustomError{Title: "Unable to fetch group repository", Message: "HTTP status code: " + resp.Status + ": " + string(body), Code: resp.StatusCode}
	}

	if e3 := json.NewDecoder(resp.Body).Decode(&group); e3 != nil {
		return group, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	return group, nil
}
