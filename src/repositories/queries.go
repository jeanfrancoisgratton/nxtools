// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/10 01:15
// Original filename: src/repositories/queries.go

package repositories

import (
	cerr "github.com/jeanfrancoisgratton/customError/v3"
)

// QueryRepoType returns the format of a repo (docker, yum, apt, etc) for a given repo name.
func QueryRepoType(reponame string) (string, *cerr.CustomError) {
	repo, err := GetRepositorySummary(reponame)
	if err != nil {
		return "", err
	}

	return repo.Format, nil
}
