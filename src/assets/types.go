// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/10 15:58
// Original filename: src/assets/types.go

package assets

import "time"

type ListAssetResponse struct {
	Items             []AssetSummary `json:"items"`
	ContinuationToken *string        `json:"continuationToken"`
}

// AssetXO represents a single asset item returned by the Nexus REST API.
type AssetSummary struct {
	DownloadURL    string            `json:"downloadUrl"`
	Path           string            `json:"path"`
	ID             string            `json:"id"`
	Repository     string            `json:"repository"`
	Format         string            `json:"format"`
	Checksum       map[string]string `json:"checksum"`
	ContentType    string            `json:"contentType"`
	LastModified   time.Time         `json:"lastModified"`
	LastDownloaded time.Time         `json:"lastDownloaded"`
	Uploader       string            `json:"uploader"`
	UploaderIP     string            `json:"uploaderIp"`
	FileSize       int64             `json:"fileSize"`
	BlobCreated    time.Time         `json:"blobCreated"`
	BlobStoreName  string            `json:"blobStoreName"`
}
