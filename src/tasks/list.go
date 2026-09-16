// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/09/16 20:00
// Original filename: src/tasks/list.go

package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"nxtools/rest"
	"nxtools/shared"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// ListTasks prints every task defined on the server. When runningOnly is set,
// only tasks whose CurrentState is RUNNING are included.
//
// Endpoint:
//
//	GET /v1/tasks
func ListTasks(displayOut bool, runningOnly bool) ([]TaskSummary, *cerr.CustomError) {
	items, err := fetchAllTasks()
	if err != nil {
		return nil, err
	}

	if runningOnly {
		var running []TaskSummary
		for _, t := range items {
			if strings.EqualFold(t.CurrentState, "RUNNING") {
				running = append(running, t)
			}
		}
		items = running
	}

	if !displayOut {
		return items, nil
	}

	fmt.Printf("Number of tasks: %s\n", hftx.Green(fmt.Sprintf("%d", len(items))))
	printTaskTable(items, table.Row{"ID", "Name", "Type", "State", "Schedule", "Last run", "Next run", "Last result"},
		func(t TaskSummary) table.Row {
			return table.Row{hftx.Green(t.ID), hftx.Green(truncateName(t.Name)), hftx.Green(t.Type), stateLabel(t.CurrentState), deref(t.Schedule),
				formatTimestamp(t.LastRun), formatTimestamp(t.NextRun), deref(t.LastRunResult)}
		})

	return items, nil
}

// fetchAllTasks walks every page of GET /v1/tasks and returns the combined item list.
func fetchAllTasks() ([]TaskSummary, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	var all []TaskSummary
	continuationToken := ""

	for {
		q := url.Values{}
		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		resp, e2 := c.Do(context.Background(), http.MethodGet, "/service/rest/v1/tasks", q, nil, nil)
		if e2 != nil {
			return nil, e2
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return nil, &cerr.CustomError{Title: "Unable to list tasks", Message: "HTTP status code: " + resp.Status}
		}

		var payload ListTasksResponse
		dec := json.NewDecoder(resp.Body)
		if e3 := dec.Decode(&payload); e3 != nil {
			resp.Body.Close()
			return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
		}
		resp.Body.Close()

		all = append(all, payload.Items...)

		if payload.ContinuationToken == nil || strings.TrimSpace(*payload.ContinuationToken) == "" {
			break
		}
		continuationToken = strings.TrimSpace(*payload.ContinuationToken)
	}

	return all, nil
}

func printTaskTable(items []TaskSummary, header table.Row, row func(TaskSummary) table.Row) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)

	for _, item := range items {
		t.AppendRow(row(item))
	}

	t.SortBy([]table.SortBy{{Name: "Name", Mode: table.Asc}})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}

func stateLabel(state string) string {
	switch strings.ToUpper(state) {
	case "RUNNING":
		return hftx.Green(state)
	case "WAITING":
		return hftx.Yellow(state)
	default:
		return state
	}
}

func truncateName(name string) string {
	if len(name) <= maxNameLength {
		return name
	}
	return name[:maxNameLength-3] + "..."
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// formatTimestamp renders a Nexus RFC3339 timestamp (e.g. "2026-09-16T05:00:00.022+00:00")
// without the fractional seconds and UTC offset, which add noise but no useful info here.
func formatTimestamp(s *string) string {
	if s == nil {
		return ""
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return *s
	}
	return t.Format("2006-01-02 15:04:05")
}
