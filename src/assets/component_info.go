// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/14 19:01
// Original filename: src/assets/package_information.go

package assets

import (
	"context"
	"encoding/json"
	"net/http"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/rest"
	"nxtools/shared"
)

// PackageInfo
// This will list all available versions of a given package in a given repo
// Optional flags provide extra information

func PackageInfo(pkg, repository string, displayOutput bool) ([]ComponentSummary, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/componensts", nil, nil, nil)
	if e2 != nil {
		return nil, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &cerr.CustomError{Title: "Unable to list repositories", Message: "HTTP status code: " + resp.Status}
	}
	var components []ComponentSummary
	dec := json.NewDecoder(resp.Body)
	if e3 := dec.Decode(&components); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	if !displayOutput {
		return components, nil
	}
	return nil, nil
}
