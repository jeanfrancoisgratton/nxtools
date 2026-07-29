// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/24 21:55
// Original filename: src/repositories/migrate.go

package assets

import (
	"fmt"
	"os"
	"path"
	"strings"

	"nxtools/repositories"
	"nxtools/shared"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
)

// This implements a feature that does not currently exist in NxRM's API :
// The migration command is actually a copy/move command : we copy assets from a repo, to another one
// If the target repo does not exist, it gets created with the same config/settings than the source
// If the -k flag is set, it is a copy command (ie: we keep the source repo's contents once the actions are completed)

// The target repository does not need to exist ahead of using this command. If the repo does not exit, it will be created
// Mostly using the same config of the source repo, and the target repo will be associated with the same blobstore as the
// Source repo. You create the target repo ahead of time when you wish to migrate in a new blob store as well.

func MigrateRepo(oldrepo, newrepo string) *cerr.CustomError {
	var rs []repositories.RepositorySummary
	var oldr, newr repositories.RepositorySummary
	var ce *cerr.CustomError
	var nMigratedAssets, nTotalAssets uint

	if rs, ce = repositories.ListRepositories(false); ce != nil {
		return ce
	}

	// find out if oldrepo and newrepo exist, and fetch their configs
	for _, r := range rs {
		if r.Name == oldrepo {
			oldr = r
		}
		if r.Name == newrepo {
			newr = r
		}
	}

	// The source repo has not been found
	if oldr.Name == "" {
		return &cerr.CustomError{Title: "The source repository does not exist.", Message: "It might have been misspelled ?"}
	}

	// Sanity check : we do not allow migrations of proxied or grouped repos
	if strings.ToLower(oldr.Type) != "hosted" {
		return &cerr.CustomError{Fatality: cerr.Warning, Title: "Migration not allowed", Message: "Migrations are only allowed for hosted type repositories"}
	}

	// If the target repo does not exist, we have to create it, based on oldrepo's config
	if newr.Name == "" {
		newr = oldr // copying oldrepo config to newrepo
		newr.Name = newrepo
		if ce = repositories.CreateRepository(newrepo, newr.Storage.BlobStoreName); ce != nil {
			return ce
		}
	}

	// Now we download each asset, one at a time from source repo, and upload it to the new repo
	if nMigratedAssets, nTotalAssets, ce = migrateAssets(oldrepo, newrepo, newr.Format); ce != nil {
		return ce
	}

	// Do we keep the old repo ?
	if !repositories.KeepSource {
		if ce = repositories.DeleteRepository(oldrepo); ce != nil {
			return ce
		} else {
			if !shared.QuietOutput {
				fmt.Println(hftx.EnabledSign("Source repository " + hftx.Green(oldrepo) + " has been deleted."))
			}
		}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign(fmt.Sprintf("Migration from %s to %s is complete. %d out of %d assets were migrated",
			hftx.Green(oldrepo), hftx.Green(newrepo), nMigratedAssets, nTotalAssets)))
	}
	return nil
}

/*
	This is the actual migration action. The procedure is thus:

1. Create a temp directory
2. Download the assets from oldrepo, one file at a time
3. Upload it right away to newrepo
4. Remove it from the temp directory
5. Delete the temp directory

The reason that we download and immediately upload that file and then delete it is to avoid filling up the filesystem
*/
func migrateAssets(oldrepo, newrepo, rformat string) (uint, uint, *cerr.CustomError) {
	var nMovedAssets, nTotalAssets uint
	quiet := shared.QuietOutput

	items, me1 := ListAssets(oldrepo, false, false)
	if me1 != nil {
		return 0, 0, me1
	}

	if len(items) == 0 {
		return 0, 0, &cerr.CustomError{Fatality: cerr.Warning, Title: "No assets found in repository", Message: "Nothing to migrate"}
	}
	repositories.RepoFormat = strings.ToLower(items[0].Format)

	nTotalAssets = uint(len(items))

	// Download/upload each asset through a scratch directory so we never clobber files in
	// the caller's working directory and never leave stragglers behind if a transfer fails.
	tmpDir, mkErr := os.MkdirTemp("", "nxtools-migrate-")
	if mkErr != nil {
		return nMovedAssets, nTotalAssets, &cerr.CustomError{Title: "Failed to create temporary directory", Message: mkErr.Error()}
	}
	defer os.RemoveAll(tmpDir)

	for _, item := range items {
		targetFile := path.Join(tmpDir, path.Base(item.Path))

		// Get the file from the source repository
		if !shared.QuietOutput {
			fmt.Println(hftx.InProgressSign("Downloading " + hftx.Green(item.DownloadURL)))
		}
		shared.QuietOutput = false
		if me2 := DownloadAsset(item.DownloadURL, targetFile); me2 != nil {
			return nMovedAssets, nTotalAssets, me2
		}

		shared.QuietOutput = quiet
		if !shared.QuietOutput {
			fmt.Println(hftx.EnabledSign("Downloaded "+hftx.Green(item.DownloadURL)) + " from " + hftx.Green(oldrepo))
		}

		// File has been downloaded, time to upload
		displayName := path.Base(targetFile)
		if !shared.QuietOutput {
			fmt.Println(hftx.InProgressSign("Uploading " + hftx.Green(displayName)))
		}
		shared.QuietOutput = false
		if me3 := UploadAsset(newrepo, targetFile, ""); me3 != nil {
			return nMovedAssets, nTotalAssets, me3
		}
		shared.QuietOutput = quiet
		if !shared.QuietOutput {
			fmt.Println(hftx.EnabledSign("Uploaded "+hftx.Green(displayName)) + " to " + hftx.Green(newrepo))
		}

		// Ok, everything went fine, erasing the file and incrementing the counter
		if me4 := os.Remove(targetFile); me4 != nil {
			return nMovedAssets, nTotalAssets, &cerr.CustomError{Title: "Failed to remove " + hftx.Red(displayName), Message: me4.Error()}
		}
		nMovedAssets++
	}

	// All assets have been migrated. The following repo formats do not support grouped
	// type repositories, so there are no group memberships to update for them.
	if repositories.RepoFormat == "apt" || repositories.RepoFormat == "alpine" || repositories.RepoFormat == "gitlfs" || repositories.RepoFormat == "cocoapods" ||
		repositories.RepoFormat == "composer" || repositories.RepoFormat == "helm" || repositories.RepoFormat == "huggingface" ||
		repositories.RepoFormat == "p2" || repositories.RepoFormat == "swift" {
		return nMovedAssets, nTotalAssets, nil
	}
	if me5 := updateGroups(oldrepo, newrepo, rformat); me5 != nil {
		return nMovedAssets, nTotalAssets, me5
	}
	return nMovedAssets, nTotalAssets, nil
}

// This is where we remove oldrepo from all groups it belongs to, unless repositories.KeepSource is set
// This is how it goes :
//		1. we loop through all the repos
//		2. is the repo of the "grouped" type ? no -> return to step 1
//		3. is that grouped repo of the same format as old/new repo ? no -> return to step 1
//		4. is that reponame "newrepo" ?
//			yes -> add newrepo to members in that repo
//		5. is that reponame "oldrepo" ? no -> return to step 1
//			yes -> remove from member list if KeepSource is false
//

func updateGroups(old, new, rformat string) *cerr.CustomError {
	repogrps, uge1 := repositories.ListRepositories(false)
	if uge1 != nil {
		return uge1
	}

	for _, repo := range repogrps {
		if strings.ToLower(repo.Type) != "group" {
			continue
		}
		if strings.ToLower(repo.Format) != strings.ToLower(rformat) {
			continue
		}
	}

	return nil
}

//
//	if !repositories.KeepSource {
//		for _, repo := range repogrps {
//			if repo.Format == repositories.RepoFormat {
//				if repo.Name == old {
//					repo.
//				}
//			}
//		}
//	}
//}
