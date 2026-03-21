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
