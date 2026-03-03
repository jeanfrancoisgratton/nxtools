// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
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
