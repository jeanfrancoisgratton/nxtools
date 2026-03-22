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
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

func RemoveBlob(blobname string) *cerr.CustomError {
	var blobs []BlobStoreSummary
	var e1 *cerr.CustomError

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	if blobs, e1 = ListBlobs(false); e1 != nil {
		return e1
	}
	for _, blob := range blobs {
		if blob.Name != blobname {
			continue
		}
		if blob.BlobCount > 0 {
			return &cerr.CustomError{Title: "Remove blob " + blobname + " failed", Message: fmt.Sprintf("Blob non-empty (still contains %d assets)", blob.BlobCount)}
		}

		resp, e2 := c.Do(context.Background(), http.MethodDelete, "/service/rest/v1/blobstores/"+blobname, nil, nil, nil)
		if e2 != nil {
			return e2
		}
		defer resp.Body.Close()

		if resp.StatusCode != 204 {
			return &cerr.CustomError{Title: "Unable to delete blob " + blobname, Message: "HTTP status code: " + resp.Status}
		}
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Successfully deleted blob " + blobname))
	}
	return nil
}
