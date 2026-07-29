// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/09
// Original filename: src/assets/upload.go

package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/repositories"
	"nxtools/shared"
	"nxtools/tasks"
)

// formatsNeedingReindex lists repo formats whose metadata NxRM does not rebuild on its own after
// a component upload, unlike raw/alpine/etc. There is no REST endpoint to ask NxRM to do this
// implicitly, so nxtools triggers the rebuild-metadata task itself right after a successful upload.
var formatsNeedingReindex = map[string]bool{
	"yum": true,
	"apt": true,
}

func UploadAsset(repoName, filePath, directory string) *cerr.CustomError {
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

	repo, err := repositories.GetRepositorySummary(repoName)
	if err != nil {
		return err
	}
	if err = repositories.EnsureUploadableRepository(repo); err != nil {
		return err
	}

	format := normalizeUploadFormat(repo.Format)
	var uploadErr *cerr.CustomError
	switch format {
	case "raw":
		uploadErr = uploadRaw(repoName, filePath, directory)
	case "yum":
		uploadErr = uploadYum(repoName, filePath, directory)
	case "alpine":
		uploadErr = uploadAlpine(repoName, filePath, directory)
	case "apt", "helm", "npm", "nuget", "pypi", "r", "rubygems", "cargo", "terraform", "swift", "gitlfs", "conan":
		uploadErr = uploadSingleAssetComponent(repoName, filePath, format)
	default:
		return &cerr.CustomError{
			Title:   "Unsupported repository format",
			Message: "repository format " + repo.Format + " is not supported by nxtools upload",
		}
	}
	if uploadErr != nil {
		return uploadErr
	}

	// yum/apt do not rebuild their metadata on their own; trigger it now so the uploaded
	// asset is actually visible to consumers. A failure here does not undo the upload, so
	// it's surfaced as a warning rather than an error.
	if formatsNeedingReindex[format] {
		if re := tasks.ReindexRepo(repoName); re != nil && !shared.QuietOutput {
			fmt.Println(hftx.WarningSign("Upload succeeded, but reindexing " + repoName + " failed: " + re.Error()))
		}
	}
	return nil
}
