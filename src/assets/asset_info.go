// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/12 01:59
// Original filename: src/assets/asset_info.go

package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hfjson "github.com/jeanfrancoisgratton/helperFunctions/v4/prettyjson"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
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
		return e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &cerr.CustomError{Title: "Unable to fetch asset info", Message: "HTTP status code: " + resp.Status}
	}

	//body, e3 := io.ReadAll(resp.Body)
	//if e3 != nil {
	//	return &cerr.CustomError{Title: "Unable to read server response", Message: e3.Error()}
	//}

	if JsonOutput {
		body, e3 := io.ReadAll(resp.Body)
		if e3 != nil {
			return &cerr.CustomError{Title: "Unable to read server response", Message: e3.Error()}
		}
		if e4 := hfjson.Print(body); e4 != nil {
			return &cerr.CustomError{Title: "Unable to parse server response", Message: e4.Error()}
		}
		return nil
	}

	var asset AssetSummary
	dec := json.NewDecoder(resp.Body)
	if e3 := dec.Decode(&asset); e3 != nil {
		return &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}
	tabulateSummaryOutput(asset)
	return nil
}

func tabulateSummaryOutput(aInfo AssetSummary) {
	w := tabwriter.NewWriter(os.Stdout, 1, 4, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Name"), aInfo.Path[1:])
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("ID"), aInfo.ID)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("File size"), shared.FormatSize(aInfo.FileSize))
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Download URL"), aInfo.DownloadURL)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Content type"), aInfo.ContentType)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Repository name"), aInfo.Repository)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Repository format"), aInfo.Format)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Last modified"), aInfo.LastModified.Format("2006.01.02 15:04:05"))
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Last downloaded"), aInfo.LastDownloaded.Format("2006.01.02 15:04:05"))
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Uploaded by"), aInfo.Uploader)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Uploader IP address"), aInfo.UploaderIP)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Stored in blobstore"), aInfo.BlobStoreName)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", hftx.Blue("Blobstore created"), aInfo.BlobCreated.Format("2006.01.02 15:04:05"))
	w.Flush()
}
