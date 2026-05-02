// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/23 20:52
// Original filename: src/repositories/create_helpers.go

package repositories

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
)

func createApt(reponame, blobname string) ([]byte, *cerr.CustomError) {
	kp, err := readFileAsString(RepoSigningFile)
	if err != nil {
		return nil, err
	}

	payload := AptRepoSettingsStruct{
		HostedRepoCommonAttributesStruct: HostedRepoCommonAttributesStruct{
			Name:   reponame,
			Online: true,
			Storage: StorageAttributesStruct{
				BlobStoreName:               blobname,
				StrictContentTypeValidation: StorageStrictContentValidation,
				WritePolicy:                 StorageWritePolicy,
			},
		},
		AptSigning: RepoSigningStruct{
			Keypair:    kp,
			Passphrase: RepoSigningPassphrase,
		},
	}
	payload.Apt.Distribution = RepoAptDistro

	pload, e2 := json.MarshalIndent(payload, "", "  ")
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "failed to marshal APT repo payload", Message: e2.Error()}
	}
	return pload, nil
}

func createYum(reponame, blobname string) ([]byte, *cerr.CustomError) {
	// preflight
	if strings.ToLower(YumDeployPolicy) != "PERMISSIVE" && strings.ToLower(YumDeployPolicy) != "STRICT" {
		return nil, &cerr.CustomError{Title: "yum deploy policy not supported", Message: fmt.Sprintf("unsupported policy: %s", YumDeployPolicy)}
	}
	payload := YumRepoSettingsStruct{
		HostedRepoCommonAttributesStruct: HostedRepoCommonAttributesStruct{
			Name:   reponame,
			Online: true,
			Storage: StorageAttributesStruct{
				BlobStoreName:               blobname,
				StrictContentTypeValidation: StorageStrictContentValidation,
				WritePolicy:                 StorageWritePolicy,
			},
		},
		Yum: YumSettings{
			RepodataDepth: YumRepodataDepth,
			DeployPolicy:  YumDeployPolicy,
		},
	}

	pload, e2 := json.MarshalIndent(payload, "", "  ")
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "failed to marshal YUM repo payload", Message: e2.Error()}
	}
	return pload, nil
}

func createMaven(reponame, blobname string) ([]byte, *cerr.CustomError) {
	// some sanity checks here, as Maven can be a bit anal when it comes to JSON payload validation
	a := strings.ToLower(MavenVersionPolicy)
	if a != "release" && a != "snapshot" && a != "mixed" {
		return nil, &cerr.CustomError{Title: "maven version policy not supported", Message: fmt.Sprintf("unsupported version: %s", a)}
	}
	b := strings.ToLower(MavenLayoutPolicy)
	if b != "strict" && b != "permissive" {
		return nil, &cerr.CustomError{Title: "maven layout policy not supported", Message: fmt.Sprintf("unsupported layout: %s", b)}
	}
	c := strings.ToLower(RepoContentDisposition)
	if c != "inline" && c != "attachment" {
		return nil, &cerr.CustomError{Title: "maven content disposition not supported", Message: fmt.Sprintf("unsupported disposition: %s", b)}
	}

	// ok preflight is done, let's proceed
	payload := MavenRepoSettingsStruct{
		HostedRepoCommonAttributesStruct: HostedRepoCommonAttributesStruct{
			Name:   reponame,
			Online: true,
			Storage: StorageAttributesStruct{
				BlobStoreName:               blobname,
				StrictContentTypeValidation: StorageStrictContentValidation,
				WritePolicy:                 StorageWritePolicy,
			},
		},
		Maven: MavenSettings{
			VersionPolicy:      MavenVersionPolicy,
			LayoutPolicy:       MavenLayoutPolicy,
			ContentDisposition: RepoContentDisposition,
		},
	}

	pload, e2 := json.MarshalIndent(payload, "", "  ")
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "failed to marshal MAVEN repo payload", Message: e2.Error()}
	}
	return pload, nil
}

func createDocker(reponame, blobname string) ([]byte, *cerr.CustomError) {
	// some sanity checks
	if DockerHttpPort == 0 && DockerHttpsPort == 0 && DockerSubdomain == "" {
		return nil, &cerr.CustomError{Title: "Docker connection config missing", Message: "You need to set either of an http(s) port or a subdomain"}
	}

	// ok preflight is done, let's proceed
	payload := DockerRepoSettingsStruct{
		HostedRepoCommonAttributesStruct: HostedRepoCommonAttributesStruct{
			Name:   reponame,
			Online: true,
			Storage: StorageAttributesStruct{
				BlobStoreName:               blobname,
				StrictContentTypeValidation: StorageStrictContentValidation,
				WritePolicy:                 StorageWritePolicy,
			},
		},
		Docker: DockerSettings{
			V1Enabled:      DockerV1Enabled,
			ForceBasicAuth: DockerForceBasicAuth,
			HttpPort:       DockerHttpPort,
			HttpsPort:      DockerHttpsPort,
			Subdomain:      DockerSubdomain,
			PathEnabled:    DockerPathEnabled,
		},
	}

	pload, e2 := json.MarshalIndent(payload, "", "  ")
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "failed to marshal Docker repo payload", Message: e2.Error()}
	}
	return pload, nil
}

// this is the general-purpose repo generation function; most repo formats should use this one

func createGeneric(reponame, blobname string) ([]byte, *cerr.CustomError) {
	payload := HostedRepoCommonAttributesStruct{
		Name:   reponame,
		Online: true,
		Storage: StorageAttributesStruct{
			BlobStoreName:               blobname,
			StrictContentTypeValidation: StorageStrictContentValidation,
			WritePolicy:                 StorageWritePolicy,
		},
	}
	// The NPM format mandates strict content validation
	if strings.ToLower(RepoFormat) == "npm" {
		payload.Storage.StrictContentTypeValidation = true
	}
	pload, e2 := json.MarshalIndent(payload, "", "  ")
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "failed to marshal APT repo payload", Message: e2.Error()}
	}
	return pload, nil
}

// readFileAsString reads a file and assigns it to a string variable

func readFileAsString(path string) (string, *cerr.CustomError) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", &cerr.CustomError{Title: "failed to read file", Message: err.Error()}
	}
	return string(b), nil
}
