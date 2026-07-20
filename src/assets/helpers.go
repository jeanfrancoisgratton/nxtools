// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/12 19:03
// Original filename: src/assets/upload_helpers.go

package assets

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/repositories"
	"nxtools/shared"
)

func normalizeUploadFormat(format string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	format = strings.ReplaceAll(format, "_", "")
	format = strings.ReplaceAll(format, "-", "")

	switch format {
	case "apk":
		return "alpine"
	case "ruby", "gem", "gems", "rubygem":
		return "rubygems"
	case "python":
		return "pypi"
	case "golang":
		return "go"
	case "git-lfs":
		return "gitlfs"
	case "conane": // historical typo in an older upload switch
		return "conan"
	default:
		return format
	}
}

func inferRawDirectory(directory string) string {
	return shared.NormalizeDirectory(directory)
}

func inferYumDirectory(repoName, filePath, directory string) (string, *cerr.CustomError) {
	if strings.TrimSpace(directory) != "" {
		return shared.NormalizeDirectory(directory), nil
	}

	repo, err := repositories.GetYumHostedRepository(repoName)
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

func uploadSingleAssetComponent(repoName, filePath, format string) *cerr.CustomError {
	format = normalizeUploadFormat(format)
	spec, ok := singleAssetComponentUploadSpecs[format]
	if !ok || strings.TrimSpace(spec.FileField) == "" {
		return &cerr.CustomError{
			Title:   "Unsupported repository format",
			Message: "repository format " + format + " has no single-asset component upload route configured",
		}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}

	if e := shared.UploadComponentMultipart(repoName, spec.FileField, filePath, nil); e != nil {
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

func uploadApt(repoName, filePath string) *cerr.CustomError {
	return uploadSingleAssetComponent(repoName, filePath, "apt")
}

// uploadAlpine uploads an .apk into an Alpine hosted repository. Unlike the
// plain single-asset formats, Nexus requires two component coordinates on every
// Alpine upload — the Alpine version (e.g. edge, v3.21) and the repository
// section (e.g. main, community). Those cannot be inferred from the .apk itself,
// so they are supplied through the -d/--directory flag as "<version>/<repository>".
func uploadAlpine(repoName, filePath, directory string) *cerr.CustomError {
	version, section, err := parseAlpineCoordinates(directory)
	if err != nil {
		return err
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}

	fields := map[string]string{
		"alpine.version":    version,
		"alpine.repository": section,
	}

	if e := shared.UploadComponentMultipart(repoName, "alpine.asset", filePath, fields); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Uploaded " + hftx.Green(filePath) + " to " + hftx.Green(repoName) + " (" + version + "/" + section + ")"))
	}
	return nil
}

const (
	defaultAlpineVersion    = "edge"
	defaultAlpineRepository = "main"
)

// parseAlpineCoordinates splits the -d/--directory value into the Alpine
// version and repository section. It expects exactly two non-empty path
// segments, e.g. "edge/main" or "v3.21/community". When -d is omitted (the
// flag's "/" default) or empty, it falls back to edge/main so existing
// tooling that never passed -d keeps working. A value with the wrong number
// of segments (e.g. a single "community") is rejected rather than guessed.
func parseAlpineCoordinates(directory string) (string, string, *cerr.CustomError) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(directory), "\\", "/")

	segments := make([]string, 0, 2)
	for _, segment := range strings.Split(cleaned, "/") {
		if segment = strings.TrimSpace(segment); segment != "" {
			segments = append(segments, segment)
		}
	}

	if len(segments) == 0 {
		return defaultAlpineVersion, defaultAlpineRepository, nil
	}

	if len(segments) != 2 {
		return "", "", &cerr.CustomError{
			Title:   "Invalid Alpine coordinates",
			Message: "alpine uploads expect -d <version>/<repository> (e.g. -d edge/main or -d v3.21/community); omit -d to default to edge/main",
		}
	}

	return segments[0], segments[1], nil
}

func uploadRaw(repoName, filePath, directory string) *cerr.CustomError {
	resolvedDirectory := inferRawDirectory(directory)
	fields := map[string]string{
		"raw.directory":       resolvedDirectory,
		"raw.asset1.filename": filepath.Base(filePath),
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}
	if e := shared.UploadComponentMultipart(repoName, "raw.asset1", filePath, fields); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Uploaded "+hftx.Green(filePath)+" in repository "+hftx.Green(repoName)) + " as " + hftx.Green(uploadedTarget(repoName, resolvedDirectory, filepath.Base(filePath))))
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

	fields := map[string]string{
		"yum.directory":      resolvedDirectory,
		"yum.asset.filename": filepath.Base(filePath),
	}

	if e := shared.UploadComponentMultipart(repoName, "yum.asset", filePath, fields); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}
	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Uploaded "+hftx.Green(filePath)+" in repository "+hftx.Green(repoName)) + " as " + hftx.Green(uploadedTarget(repoName, resolvedDirectory, filepath.Base(filePath))))
	}
	return nil
}

func uploadedTarget(repoName, directory, filename string) string {
	dir := shared.NormalizeDirectory(directory)
	if dir == "/" {
		return repoName + "/" + filename
	}
	return repoName + dir + "/" + filename
}
