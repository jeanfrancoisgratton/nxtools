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

func inferRawDirectory(directory string) string {
	dir := strings.TrimSpace(directory)

	if dir == "" {
		return ""
	}

	// remove leading slash (critical: nexus might have inconsistent behaviour here, from version to version)
	dir = strings.TrimPrefix(dir, "/")

	// normalize separators
	dir = filepath.ToSlash(dir)

	// remove trailing slash
	dir = strings.TrimSuffix(dir, "/")

	return dir
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

func uploadApt(repoName, filePath string) *cerr.CustomError {
	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Uploading " + filepath.Base(filePath) + " to " + repoName))
	}
	if e := shared.UploadComponentMultipart(repoName, "apt.asset", filePath, nil); e != nil {
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
	if e := shared.UploadComponentMultipart(repoName, "helm.asset", filePath, nil); e != nil {
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
	if e := shared.UploadComponentMultipart(repoName, "raw.asset1", filePath, fields); e != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.ErrorSign("Failed to upload " + hftx.Red(filePath) + " to " + hftx.Red(repoName)))
		}
		return e
	}
	if !shared.QuietOutput {
		if directory != "" {
			directory = "/" + directory + "/"
		}
		fmt.Println(hftx.EnabledSign("Uploaded "+hftx.Green(filePath)+" in repository "+hftx.Green(repoName)) + " as " + hftx.Green(repoName+directory+filepath.Base(filePath)))
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

	if e := shared.UploadRepositoryPath(repoName, filePath, resolvedDirectory); e != nil {
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
