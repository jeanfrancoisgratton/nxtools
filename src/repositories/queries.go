// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/10 01:15
// Original filename: src/repositories/queries.go

package repositories

import cerr "github.com/jeanfrancoisgratton/customError/v3"

// QueryRepoType returns the type of repo (docker, yum, apt, etc) for a given repo name

func QueryRepoType(reponame string) (string, *cerr.CustomError) {
	var rs []RepositorySummary
	var err *cerr.CustomError
	if rs, err = ListRepositories(false); err != nil {
		return "", err
	}

	for _, repo := range rs {
		if repo.Name == reponame {
			return repo.Format, nil
		}
	}
	return "", &cerr.CustomError{Title: "The repository " + reponame + " was not found"}
}
