// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 23:24
// Original filename: src/shared/types.go

package shared

import (
	"encoding/json"
	"time"
)

// Global variables and structs

var Envfile = "defaultEnv.json"
var QuietOutput = false

// ListAssetResponse and AssetSummary had to be moved in shared as mnay packages do need it.
// Moving them here prevents circular imports
type ListAssetResponse struct {
	Items             []AssetSummary `json:"items"`
	ContinuationToken *string        `json:"continuationToken"`
}

// AssetSummary represents a single asset item returned by the Nexus REST API.

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
