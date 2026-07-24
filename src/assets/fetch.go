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
//
// Formats that Nexus does not model as first-class components (raw, and any
// format for which Nexus has no dedicated support — e.g. Arch Linux packages
// stored in a raw repository) do not expose a package "name": their component
// name is the full asset path. The component search by name therefore returns
// nothing for them. In that case we fall back to scanning the repository's
// assets and matching them by filename.
func FetchLatestAsset(repository, pkg string) *cerr.CustomError {
	repository = strings.TrimSpace(repository)
	pkg = strings.TrimSpace(pkg)

	components, err := PackageInfo(repository, pkg, false)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		return fetchAssetsByFilename(repository, pkg)
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

	return downloadAssets(toDownload)
}

// fetchAssetsByFilename resolves a package by scanning every asset in the
// repository and matching it against pkg by filename. This is the path used for
// raw repositories (including Arch Linux packages, which Nexus does not model as
// components), where the package name is embedded in the filename rather than in
// component metadata.
//
// A filename matches pkg when its basename is exactly pkg, or when it begins
// with "<pkg><sep><digit>" where sep is '-' or '_' (the version separator used
// by .deb/.rpm/.pkg.tar.zst naming). Requiring the separator to be followed by a
// digit avoids matching a different package that merely shares a prefix (e.g.
// "stubber-utils-1.0..." must not match "stubber").
//
// Among the matches the latest one is kept, using a natural (version-aware)
// comparison of the filename so that, e.g., 2.10 sorts after 2.9.
func fetchAssetsByFilename(repository, pkg string) *cerr.CustomError {
	assets, err := ListAssets(repository, false, false)
	if err != nil {
		return err
	}

	var (
		bestKey  string
		best     []shared.AssetSummary
		haveBest bool
	)
	for _, asset := range assets {
		name := displayAssetName(asset)
		if !filenameMatchesPackage(name, pkg) {
			continue
		}
		key := strings.ToLower(name)
		switch {
		case !haveBest || naturalCompare(key, bestKey) > 0:
			haveBest = true
			bestKey = key
			best = []shared.AssetSummary{asset}
		case key == bestKey:
			best = append(best, asset)
		}
	}

	if !haveBest {
		return &cerr.CustomError{
			Title:   "Package not found",
			Message: fmt.Sprintf("no package named %q was found in repository %q", pkg, repository),
		}
	}

	return downloadAssets(best)
}

// filenameMatchesPackage reports whether the asset basename belongs to package
// pkg. See fetchAssetsByFilename for the matching rules.
func filenameMatchesPackage(basename, pkg string) bool {
	name := strings.ToLower(strings.TrimSpace(basename))
	want := strings.ToLower(strings.TrimSpace(pkg))
	if name == "" || want == "" {
		return false
	}
	if name == want {
		return true
	}
	if !strings.HasPrefix(name, want) {
		return false
	}
	rest := name[len(want):]
	if len(rest) < 2 {
		return false
	}
	if rest[0] != '-' && rest[0] != '_' {
		return false
	}
	return rest[1] >= '0' && rest[1] <= '9'
}

func downloadAssets(assets []shared.AssetSummary) *cerr.CustomError {
	for _, asset := range assets {
		if strings.TrimSpace(asset.DownloadURL) == "" {
			continue
		}
		if e := DownloadAsset(asset.DownloadURL, ""); e != nil {
			return e
		}
	}
	return nil
}

// naturalCompare compares two strings using a version-aware ordering: runs of
// digits are compared numerically (so "2.9" < "2.10"), everything else
// byte-by-byte. It returns -1, 0 or 1.
func naturalCompare(a, b string) int {
	ia, ib := 0, 0
	for ia < len(a) && ib < len(b) {
		da := a[ia] >= '0' && a[ia] <= '9'
		db := b[ib] >= '0' && b[ib] <= '9'
		if da && db {
			ja := ia
			for ja < len(a) && a[ja] >= '0' && a[ja] <= '9' {
				ja++
			}
			jb := ib
			for jb < len(b) && b[jb] >= '0' && b[jb] <= '9' {
				jb++
			}
			na := strings.TrimLeft(a[ia:ja], "0")
			nb := strings.TrimLeft(b[ib:jb], "0")
			if len(na) != len(nb) {
				if len(na) < len(nb) {
					return -1
				}
				return 1
			}
			if na != nb {
				if na < nb {
					return -1
				}
				return 1
			}
			ia, ib = ja, jb
			continue
		}
		if a[ia] != b[ib] {
			if a[ia] < b[ib] {
				return -1
			}
			return 1
		}
		ia++
		ib++
	}
	switch {
	case ia < len(a):
		return 1
	case ib < len(b):
		return -1
	default:
		return 0
	}
}
