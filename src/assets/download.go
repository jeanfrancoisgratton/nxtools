// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/12 19:00
// Original filename: src/assets/download.go

package assets

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

// DownloadAsset downloads the content at rawURL and saves it to destFile.
func DownloadAsset(rawURL, destFile string) *cerr.CustomError {
	var client *rest.Client
	var err *cerr.CustomError

	if rawURL == "" {
		return &cerr.CustomError{
			Title:   "Missing URL",
			Message: "no URL was provided",
		}
	}
	// destFile might be empty; this means that the outputDir will be ".", and the filename will be the one
	// parsed from the URL
	if destFile == "" {
		u, err := url.Parse(rawURL)
		if err != nil {
			return &cerr.CustomError{Title: "Invalid URL"}
		}
		// Use path.Base for URL paths (always forward slashes)
		destFile = path.Base(u.Path)
	}

	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return &cerr.CustomError{
			Title:   "Invalid URL",
			Message: err.Error(),
		}
	}

	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return &cerr.CustomError{
			Title:   "Unable to create destination directory",
			Message: err.Error(),
		}
	}

	if client, err = rest.NewClientFromEnvFile(shared.Envfile); err != nil {
		return err
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Downloading " + hftx.Blue(filepath.Base(destFile))))
	}
	resp, err := client.Get(rawURL)
	if err != nil {
		return &cerr.CustomError{
			Title:   "Download failed",
			Message: err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &cerr.CustomError{Title: "Download failed",
			Message: fmt.Sprintf("HTTP status code: %s", resp.Status)}
	}

	f, err2 := os.Create(destFile)
	if err2 != nil {
		return &cerr.CustomError{
			Title:   "Unable to create destination file",
			Message: err2.Error(),
		}
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return &cerr.CustomError{
			Title:   "Unable to save downloaded file",
			Message: err.Error(),
		}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Downloaded " + hftx.Green(filepath.Base(destFile))))
	}
	return nil
}
