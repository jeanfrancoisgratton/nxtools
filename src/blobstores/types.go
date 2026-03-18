// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 19:10
// Original filename: src/blobstores/types.go

package blobstores

// ─── Shared ──────────────────────────────────────────────────────────────────

// SoftQuotaStruct defines an optional space/age constraint on a blob store.
// Type values: "spaceRemainingQuota" | "spaceUsedQuota"
type SoftQuotaStruct struct {
	Type  string `json:"type"`
	Limit int64  `json:"limit"` // in bytes
	// type is of : "enum" : [ "spaceRemainingQuota", "spaceUsedQuota" ]
}

// Blob types
// Available types are : "file", "google", "azure", "s3"
// Currently we only support "file"

var Blobtype = "file"
var FileBlobPath = ""
var SoftQuotaEnabled = false
var SoftQuotaType = "spaceRemainingQuota"
var SoftQuotaLimit int64 = 0

var SoftQuotaSummary = SoftQuotaStruct{Type: "spaceRemainingQuota", Limit: 0}

// ─── List / Summary ───────────────────────────────────────────────────────────

// BlobStoreSummary is returned by GET /v1/blobstores
type BlobStoreSummary struct {
	Name                  string           `json:"name"`
	Type                  string           `json:"type"` // "File", "S3", "Azure", "Group"
	Unavailable           bool             `json:"unavailable"`
	BlobCount             int64            `json:"blobCount"`
	TotalSizeInBytes      int64            `json:"totalSizeInBytes"`
	AvailableSpaceInBytes int64            `json:"availableSpaceInBytes"`
	SoftQuota             *SoftQuotaStruct `json:"softQuota,omitempty"`
}

// QuotaStatusSummary is returned by GET /v1/blobstores/{name}/quota-status
type QuotaStatusSummary struct {
	IsViolation   bool   `json:"isViolation"`
	Message       string `json:"message"`
	BlobStoreName string `json:"blobStoreName"`
}

// ─── File Blob Store ──────────────────────────────────────────────────────────

// FileQuotaConfig
type FileQuotaConfig struct {
	QuotaStatus *QuotaStatusSummary `json:"quotaStatus,omitempty"`
	SoftQuota   *SoftQuotaStruct    `json:"softQuota,omitempty"`
	Path        string              `json:"path"`
}

// FileCreateRequest is the payload for POST /v1/blobstores/file
type FileCreateRequest struct {
	Name      string           `json:"name,omitempty"`
	SoftQuota *SoftQuotaStruct `json:"softQuota,omitempty"`
	Path      string           `json:"path"` // convenience field; maps to attributes
}

// FileUpdateRequest is the payload for PUT /v1/blobstores/file/{name}
// Identical shape to create; name comes from the URL path, not the body.
type FileUpdateRequest struct {
	SoftQuota *SoftQuotaStruct `json:"softQuota,omitempty"`
	Path      string           `json:"path"`
}

// FileResponse is returned by GET /v1/blobstores/file/{name}
type FileResponse struct {
	Name      string           `json:"name"`
	SoftQuota *SoftQuotaStruct `json:"softQuota,omitempty"`
	Path      string           `json:"path"`
}

// ─── S3 Blob Store ────────────────────────────────────────────────────────────

// S3BucketConfiguration describes the target S3 bucket and credentials.
type S3BucketConfiguration struct {
	Region          string `json:"region"`
	Name            string `json:"name"` // bucket name
	Prefix          string `json:"prefix,omitempty"`
	Expiration      int32  `json:"expiration"` // days; -1 = never
	AccessKeyID     string `json:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	// Role-based auth (optional, mutually exclusive with key-based)
	AssumeRole   string `json:"assumeRole,omitempty"`
	SessionToken string `json:"sessionToken,omitempty"`
}

// S3EncryptionConfiguration describes optional server-side encryption.
// EncryptionType values: "none" | "s3ManagedEncryption" | "kmsManagedEncryption"
type S3EncryptionConfiguration struct {
	EncryptionType string `json:"encryptionType"`
	EncryptionKey  string `json:"encryptionKey,omitempty"` // KMS key ARN or alias
}

// S3AdvancedBucketConnection holds optional tunables for the S3 client.
type S3AdvancedBucketConnection struct {
	Endpoint              string `json:"endpoint,omitempty"`       // custom S3-compatible endpoint
	SignerType            string `json:"signerType,omitempty"`     // e.g. "DEFAULT", "S3SignerType"
	ForcePathStyle        *bool  `json:"forcePathStyle,omitempty"` // required for MinIO / Ceph
	MaxConnectionPoolSize *int32 `json:"maxConnectionPoolSize,omitempty"`
}

// S3CreateRequest is the payload for POST /v1/blobstores/s3
type S3CreateRequest struct {
	Name                     string                      `json:"name"`
	SoftQuota                *SoftQuotaStruct            `json:"softQuota,omitempty"`
	BucketConfiguration      S3BucketConfiguration       `json:"bucketConfiguration"`
	Encryption               *S3EncryptionConfiguration  `json:"encryption,omitempty"`
	AdvancedBucketConnection *S3AdvancedBucketConnection `json:"advancedBucketConnection,omitempty"`
}

// S3UpdateRequest is the payload for PUT /v1/blobstores/s3/{name}
type S3UpdateRequest struct {
	SoftQuota                *SoftQuotaStruct            `json:"softQuota,omitempty"`
	BucketConfiguration      S3BucketConfiguration       `json:"bucketConfiguration"`
	Encryption               *S3EncryptionConfiguration  `json:"encryption,omitempty"`
	AdvancedBucketConnection *S3AdvancedBucketConnection `json:"advancedBucketConnection,omitempty"`
}

// S3Response is returned by GET /v1/blobstores/s3/{name}
type S3Response struct {
	Name                     string                      `json:"name"`
	SoftQuota                *SoftQuotaStruct            `json:"softQuota,omitempty"`
	BucketConfiguration      S3BucketConfiguration       `json:"bucketConfiguration"`
	Encryption               *S3EncryptionConfiguration  `json:"encryption,omitempty"`
	AdvancedBucketConnection *S3AdvancedBucketConnection `json:"advancedBucketConnection,omitempty"`
}

// ─── Azure Blob Store ─────────────────────────────────────────────────────────

// AzureAuthenticationMethod values: "accountkey" | "managedidentity" | "environmentvariable"
type AzureBucketConfiguration struct {
	AccountName          string `json:"accountName"`
	ContainerName        string `json:"containerName"`
	AuthenticationMethod string `json:"authenticationMethod"`
	AccountKey           string `json:"accountKey,omitempty"` // only for "accountkey" method
}

// AzureCreateRequest is the payload for POST /v1/blobstores/azure
type AzureCreateRequest struct {
	Name                string                   `json:"name"`
	SoftQuota           *SoftQuotaStruct         `json:"softQuota,omitempty"`
	BucketConfiguration AzureBucketConfiguration `json:"bucketConfiguration"`
}

// AzureUpdateRequest is the payload for PUT /v1/blobstores/azure/{name}
type AzureUpdateRequest struct {
	SoftQuota           *SoftQuotaStruct         `json:"softQuota,omitempty"`
	BucketConfiguration AzureBucketConfiguration `json:"bucketConfiguration"`
}

// AzureResponse is returned by GET /v1/blobstores/azure/{name}
type AzureResponse struct {
	Name                string                   `json:"name"`
	SoftQuota           *SoftQuotaStruct         `json:"softQuota,omitempty"`
	BucketConfiguration AzureBucketConfiguration `json:"bucketConfiguration"`
}

// ─── Group Blob Store ─────────────────────────────────────────────────────────

// GroupFillPolicy values: "roundRobin" | "writeToFirst"
type GroupFillPolicy = string

const (
	FillPolicyRoundRobin GroupFillPolicy = "roundRobin"
	FillPolicyWriteFirst GroupFillPolicy = "writeToFirst"
)

// GroupCreateRequest is the payload for POST /v1/blobstores/group
type GroupCreateRequest struct {
	Name       string           `json:"name"`
	SoftQuota  *SoftQuotaStruct `json:"softQuota,omitempty"`
	Members    []string         `json:"members"` // ordered list of member blob store names
	FillPolicy GroupFillPolicy  `json:"fillPolicy"`
}

// GroupUpdateRequest is the payload for PUT /v1/blobstores/group/{name}
type GroupUpdateRequest struct {
	SoftQuota  *SoftQuotaStruct `json:"softQuota,omitempty"`
	Members    []string         `json:"members"`
	FillPolicy GroupFillPolicy  `json:"fillPolicy"`
}

// GroupResponse is returned by GET /v1/blobstores/group/{name}
type GroupResponse struct {
	Name       string           `json:"name"`
	SoftQuota  *SoftQuotaStruct `json:"softQuota,omitempty"`
	Members    []string         `json:"members"`
	FillPolicy GroupFillPolicy  `json:"fillPolicy"`
}
