// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/12 01:59
// Original filename: src/assets/info.go

package assets

import (
	"context"
	"io"
	"net/http"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hfjson "github.com/jeanfrancoisgratton/helperFunctions/v4/prettyjson"
	"nxtools/rest"
	"nxtools/shared"
)

func AssetInformation(assetID string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/assets/"+assetID, nil, nil, nil)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &cerr.CustomError{Title: "Unable to fetch asset info", Message: "HTTP status code: " + resp.Status}
	}

	body, e3 := io.ReadAll(resp.Body)
	if e3 != nil {
		return &cerr.CustomError{Title: "Unable to read server response", Message: e3.Error()}
	}

	if e4 := hfjson.Print(body); e4 != nil {
		return &cerr.CustomError{Title: "Unable to parse server response", Message: e4.Error()}
	}

	return nil
}
