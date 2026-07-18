// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/10 15:58
// Original filename: src/assets/types.go

package assets

import (
	"nxtools/shared"
)

var LatestAssetsOnly bool
var AlternateInfo bool
var UploadDirectory string
var JsonOutput bool
var ReindexRepo = false

type componentUploadSpec struct {
	FileField string
}

var singleAssetComponentUploadSpecs = map[string]componentUploadSpec{
	"apt":       {FileField: "apt.asset"},
	"alpine":    {FileField: "alpine.asset"},
	"helm":      {FileField: "helm.asset"},
	"npm":       {FileField: "npm.asset"},
	"nuget":     {FileField: "nuget.asset"},
	"pypi":      {FileField: "pypi.asset"},
	"r":         {FileField: "r.asset"},
	"rubygems":  {FileField: "rubygems.asset"},
	"cargo":     {FileField: "cargo.asset"},
	"terraform": {FileField: "terraform.asset"},
	"swift":     {FileField: "swift.asset"},
	"gitlfs":    {FileField: "gitlfs.asset"},
	"conan":     {FileField: "conan.asset"},
}

const (
	rpmLeadSize      = 96
	rpmTagArch       = 1022
	rpmTypeString    = 6
	rpmTypeI18N      = 9
	rpmAutoDirPrefix = "packages"
)

type rpmHeaderIndex struct {
	Tag    uint32
	Type   uint32
	Offset uint32
	Count  uint32
}

type ListComponentResponse struct {
	Items             []ComponentSummary `json:"items"`
	ContinuationToken *string            `json:"continuationToken"`
}

// ComponentSummary is the search API representation of one component.
type ComponentSummary struct {
	ID         string                `json:"id"`
	Repository string                `json:"repository"`
	Format     string                `json:"format"`
	Group      string                `json:"group,omitempty"`
	Name       string                `json:"name,omitempty"`
	Version    string                `json:"version,omitempty"`
	Assets     []shared.AssetSummary `json:"assets"`
}
