// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/07/20
// Original filename: src/assets/fetch.go

package assets

import (
	"fmt"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/shared"
)

// FetchLatestAsset downloads the latest version of the package named pkg from
// the given repository. It reuses the component search (PackageInfo) to locate
// every version of the package, keeps the assets attached to the latest
// version, and downloads them into the current directory.
//
// The component search sorts by version with the latest first, so the first
// component returned holds the latest version.
func FetchLatestAsset(repository, pkg string) *cerr.CustomError {
	repository = strings.TrimSpace(repository)
	pkg = strings.TrimSpace(pkg)

	components, err := PackageInfo(repository, pkg, false)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return &cerr.CustomError{
			Title:   "Package not found",
			Message: fmt.Sprintf("no package named %q was found in repository %q", pkg, repository),
		}
	}

	latestVersion := strings.TrimSpace(components[0].Version)

	var toDownload []shared.AssetSummary
	for _, component := range components {
		if strings.TrimSpace(component.Version) != latestVersion {
			continue
		}
		toDownload = append(toDownload, filterAssetsByFormat(component.Format, component.Assets)...)
	}

	if len(toDownload) == 0 {
		return &cerr.CustomError{
			Title:   "No downloadable asset",
			Message: fmt.Sprintf("package %q (version %q) in repository %q has no downloadable asset", pkg, latestVersion, repository),
		}
	}

	for _, asset := range toDownload {
		if strings.TrimSpace(asset.DownloadURL) == "" {
			continue
		}
		if e := DownloadAsset(asset.DownloadURL, ""); e != nil {
			return e
		}
	}

	return nil
}
