// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/12 18:04
// Original filename: src/shared/upload_helpers.go

package shared

import (
	"context"
	"fmt"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/rest"
)

func FormatSize(sz int64) string {
	numSize := (float64)(sz) / 1000.0 / 1000.0 // this will give us the size in MB
	if (int)(math.Log10(float64(numSize))) > 2 {
		return fmt.Sprintf("%.3f GB", numSize/1000.0)
	} else {
		return fmt.Sprintf("%.3f MB", numSize)
	}
}

func NormalizeDirectory(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" || dir == "." {
		return "/"
	}

	dir = strings.ReplaceAll(dir, "\\", "/")
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}

	if len(dir) > 1 {
		dir = strings.TrimRight(dir, "/")
		if dir == "" {
			return "/"
		}
	}

	return dir
}

func buildRepositoryAssetPath(repoName, directory, filename string) string {
	repoName = strings.TrimSpace(repoName)
	directory = strings.TrimSpace(directory)
	filename = strings.TrimSpace(filename)

	segments := []string{"", "repository", url.PathEscape(repoName)}
	if directory != "" && directory != "/" && directory != "." {
		for _, segment := range strings.Split(strings.ReplaceAll(directory, "\\", "/"), "/") {
			segment = strings.TrimSpace(segment)
			if segment == "" || segment == "." {
				continue
			}
			segments = append(segments, url.PathEscape(segment))
		}
	}

	segments = append(segments, url.PathEscape(filename))
	return strings.Join(segments, "/")
}

func createMultipartBody(fileField, filePath string, extraFields map[string]string) (io.Reader, string, *cerr.CustomError) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil, "", &cerr.CustomError{Title: "Missing parameters", Message: "filename is required"}
	}

	if _, err := os.Stat(filePath); err != nil {
		return nil, "", &cerr.CustomError{Title: "Unable to access file", Message: err.Error()}
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		defer func() {
			_ = mw.Close()
			_ = pw.Close()
		}()

		for k, v := range extraFields {
			if err := mw.WriteField(k, v); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}

		part, err := mw.CreateFormFile(fileField, filepath.Base(filePath))
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		f, err := os.Open(filepath.Clean(filePath))
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		defer f.Close()

		if _, err = io.Copy(part, f); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
	}()

	return pr, mw.FormDataContentType(), nil
}

func UploadComponentMultipart(repoName, fileField, filePath string, extraFields map[string]string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(Envfile)
	if err != nil {
		return err
	}

	body, contentType, e1 := createMultipartBody(fileField, filePath, extraFields)
	if e1 != nil {
		return e1
	}

	query := url.Values{}
	query.Set("repository", repoName)

	headers := http.Header{}
	headers.Set("Accept", "application/json")
	headers.Set("Content-Type", contentType)

	resp, e2 := c.Do(context.Background(), http.MethodPost, "/service/rest/v1/components", query, body, headers)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		msg := "HTTP status code: " + resp.Status
		if len(strings.TrimSpace(string(payload))) > 0 {
			msg += "; response body: " + strings.TrimSpace(string(payload))
		}
		return &cerr.CustomError{Title: "Upload failed", Message: msg}
	}

	return nil
}

func UploadRepositoryPath(repoName, filePath, directory string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(Envfile)
	if err != nil {
		return err
	}

	f, e1 := os.Open(filepath.Clean(filePath))
	if e1 != nil {
		return &cerr.CustomError{Title: "Unable to open file", Message: e1.Error()}
	}
	defer f.Close()

	headers := http.Header{}
	headers.Set("Accept", "application/json")

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath)))
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	headers.Set("Content-Type", contentType)

	path := buildRepositoryAssetPath(repoName, directory, filepath.Base(filePath))
	resp, e2 := c.Do(context.Background(), http.MethodPut, path, nil, f, headers)
	if e2 != nil {
		return &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		msg := "HTTP status code: " + resp.Status
		if len(strings.TrimSpace(string(payload))) > 0 {
			msg += "; response body: " + strings.TrimSpace(string(payload))
		}
		return &cerr.CustomError{Title: "Upload failed", Message: msg}
	}

	return nil
}
