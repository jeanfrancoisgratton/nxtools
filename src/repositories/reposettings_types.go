// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/19 02:26
// Original filename: src/repositories/reposettings_types.go

package repositories

var RepoListJSONOutput bool
var RepoFormat string
var RepoType = "hosted"
var RepoSigningFile string
var RepoAptDistro = "nexus"
var RepoSigningPassphrase = ""
var StorageWritePolicy = "ALLOW"
var StorageStrictContentValidation = true
var MavenVersionPolicy = "RELEASE"
var MavenLayoutPolicy = "STRICT"
var MavenContentDisposition = "INLINE"
var YumRepodataDepth uint = 0
var YumDeployPolicy = "PERMISSIVE"

// The data types in this file deal with service/rest/v1/repositories/$REPOFORMAT/hosted/$REPONAME API endpoints

// ---------------------------
// COMMON TYPES TO ALL FORMATS
// ---------------------------

// StorageSpecStruct describes the blog store associated with the repo

type StorageSpecStruct struct {
	BlobStoreName               string `json:"blobStoreName"`
	StrictContentTypeValidation bool   `json:"strictContentTypeValidation,omitempty"`
	WritePolicy                 string `json:"writePolicy,omitempty"`
}

// CleanupPolicySpecStruct deals with the repo's cleanup policies, if any
type CleanupPolicySpecStruct struct {
	PolicyNames []string `json:"policyNames,omitempty"`
}

// ComponentSpecStruct details if the repo allows for proprietary components
type ComponentSpecStruct struct {
	ProprietaryComponents bool `json:"proprietaryComponents,omitempty"`
}

// RepoSigningStruct is the structure used for repos needing gpg key signing (APT, TF)
type RepoSigningStruct struct {
	Keypair    string `json:"keypair,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

// APT FORMAT

type HostedRepoCommonSettingsStruct struct {
	Name      string                   `json:"name"`
	Format    string                   `json:"format,omitempty"`
	Type      string                   `json:"type,omitempty"`
	Url       string                   `json:"url,omitempty"`
	Online    bool                     `json:"online,omitempty"`
	Storage   StorageSpecStruct        `json:"storage,omitempty"`
	Cleanup   *CleanupPolicySpecStruct `json:"cleanup,omitempty"`
	Component *ComponentSpecStruct     `json:"component,omitempty"`
}

type AptRepoSettingsStruct struct {
	HostedRepoCommonSettingsStruct
	Apt struct {
		Distribution string `json:"distribution"`
	} `json:"apt"`
	AptSigning RepoSigningStruct `json:"aptSigning"`
}

// ============================================================
// YUM FORMAT

type YumSettings struct {
	RepodataDepth uint   `json:"repodataDepth,omitempty"`
	DeployPolicy  string `json:"deployPolicy,omitempty"`
}

type YumRepoSettingsStruct struct {
	HostedRepoCommonSettingsStruct
	Yum YumSettings `json:"yum"`
}

// ============================================================
// DOCKER FORMAT

type DockerSettings struct {
	V1Enabled      bool   `json:"v1Enabled,omitempty"`
	ForceBasicAuth bool   `json:"forceBasicAuth,omitempty"`
	HttpPort       int    `json:"httpPort,omitempty"`
	HttpsPort      int    `json:"httpsPort,omitempty"`
	Subdomain      string `json:"subdomain,omitempty"`
	PathEnabled    bool   `json:"pathEnabled,omitempty"`
}
type DockerRepoSettingsStruct struct {
	HostedRepoCommonSettingsStruct
	Docker DockerSettings `json:"docker"`
}

// ============================================================
// MAVEN FORMAT

type MavenSettings struct {
	VersionPolicy      string `json:"versionPolicy"`
	LayoutPolicy       string `json:"layoutPolicy"`
	ContentDisposition string `json:"contentDisposition"`
}

type MavenRepoSettingsStruct struct {
	HostedRepoCommonSettingsStruct
	Maven MavenSettings `json:"maven"`
}
