// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/10 15:58
// Original filename: src/assets/types.go

package assets

import (
	"encoding/json"
	"time"
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

type ListAssetResponse struct {
	Items             []AssetSummary `json:"items"`
	ContinuationToken *string        `json:"continuationToken"`
}

// AssetXO represents a single asset item returned by the Nexus REST API.
type AssetSummary struct {
	DownloadURL    string          `json:"downloadUrl"`
	Path           string          `json:"path"`
	ID             string          `json:"id"`
	Repository     string          `json:"repository"`
	Format         string          `json:"format"`
	Checksum       json.RawMessage `json:"checksum"`
	ContentType    string          `json:"contentType"`
	LastModified   time.Time       `json:"lastModified"`
	LastDownloaded time.Time       `json:"lastDownloaded"`
	Uploader       string          `json:"uploader"`
	UploaderIP     string          `json:"uploaderIp"`
	FileSize       int64           `json:"fileSize"`
	BlobCreated    time.Time       `json:"blobCreated"`
	BlobStoreName  string          `json:"blobStoreName"`
}

type ListComponentResponse struct {
	Items             []ComponentSummary `json:"items"`
	ContinuationToken *string            `json:"continuationToken"`
}

// ComponentSummary is the search API representation of one component.
type ComponentSummary struct {
	ID         string         `json:"id"`
	Repository string         `json:"repository"`
	Format     string         `json:"format"`
	Group      string         `json:"group,omitempty"`
	Name       string         `json:"name,omitempty"`
	Version    string         `json:"version,omitempty"`
	Assets     []AssetSummary `json:"assets"`
}
