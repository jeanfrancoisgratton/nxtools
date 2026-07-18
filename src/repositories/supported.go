// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/24 18:43
// Original filename: src/repositories/supported.go

package repositories

import (
	"os"

	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// This is clunky as for now, I do not see a way to dynamically maintain that list.
// The closest Nexus' API has to offer is the /service/rest/v1/formats/upload-specs endpoint

var supportedformats = []FormatStatusStruct{
	{Name: "alpine", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "apt", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: nil},
	{Name: "cargo", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "cocoapods", SupportsHosted: nil, SupportsGrouped: nil, SupportsProxied: &FalseVal},
	{Name: "composer", SupportsHosted: nil, SupportsGrouped: nil, SupportsProxied: &FalseVal},
	{Name: "conan", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "conda", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "docker", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "gitlfs", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "go", SupportsHosted: nil, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "helm", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "huggingface", SupportsHosted: nil, SupportsGrouped: nil, SupportsProxied: &FalseVal},
	{Name: "maven", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "npm", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "nuget", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "p2", SupportsHosted: nil, SupportsGrouped: nil, SupportsProxied: &FalseVal},
	{Name: "pub", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "pypi", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "r", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "raw", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "rubygems", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "swift", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "terraform", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
	{Name: "yum", SupportsHosted: &TrueVal, SupportsGrouped: &FalseVal, SupportsProxied: &FalseVal},
}

func getLogicalVal(supportedField *bool) string {
	switch supportedField {
	case &TrueVal:
		return hftx.EnabledSign("yes")
	case &FalseVal:
		return hftx.ErrorSign("no")
	default:
		return hftx.Yellow("n/a")
	}
}

// Lists all repo formats and their status, per repo type

func ListSupportedFormats() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Repository format", "Supports hosted", "Supports grouped", "Supports proxied"})

	for _, sf := range supportedformats {
		t.AppendRow([]interface{}{sf.Name, getLogicalVal(sf.SupportsHosted), getLogicalVal(sf.SupportsGrouped), getLogicalVal(sf.SupportsProxied)})
	}
	t.SortBy([]table.SortBy{
		{Name: "Repository name", Mode: table.Asc},
	})
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}
