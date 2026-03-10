// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/09
// Original filename: src/repositories/upload.go

package repositories

import (
	"os"
	"path/filepath"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
)

func UploadFile(repoName, filePath, directory string) *cerr.CustomError {
	repoName = strings.TrimSpace(repoName)
	filePath = strings.TrimSpace(filePath)
	directory = strings.TrimSpace(directory)

	if repoName == "" {
		return &cerr.CustomError{Title: "Missing parameters", Message: "repository name is required"}
	}
	if filePath == "" {
		return &cerr.CustomError{Title: "Missing parameters", Message: "filename is required"}
	}

	if _, err := os.Stat(filepath.Clean(filePath)); err != nil {
		return &cerr.CustomError{Title: "Unable to access file", Message: err.Error()}
	}

	repo, err := getRepositorySummary(repoName)
	if err != nil {
		return err
	}
	if err = ensureUploadableRepository(repo); err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(repo.Format)) {
	case "apt":
		return uploadApt(repoName, filePath)
	case "helm":
		return uploadHelm(repoName, filePath)
	case "raw":
		return uploadRaw(repoName, filePath, directory)
	case "yum":
		return uploadYum(repoName, filePath, directory)
	default:
		return &cerr.CustomError{
			Title:   "Unsupported repository format",
			Message: "repository format " + repo.Format + " is not supported by nxtools upload",
		}
	}
}
