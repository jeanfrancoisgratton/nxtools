// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/repositories/list.go

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"nxtools/rest"
	"nxtools/shared"
)

// ListRepositories prints all repositories the configured user can browse.
//
// Endpoint:
//
//	GET /service/rest/v1/repositories
func ListRepositories() *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}

	resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/repositories", nil, nil, nil)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return &cerr.CustomError{Title: "Unable to list repositories", Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(b))}
	}

	var repos []RepositorySummary
	dec := json.NewDecoder(resp.Body)
	// The payload may evolve across Nexus versions; be liberal in what we accept.
	if e3 := dec.Decode(&repos); e3 != nil {
		return &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	fmt.Printf("Number of repositories: %s\n", hftx.Green(fmt.Sprintf("%d", len(repos))))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Format", "Type", "URL"})

	for _, r := range repos {
		t.AppendRow(table.Row{
			hftx.Green(r.Name),
			hftx.Green(r.Format),
			hftx.Green(r.Type),
			hftx.Green(r.URL),
		})
	}

	t.SortBy([]table.SortBy{{Name: "Name", Mode: table.Asc}})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()

	return nil
}
