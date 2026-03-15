// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/13 22:01
// Original filename: src/assets/delete.go

package assets

import (
	"context"
	"fmt"
	"net/http"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

func DeleteAssets(assets []string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	for _, asset := range assets {
		resp, e2 := c.Do(context.Background(), http.MethodDelete, "/service/rest/v1/assets/"+asset, nil, nil, nil)
		if e2 != nil {
			return e2
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return &cerr.CustomError{Title: "Unable to fetch asset info", Message: "HTTP status code: " + resp.Status}
		}
		if !shared.QuietOutput {
			fmt.Println(hftx.EnabledSign("Deleted " + asset))
		}
	}
	return nil
}
