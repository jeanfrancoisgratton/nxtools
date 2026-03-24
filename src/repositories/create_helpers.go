// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/23 20:52
// Original filename: src/repositories/create_helpers.go

package repositories

import (
	"encoding/json"
	"os"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
)

func createApt(reponame, blobname string) ([]byte, *cerr.CustomError) {
	kp, err := readFileAsString(RepoSigningFile)
	if err != nil {
		return nil, err
	}

	payload := AptRepoSettingsStruct{
		HostedRepoCommonSettingsStruct: HostedRepoCommonSettingsStruct{
			Name:   reponame,
			Online: true,
			Storage: StorageSpecStruct{
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
	return nil, nil
}

func createMaven(reponame, blobname string) ([]byte, *cerr.CustomError) {
	return nil, nil
}

func createDocker(reponame, blobname string) ([]byte, *cerr.CustomError) {
	return nil, nil
}

func createGeneric(reponame, blobname string) ([]byte, *cerr.CustomError) {
	return nil, nil
}

func readFileAsString(path string) (string, *cerr.CustomError) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", &cerr.CustomError{Title: "failed to read file", Message: err.Error()}
	}
	return string(b), nil
}
