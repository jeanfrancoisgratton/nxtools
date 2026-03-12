// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/repositories/types.go

package repositories

var RepoUploadDirectory = ""

// RepositorySummary is the (limited) repository representation returned by:
//
//	GET /service/rest/v1/repositories
//
// Fields may vary depending on format/type.
type RepositorySummary struct {
	Name       string         `json:"name"`
	Format     string         `json:"format"`
	Type       string         `json:"type"`
	URL        string         `json:"url"`
	Size       uint64         `json:"size,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// YumHostedRepository contains the fields needed by upload helpers when
// inferring a default upload path for hosted Yum repositories.
type YumHostedRepository struct {
	Name   string              `json:"name"`
	Format string              `json:"format"`
	Type   string              `json:"type"`
	Yum    YumHostedAttributes `json:"yum"`
}

// YumHostedAttributes contains the Yum-specific hosted settings returned by
// GET /service/rest/v1/repositories/yum/hosted/{repositoryName}.
type YumHostedAttributes struct {
	RepodataDepth int    `json:"repodataDepth"`
	DeployPolicy  string `json:"deployPolicy,omitempty"`
}
