// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/09
// Original filename: src/repositories/shared.go

package repositories

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v4/terminalfx"
	"nxtools/rest"
	"nxtools/shared"
)

const (
	rpmLeadSize      = 96
	rpmTagArch       = 1022
	rpmTypeString    = 6
	rpmTypeI18N      = 9
	rpmAutoDirPrefix = "packages"
)

type rpmHeaderIndex struct {
	Tag    uint32
	Type   uint32
	Offset uint32
	Count  uint32
}

func getRepositorySummary(repoName string) (*RepositorySummary, *cerr.CustomError) {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return nil, &cerr.CustomError{Title: "Missing parameters", Message: "repository name is required"}
	}

	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return nil, err
	}

	path := "/service/rest/v1/repositories/" + url.PathEscape(repoName)
	resp, e2 := c.Do(context.Background(), http.MethodGet, path, nil, nil, nil)
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		msg := "HTTP status code: " + resp.Status
		if len(strings.TrimSpace(string(body))) > 0 {
			msg += "; response body: " + strings.TrimSpace(string(body))
		}
		return nil, &cerr.CustomError{Title: "Unable to fetch repository details", Message: msg}
	}

	var repo RepositorySummary
	if e3 := json.NewDecoder(resp.Body).Decode(&repo); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	return &repo, nil
}

func getYumHostedRepository(envFile, repoName string) (*YumHostedRepository, *cerr.CustomError) {
	c, err := rest.NewClientFromEnvFile(envFile)
	if err != nil {
		return nil, err
	}

	path := "/service/rest/v1/repositories/yum/hosted/" + url.PathEscape(strings.TrimSpace(repoName))
	resp, e2 := c.Do(context.Background(), http.MethodGet, path, nil, nil, nil)
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "HTTP request failed", Message: e2.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		msg := "HTTP status code: " + resp.Status
		if len(strings.TrimSpace(string(body))) > 0 {
			msg += "; response body: " + strings.TrimSpace(string(body))
		}
		return nil, &cerr.CustomError{Title: "Unable to fetch yum repository details", Message: msg}
	}

	var repo YumHostedRepository
	if e3 := json.NewDecoder(resp.Body).Decode(&repo); e3 != nil {
		return nil, &cerr.CustomError{Title: "Unable to parse server response", Message: e3.Error()}
	}

	return &repo, nil
}

func ensureUploadableRepository(repo *RepositorySummary) *cerr.CustomError {
	if repo == nil {
		return &cerr.CustomError{Title: "Repository lookup failed", Message: "repository details are missing"}
	}
	if !strings.EqualFold(strings.TrimSpace(repo.Type), "hosted") {
		return &cerr.CustomError{
			Title:   "Unsupported repository type",
			Message: fmt.Sprintf("repository %q is of type %q; uploads are only supported to hosted repositories", repo.Name, repo.Type),
		}
	}
	return nil
}

func normalizeDirectory(dir string) string {
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

func uploadComponentMultipart(repoName, fileField, filePath string, extraFields map[string]string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
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

func uploadRepositoryPath(repoName, filePath, directory string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
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

func inferRawDirectory(directory string) string {
	return normalizeDirectory(directory)
}

func inferYumDirectory(repoName, filePath, directory string) (string, *cerr.CustomError) {
	if strings.TrimSpace(directory) != "" {
		return normalizeDirectory(directory), nil
	}

	repo, err := getYumHostedRepository(shared.Envfile, repoName)
	if err != nil {
		return "", err
	}

	depth := repo.Yum.RepodataDepth
	if depth <= 0 {
		return "/", nil
	}

	arch := inferRPMArchitecture(filePath)
	parts := make([]string, 0, depth)
	for i := 0; i < depth; i++ {
		parts = append(parts, rpmAutoDirPrefix)
	}
	parts[len(parts)-1] = arch

	return "/" + strings.Join(parts, "/"), nil
}

func inferRPMArchitecture(filePath string) string {
	if arch, err := readRPMArchitectureFromHeader(filePath); err == nil {
		if seg := normalizeRPMArchitectureSegment(arch); seg != "" {
			return seg
		}
	}

	if arch := readRPMArchitectureFromFilename(filePath); arch != "" {
		return arch
	}

	return rpmAutoDirPrefix
}

func normalizeRPMArchitectureSegment(arch string) string {
	arch = strings.TrimSpace(strings.ToLower(arch))
	if arch == "" {
		return ""
	}

	switch arch {
	case "src", "nosrc":
		return "SRPMS"
	default:
		return arch
	}
}

func readRPMArchitectureFromFilename(filePath string) string {
	name := strings.ToLower(strings.TrimSpace(filepath.Base(filePath)))
	if !strings.HasSuffix(name, ".rpm") {
		return ""
	}

	name = strings.TrimSuffix(name, ".rpm")
	if strings.HasSuffix(name, ".src") || strings.HasSuffix(name, ".nosrc") {
		return "SRPMS"
	}

	idx := strings.LastIndex(name, ".")
	if idx <= 0 || idx >= len(name)-1 {
		return ""
	}

	arch := normalizeRPMArchitectureSegment(name[idx+1:])
	switch arch {
	case "x86_64", "noarch", "aarch64", "arm64", "s390x", "ppc64le", "i686", "i586", "i386", "armhfp", "armv7hl", "SRPMS":
		return arch
	default:
		return ""
	}
}

func readRPMArchitectureFromHeader(filePath string) (string, error) {
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}
	defer f.Close()

	lead := make([]byte, rpmLeadSize)
	if _, err = io.ReadFull(f, lead); err != nil {
		return "", err
	}

	_, _, sigSize, err := readRPMHeaderSection(f)
	if err != nil {
		return "", err
	}

	if pad := alignRPMSection(sigSize) - sigSize; pad > 0 {
		if _, err = io.CopyN(io.Discard, f, int64(pad)); err != nil {
			return "", err
		}
	}

	entries, store, _, err := readRPMHeaderSection(f)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.Tag != rpmTagArch {
			continue
		}
		if entry.Type != rpmTypeString && entry.Type != rpmTypeI18N {
			continue
		}
		start := int(entry.Offset)
		if start < 0 || start >= len(store) {
			continue
		}
		end := start
		for end < len(store) && store[end] != 0 {
			end++
		}
		if end > start {
			return string(store[start:end]), nil
		}
	}

	return "", fmt.Errorf("rpm architecture tag not found")
}

func readRPMHeaderSection(r io.Reader) ([]rpmHeaderIndex, []byte, int64, error) {
	record := make([]byte, 16)
	if _, err := io.ReadFull(r, record); err != nil {
		return nil, nil, 0, err
	}
	if record[0] != 0x8e || record[1] != 0xad || record[2] != 0xe8 || record[3] != 0x01 {
		return nil, nil, 0, fmt.Errorf("invalid rpm header magic")
	}

	nIndex := binary.BigEndian.Uint32(record[8:12])
	storeSize := binary.BigEndian.Uint32(record[12:16])

	entries := make([]rpmHeaderIndex, nIndex)
	for i := uint32(0); i < nIndex; i++ {
		buf := make([]byte, 16)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, nil, 0, err
		}
		entries[i] = rpmHeaderIndex{
			Tag:    binary.BigEndian.Uint32(buf[0:4]),
			Type:   binary.BigEndian.Uint32(buf[4:8]),
			Offset: binary.BigEndian.Uint32(buf[8:12]),
			Count:  binary.BigEndian.Uint32(buf[12:16]),
		}
	}

	store := make([]byte, storeSize)
	if _, err := io.ReadFull(r, store); err != nil {
		return nil, nil, 0, err
	}

	totalSize := int64(16) + int64(nIndex)*16 + int64(storeSize)
	return entries, store, totalSize, nil
}

func alignRPMSection(n int64) int64 {
	if n%8 == 0 {
		return n
	}
	return ((n / 8) + 1) * 8
}

func uploadApt(repoName, filePath string) *cerr.CustomError {
	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}
	if e := uploadComponentMultipart(repoName, "apt.asset", filePath, nil); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Uploaded " + hftx.Green(filePath) + " to " + hftx.Green(repoName)))
	}
	return nil
}

func uploadHelm(repoName, filePath string) *cerr.CustomError {
	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}
	if e := uploadComponentMultipart(repoName, "helm.asset", filePath, nil); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Uploaded " + hftx.Green(filePath) + " to " + hftx.Green(repoName)))
	}
	return nil
}

func uploadRaw(repoName, filePath, directory string) *cerr.CustomError {
	fields := map[string]string{
		"raw.directory":       inferRawDirectory(directory),
		"raw.asset1.filename": filepath.Base(filePath),
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}
	if e := uploadComponentMultipart(repoName, "raw.asset1", filePath, fields); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Uploaded " + hftx.Green(filePath) + " to " + hftx.Green(repoName)))
	}
	return nil
}

func uploadYum(repoName, filePath, directory string) *cerr.CustomError {
	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}
	resolvedDirectory, err := inferYumDirectory(repoName, filePath, directory)
	if err != nil {
		return err
	}
	if e := uploadRepositoryPath(repoName, filePath, resolvedDirectory); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Uploaded " + hftx.Green(filePath) + " to " + hftx.Green(repoName)))
	}
	return nil
}
