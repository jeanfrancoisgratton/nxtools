// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 19:31
// Original filename: src/blobstores/list.go

package blobstores

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hf "github.com/jeanfrancoisgratton/helperFunctions/v4"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"nxtools/rest"
	"nxtools/shared"
)

// ListBlobs prints all blobs the configured user can browse.
//
// Endpoint:
//
//	GET /v1/blobstores
func ListBlobs(envFile string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(envFile)
	if err != nil {
		return err
	}

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/blobstores", nil, nil, nil)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return &cerr.CustomError{Title: "Unable to list blobs", Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(b))}
	}

	var blobs []BlobStoreSummary
	dec := json.NewDecoder(resp.Body)
	// The payload may evolve across Nexus versions; be liberal in what we accept.
	if e3 := dec.Decode(&blobs); e3 != nil {
		return &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	fmt.Printf("Number of blob stores: %s\n", hftx.Green(fmt.Sprintf("%d", len(blobs))))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Type", "Available", "Blobcount", "Used size", "Avail size", "Softquota type", "Softquota limit"})

	for _, b := range blobs {
		avail := ""
		sqtype := hftx.Yellow("n/a")
		sqlimit := hftx.Yellow("n/a")
		if !b.Unavailable {
			avail = hftx.EnabledSign("")
		} else {
			avail = hftx.ErrorSign("")
		}
		if b.SoftQuota != nil {
			sqtype = hftx.Green(b.SoftQuota.Type)
			sqlimit = hftx.Green(fmt.Sprintf("%v", b.SoftQuota.Limit))
		}
		t.AppendRow(table.Row{
			hftx.Green(b.Name),
			hftx.Green(b.Type),
			avail, // Unavailable
			//hftx.Green(fmt.Sprintf("%d", b.BlobCount)),
			hftx.Green(hf.SI(b.BlobCount)),
			hftx.Green(shared.FormatSize(b.TotalSizeInBytes)),
			hftx.Green(shared.FormatSize(b.AvailableSpaceInBytes)),
			sqtype,
			sqlimit,
		})
	}

	t.SortBy([]table.SortBy{{Name: "Name", Mode: table.Asc}})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()

	return nil
}
