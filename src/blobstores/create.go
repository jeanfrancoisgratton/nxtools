package blobstores

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

// This is where we create blob stores
// As of v3.8x, the following blob stores are available: google (gcp), s3 (aws), azure, file, group
// nxtools will eventually support all of them, but for now we only support the file blob type

func CreateBlob(blobname string) *cerr.CustomError {
	bft := strings.TrimSpace(strings.ToLower(Blobtype))
	switch bft {
	case "file":
		return CreateFileBlob(blobname)
	default:
		return &cerr.CustomError{
			Title:   "Invalid Blobtype",
			Message: bft + " is not a supported blob type",
		}
	}
}

// Create a file-based blob store
// ENDPOINT = /service/rest/v1/blobstores/file

func CreateFileBlob(blobname string) *cerr.CustomError {
	var sq *SoftQuotaStruct
	if SoftQuotaEnabled {
		sq = &SoftQuotaStruct{
			Type:  SoftQuotaType,
			Limit: SoftQuotaLimit,
		}
	}

	if FileBlobPath == "" {
		FileBlobPath = blobname
	}

	fcr := FileCreateRequest{
		Name:      blobname,
		Path:      FileBlobPath,
		SoftQuota: sq,
	}

	payload, err := json.Marshal(fcr)
	if err != nil {
		return &cerr.CustomError{
			Title:   "Unable to build JSON payload",
			Message: err.Error(),
		}
	}

	c, e0 := rest.NewClientFromEnvFile(shared.Envfile)
	if e0 != nil {
		return e0
	}

	headers := http.Header{}
	headers.Set("Accept", "application/json")
	headers.Set("Content-Type", "application/json")

	resp, e2 := c.Do(
		context.Background(),
		http.MethodPost,
		"/service/rest/v1/blobstores/file",
		nil,
		bytes.NewReader(payload),
		headers,
	)
	if e2 != nil {
		return &cerr.CustomError{
			Title:   "HTTP request failed",
			Message: e2.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{
			Title:   "Unable to create file blob store",
			Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}
	}

	fmt.Println(hftx.EnabledSign("File-based blob store " + blobname + "with path " + FileBlobPath + " has been created"))
	return nil
}
