// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/19 02:26
// Original filename: src/repositories/reposettings_types.go

package repositories

// The data types in this file deal with service/rest/v1/repositories/$REPOFORMAT/hosted/$REPONAME API endpoints

// ---------------------------
// COMMON TYPES TO ALL FORMATS
// ---------------------------

// StorageSpec describes the blog store associated with the repo

type StorageSpec struct {
	BlobStoreName               string `json:"blobStoreName"`
	StrictContentTypeValidation bool   `json:"strictContentTypeValidation,omitempty"`
	WritePolicy                 string `json:"writePolicy,omitempty"`
}

// CleanupPolicySpec deals with the repo's cleanup policies, if any
type CleanupPolicySpec struct {
	PolicyNames []string `json:"policyNames,omitempty"`
}

// ComponentSpec details if the repo allows for proprietary components
type ComponentSpec struct {
	ProprietaryComponents bool `json:"proprietaryComponents,omitempty"`
}

// ============================================================
// COMMON REPO FORMAT SHARED FIELDS

type HostedRepoCommonSettings struct {
	Name      string            `json:"name"`
	Format    string            `json:"format"`
	Type      string            `json:"type"`
	Url       string            `json:"url"`
	Online    bool              `json:"online"`
	Storage   StorageSpec       `json:"storage,omitempty"`
	Cleanup   CleanupPolicySpec `json:"cleanup,omitempty"`
	Component ComponentSpec     `json:"component,omitempty"`
}

// ============================================================
// RAW FORMAT

type RawRepoSettings struct {
	HostedRepoCommonSettings
}

// ============================================================
// HELM FORMAT

type HelmRepoSettings struct {
	HostedRepoCommonSettings
}

// ============================================================
// CARGO FORMAT

type CargoRepoSettings struct {
	HostedRepoCommonSettings
}

// ============================================================
// NPM FORMAT

type NpmRepoSettings struct {
	HostedRepoCommonSettings
}

// ============================================================
// NUGET FORMAT

type NugetRepoSettings struct {
	HostedRepoCommonSettings
}

// ============================================================
// PYPI FORMAT

type PypiRepoSettings struct {
	HostedRepoCommonSettings
}

// ============================================================
// APT FORMAT

type AptRepoSettings struct {
	HostedRepoCommonSettings
	Apt struct {
		Distribution string `json:"distribution,omitempty"`
	} `json:"apt,omitempty"`
	AptSigning struct {
		Keypair    string `json:"keypair,omitempty"`
		Passphrase string `json:"passphrase,omitempty"`
	} `json:"aptsigning,omitempty"`
}

// ============================================================
// YUM FORMAT

type YumRepoSettings struct {
	HostedRepoCommonSettings
	Yum struct {
		RepodataDepth int    `json:"repodataDepth"`
		DeployPolicy  string `json:"deployPolicy"`
	} `json:"yum"`
}

// ============================================================
// DOCKER FORMAT

type DockerRepoSettings struct {
	HostedRepoCommonSettings
	Docker struct {
		V1Enabled      bool   `json:"v1Enabled"`
		ForceBasicAuth bool   `json:"forceBasicAuth"`
		HttpPort       int    `json:"httpPort"`
		HttpsPort      int    `json:"httpsPort"`
		Subdomain      string `json:"subdomain"`
		PathEnabled    bool   `json:"pathEnabled"`
	} `json:"docker"`
}

// ============================================================
// MAVEN FORMAT

type MavenRepoSettings struct {
	HostedRepoCommonSettings
	Maven struct {
		VersionPolicy      string `json:"versionPolicy"`
		LayoutPolicy       string `json:"layoutPolicy"`
		ContentDisposition string `json:"contentDisposition"`
	} `json:"maven"`
}
