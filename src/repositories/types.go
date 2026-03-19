// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/12 18:04
// Original filename: src/repositories/types.go

package repositories

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

// HostedRepositoryExtra contains the additional hosted-repository settings
// optionally displayed by `repo ls -x`.
type HostedRepositoryExtra struct {
	BlobStoreName               string `json:"blobStoreName"`
	StrictContentTypeValidation bool   `json:"strictContentTypeValidation"`
	WritePolicy                 string `json:"writePolicy"`
	HasCleanupPolicies          bool   `json:"hasCleanupPolicies"`
	ProprietaryComponents       bool   `json:"proprietaryComponents"`
}

// RepositoryListEntry is the repository representation returned by repo list.
// Extra hosted-repository settings are populated only when ExtraOutput is set.
type RepositoryListEntry struct {
	Name   string                 `json:"name"`
	Format string                 `json:"format"`
	Type   string                 `json:"type"`
	URL    string                 `json:"url"`
	Extra  *HostedRepositoryExtra `json:"extra,omitempty"`
}

// RepoListJSONOutput toggles JSON output for `repo list`.
var RepoListJSONOutput bool

// RepoListExtraOutput toggles extra hosted-repository columns for `repo list`.
var RepoListExtraOutput bool

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

// RepositorySettings is a quick way to tie-in a repository to its blob, see which repo is part of a group, etc
type RepositorySettingsSummary struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Type   string `json:"type"`
}
