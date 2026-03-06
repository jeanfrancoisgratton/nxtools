// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 22:59
// Original filename: src/blobstores/remove.go

package blobstores

import (
	"context"
	"fmt"
	"net/http"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

func RemoveBlob(blobs []string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	for _, blob := range blobs {
		resp, e2 := c.Do(context.Background(), http.MethodDelete, "/service/rest/v1/blobstores/"+blob, nil, nil, nil)
		if e2 != nil {
			return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode != 204 {
			return &cerr.CustomError{Title: "Unable to delete blob " + blob, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)}
		}
		fmt.Println(hftx.EnabledSign("Successfully deleted blob " + blob))
	}
	return nil
}
