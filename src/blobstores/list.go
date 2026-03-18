// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 19:31
// Original filename: src/blobstores/list.go

package blobstores

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hf "github.com/jeanfrancoisgratton/helperFunctions/v5"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
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
func ListBlobs() *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/blobstores", nil, nil, nil)
	if e2 != nil {
		return e2
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Unable to list blobs", Message: "HTTP status code: " + resp.Status}
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
	t.AppendHeader(table.Row{"Name", "Type", "Available", "Blobcount", "Path", "Used size", "Avail size", "Softquota type", "Softquota limit", "Quota status"})

	for _, b := range blobs {
		var fqc FileQuotaConfig
		var err *cerr.CustomError
		var qsm string
		avail := ""
		sqtype := hftx.Yellow("n/a")
		sqlimit := hftx.Yellow("n/a")
		if !b.Unavailable {
			avail = hftx.EnabledSign("")
		} else {
			avail = hftx.ErrorSign("")
		}
		if b.Type == "File" {
			if fqc, err = getFileBlobQuotaInformation(c, b.Name); err != nil {
				return err
			}
		}
		if b.SoftQuota != nil {
			sqtype = hftx.Green(b.SoftQuota.Type)
			sqlimit = hftx.Green(shared.FormatSize(b.SoftQuota.Limit))
		}
		if fqc.QuotaStatus != nil {
			if fqc.QuotaStatus.IsViolation {
				qsm = hftx.WarningSign(fqc.QuotaStatus.Message)
			} else {
				qsm = hftx.GreenGoSign("No quota violation")
			}
		} else {
			qsm = hftx.Green("n/a")
		}
		t.AppendRow(table.Row{
			hftx.Green(b.Name),
			hftx.Green(b.Type),
			avail, // Unavailable
			hftx.Green(hf.SI(b.BlobCount)),
			hftx.Green(fqc.Path),
			hftx.Green(shared.FormatSize(b.TotalSizeInBytes)),
			hftx.Green(shared.FormatSize(b.AvailableSpaceInBytes)),
			sqtype,
			sqlimit,
			qsm,
		})
	}

	t.SortBy([]table.SortBy{{Name: "Name", Mode: table.Asc}})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()

	return nil
}

func getFileBlobQuotaInformation(c *rest.Client, bname string) (FileQuotaConfig, *cerr.CustomError) {
	var bi FileQuotaConfig
	var qi FileCreateRequest

	// First, we need to fetch the path of the blob store
	resp, e2 := c.Do(context.Background(), http.MethodGet, "service/rest/v1/blobstores/file/"+bname, nil, nil, nil)
	if e2 != nil {
		return FileQuotaConfig{}, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return FileQuotaConfig{}, &cerr.CustomError{Title: "Unable to list the blob config", Message: "HTTP status code: " + resp.Status}
	}

	//var blobs []BlobStoreSummary
	dec := json.NewDecoder(resp.Body)
	if e3 := dec.Decode(&qi); e3 != nil {
		return FileQuotaConfig{}, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}
	// Second, if there is a soft quota, we need to get its status
	if qi.SoftQuota != nil {
		if qss, err := getFileBlobQuotaViolations(c, bname); err != nil {
			return FileQuotaConfig{}, err
		} else {
			bi = FileQuotaConfig{SoftQuota: qi.SoftQuota, Path: qi.Path, QuotaStatus: qss}
		}
	}
	return bi, nil
}

func getFileBlobQuotaViolations(c *rest.Client, bname string) (*QuotaStatusSummary, *cerr.CustomError) {
	var qs *QuotaStatusSummary

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/blobstores/"+bname+"/quota-status", nil, nil, nil)
	if e2 != nil {
		return nil, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &cerr.CustomError{Title: "Unable to list the blob config", Message: "HTTP status code: " + resp.Status}
	}

	dec := json.NewDecoder(resp.Body)
	if e3 := dec.Decode(&qs); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}
	// reformat the quota status message if there's a violation
	if qs.IsViolation {
		qs.Message = strings.Replace(qs.Message, "Blob store "+bname+" is u", " U", 1)
		qs.Message = strings.Replace(qs.Message, "space and has a limit", "of space out", 1)
	}
	return qs, nil
}
