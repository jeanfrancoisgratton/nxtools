// nxtools
// Written by J.F.Gratton <jean-francois@famillegratton.net>
// Original filename: src/repositories/grouprepos_types.go
// Original timestamp: 2026/04/27 04:47:46

package repositories

type GroupAttributesStruct struct {
	MemberNames []string `json:"memberNames"`
}

type ContentDisposition string

const (
	ContentDispositionInline     ContentDisposition = "INLINE"
	ContentDispositionAttachment ContentDisposition = "ATTACHMENT"
)

type RawExtraAttributesStruct struct {
	ContentDisposition ContentDisposition `json:"contentDisposition"`
}

// GENERIC FORMAT

type GroupedRepoCommonAttributesStruct struct {
	Name    string                    `json:"name"`
	Online  bool                      `json:"online"`
	Storage StorageAttributesStruct   `json:"storage"`
	Group   GroupAttributesStruct     `json:"group"`
	Raw     *RawExtraAttributesStruct `json:"raw,omitempty"`
}
