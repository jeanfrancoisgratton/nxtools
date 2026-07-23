// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/07/23
// Original filename: src/assets/alpine_signing.go

package assets

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/repositories"
	"nxtools/rest"
	"nxtools/shared"
)

const alpineSignKeyBits = 4096

// Nexus rebuilds an Alpine hosted repo's APKINDEX.tar.gz asynchronously after
// an upload — the upload's HTTP response returns before the index is
// actually available. Confirmed empirically: an immediate GET right after
// upload 404s consistently, but succeeds within ~2s. These bounds give a
// comfortable margin over that.
const (
	alpineIndexPollAttempts = 15
	alpineIndexPollInterval = 500 * time.Millisecond
)

// CreateSignedAlpineRepo is the only supported way to create an Alpine hosted
// repository with nxtools. It exists because of a hard-won lesson: Nexus does
// not preserve the caller-supplied key material's identity for Alpine hosted
// repos. Whatever RSA keypair you hand it at creation time, it re-labels
// internally under its own identifier (observed as "key-<8 hex chars>"), and
// that's the identifier it embeds in every APKINDEX.tar.gz it signs
// (".SIGN.RSA.<identifier>.rsa.pub"). apk-tools does an exact filename match
// against /etc/apk/keys/ using that identifier — it does not try every
// trusted key present — so a perfectly valid key deployed under any other
// filename is silently ignored and the repo is reported "UNTRUSTED", no
// matter how many times the key is regenerated or reformatted.
//
// So instead of asking the operator to bring a key nxtools can't guarantee
// will ever be trusted, this generates a fresh keypair, registers it, learns
// Nexus's identifier for it the only way available (there is no API for
// this — it has to be read back out of a generated index), and saves both
// halves locally under that identifier so they can be deployed as-is.
func CreateSignedAlpineRepo(reponame, blobname, saveDir string) *cerr.CustomError {
	reponame = strings.TrimSpace(reponame)
	blobname = strings.TrimSpace(blobname)
	if reponame == "" || blobname == "" {
		return &cerr.CustomError{Title: "Missing parameters", Message: "repository name and blob store name are required"}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign(fmt.Sprintf("Generating RSA-%d signing keypair", alpineSignKeyBits)))
	}
	privPEM, pubPEM, err := generateAlpineSigningKeypair()
	if err != nil {
		return err
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Creating Alpine hosted repository " + reponame))
	}
	if err = repositories.CreateAlpineHostedRepo(reponame, blobname, privPEM); err != nil {
		return err
	}

	identifier, err := discoverAlpineSigningIdentifier(reponame)
	if err != nil {
		return &cerr.CustomError{
			Title: "Repository created but signing identifier could not be determined",
			Message: reponame + " was created, but nxtools could not learn which key identifier Nexus " +
				"assigned to it: " + err.Message + ". The repo exists but is unusable until this is resolved by hand.",
		}
	}

	if e := saveAlpineSigningKeypair(saveDir, identifier, privPEM, pubPEM); e != nil {
		return e
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Repository " + hftx.Green(reponame) + " is signed with key " + hftx.Green(identifier)))
	}
	return nil
}

// generateAlpineSigningKeypair returns a fresh RSA keypair PEM-encoded the
// way Nexus documents for Alpine hosted repos: PKCS8 private key, no
// encryption, SPKI public key. (This is also the format abuild-keygen does
// NOT produce by default — abuild-keygen/openssl genrsa emit PKCS1. Do not
// "fix" this to PKCS1; that was tried, and it's what broke this in the first
// place. Nexus's own repo-edit UI documents PKCS8 as required.)
func generateAlpineSigningKeypair() (privPEM, pubPEM string, err *cerr.CustomError) {
	key, e := rsa.GenerateKey(rand.Reader, alpineSignKeyBits)
	if e != nil {
		return "", "", &cerr.CustomError{Title: "Failed to generate RSA key", Message: e.Error()}
	}

	privDER, e := x509.MarshalPKCS8PrivateKey(key)
	if e != nil {
		return "", "", &cerr.CustomError{Title: "Failed to encode private key", Message: e.Error()}
	}
	privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}))

	pubDER, e := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if e != nil {
		return "", "", &cerr.CustomError{Title: "Failed to encode public key", Message: e.Error()}
	}
	pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))

	return privPEM, pubPEM, nil
}

// discoverAlpineSigningIdentifier learns the key identifier Nexus assigned to
// a freshly created Alpine hosted repo. There is no API for this (checked
// against Nexus's own swagger spec — nothing under /repositories/alpine
// exposes it). The only way to find out is to force an APKINDEX.tar.gz to be
// generated and read the identifier out of its embedded ".SIGN.RSA.<id>"
// entry, so that's what this does: upload a throwaway placeholder package
// (just a .PKGINFO stub — Nexus 400s on anything that doesn't parse as one,
// but doesn't require real content), download the index it triggers, parse
// the identifier out, then remove the placeholder.
func discoverAlpineSigningIdentifier(reponame string) (string, *cerr.CustomError) {
	placeholderPath, err := writePlaceholderApk()
	if err != nil {
		return "", err
	}
	defer os.Remove(placeholderPath)

	if err = uploadAlpine(reponame, placeholderPath, defaultAlpineVersion+"/"+defaultAlpineRepository); err != nil {
		return "", err
	}

	identifier, err := fetchAlpineSigningIdentifier(reponame)

	// Best-effort cleanup of the placeholder regardless of whether we managed
	// to read the identifier — leaving it behind would silently pollute every
	// real APKINDEX rebuild from here on. Deletion is, like the upload, applied
	// asynchronously: the placeholder's (deliberately incomplete — it has no
	// "C:" checksum field, which apk-tools' index parser requires) entry can
	// still be served for a few seconds after the DELETE call returns
	// (confirmed empirically). Since a real client hitting that window would
	// see "v2 database format error" on an otherwise perfectly valid repo,
	// this waits for the removal to actually propagate before declaring the
	// repo ready, rather than just firing the delete and hoping.
	if delErr := deletePlaceholderAsset(reponame, filepath.Base(placeholderPath)); delErr != nil {
		if !shared.QuietOutput {
			fmt.Println(hftx.WarningSign("Could not remove placeholder package from " + reponame + ": " + delErr.Message))
		}
	} else if waitErr := waitForPlaceholderGone(reponame); waitErr != nil && !shared.QuietOutput {
		fmt.Println(hftx.WarningSign(reponame + ": " + waitErr.Message))
	}

	return identifier, err
}

// writePlaceholderApk creates the smallest artifact Nexus will accept as an
// Alpine package: a plain tar.gz containing nothing but a .PKGINFO stub. No
// signature stream, no payload — Nexus only inspects .PKGINFO to extract
// name/version/arch for the index entry. Anything without it is rejected
// with a 400 before it's ever stored (confirmed empirically: raw bytes get
// "Failed to parse APK to extract architecture: No .PKGINFO found in APK
// archive").
func writePlaceholderApk() (string, *cerr.CustomError) {
	const pkginfo = "pkgname = nxtools-placeholder\n" +
		"pkgver = 0.0.0-r0\n" +
		"pkgdesc = placeholder package used only to trigger APKINDEX generation\n" +
		"arch = x86_64\n" +
		"size = 0\n"

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	if e := tw.WriteHeader(&tar.Header{Name: ".PKGINFO", Mode: 0644, Size: int64(len(pkginfo))}); e != nil {
		return "", &cerr.CustomError{Title: "Failed to build placeholder package", Message: e.Error()}
	}
	if _, e := tw.Write([]byte(pkginfo)); e != nil {
		return "", &cerr.CustomError{Title: "Failed to build placeholder package", Message: e.Error()}
	}
	if e := tw.Close(); e != nil {
		return "", &cerr.CustomError{Title: "Failed to build placeholder package", Message: e.Error()}
	}
	if e := gz.Close(); e != nil {
		return "", &cerr.CustomError{Title: "Failed to build placeholder package", Message: e.Error()}
	}

	f, e := os.CreateTemp("", "nxtools-placeholder-*.apk")
	if e != nil {
		return "", &cerr.CustomError{Title: "Failed to create temp file", Message: e.Error()}
	}
	defer f.Close()
	if _, e = f.Write(buf.Bytes()); e != nil {
		return "", &cerr.CustomError{Title: "Failed to write placeholder package", Message: e.Error()}
	}
	return f.Name(), nil
}

// fetchAlpineSigningIdentifier downloads the repo's just-generated
// APKINDEX.tar.gz and extracts the identifier from its
// ".SIGN.RSA.<id>.rsa.pub" entry. It polls: Nexus rebuilds the index
// asynchronously after an upload, so the file routinely 404s for a second or
// two right after the upload's HTTP response comes back (confirmed
// empirically — an immediate GET fails consistently, a GET ~2s later
// succeeds). Any non-404 failure is returned immediately rather than
// retried, since that's not the "not generated yet" condition this is
// working around.
func fetchAlpineSigningIdentifier(reponame string) (string, *cerr.CustomError) {
	var lastErr *cerr.CustomError
	for attempt := 0; attempt < alpineIndexPollAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(alpineIndexPollInterval)
		}
		identifier, notFound, err := fetchAlpineSigningIdentifierOnce(reponame)
		if err == nil {
			return identifier, nil
		}
		if !notFound {
			return "", err
		}
		lastErr = err
	}
	return "", &cerr.CustomError{
		Title: "APKINDEX never appeared",
		Message: fmt.Sprintf("gave up after %d attempts over ~%v waiting for %s's index to be generated: %s",
			alpineIndexPollAttempts, alpineIndexPollAttempts*alpineIndexPollInterval, reponame, lastErr.Message),
	}
}

// fetchAlpineSigningIdentifierOnce is a single, non-retrying attempt. The
// bool return reports whether the failure was specifically a 404 (index not
// generated yet), so the caller knows whether retrying makes sense.
func fetchAlpineSigningIdentifierOnce(reponame string) (identifier string, notFound bool, err *cerr.CustomError) {
	c, cerr2 := rest.NewClientFromEnvFile(shared.Envfile)
	if cerr2 != nil {
		return "", false, cerr2
	}

	indexPath := "/repository/" + reponame + "/" + defaultAlpineVersion + "/" + defaultAlpineRepository + "/x86_64/APKINDEX.tar.gz"
	resp, e2 := c.Do(context.Background(), http.MethodGet, indexPath, nil, nil, nil)
	if e2 != nil {
		return "", false, e2
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", resp.StatusCode == http.StatusNotFound,
			&cerr.CustomError{Title: "Unable to fetch APKINDEX", Message: "HTTP " + resp.Status + ": " + string(body)}
	}

	data, e3 := io.ReadAll(resp.Body)
	if e3 != nil {
		return "", false, &cerr.CustomError{Title: "Unable to read APKINDEX", Message: e3.Error()}
	}

	gz, e4 := gzip.NewReader(bytes.NewReader(data))
	if e4 != nil {
		return "", false, &cerr.CustomError{Title: "Unable to decompress APKINDEX", Message: e4.Error()}
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, e5 := tr.Next()
		if e5 == io.EOF {
			break
		}
		if e5 != nil {
			return "", false, &cerr.CustomError{Title: "Unable to parse APKINDEX", Message: e5.Error()}
		}
		if strings.HasPrefix(hdr.Name, ".SIGN.RSA.") && strings.HasSuffix(hdr.Name, ".rsa.pub") {
			id := strings.TrimSuffix(strings.TrimPrefix(hdr.Name, ".SIGN.RSA."), ".rsa.pub")
			if id == "" {
				return "", false, &cerr.CustomError{Title: "Malformed signature entry", Message: "signature entry name " + hdr.Name + " has no identifier"}
			}
			return id, false, nil
		}
	}
	return "", false, &cerr.CustomError{Title: "No signature entry found", Message: "APKINDEX.tar.gz for " + reponame + " has no .SIGN.RSA.* entry"}
}

// waitForPlaceholderGone polls until the repo's index either 404s (repo back
// to empty) or no longer mentions the placeholder package name, confirming
// the DELETE actually propagated. Uses the same poll budget as the upload
// side, since both were observed to have comparable propagation delay.
func waitForPlaceholderGone(reponame string) *cerr.CustomError {
	c, err := rest.NewClientFromEnvFile(shared.Envfile)
	if err != nil {
		return err
	}
	indexPath := "/repository/" + reponame + "/" + defaultAlpineVersion + "/" + defaultAlpineRepository + "/x86_64/APKINDEX.tar.gz"

	for attempt := 0; attempt < alpineIndexPollAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(alpineIndexPollInterval)
		}

		resp, e2 := c.Do(context.Background(), http.MethodGet, indexPath, nil, nil, nil)
		if e2 != nil {
			return e2
		}
		status := resp.StatusCode
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if status == http.StatusNotFound {
			return nil // repo is empty again — placeholder is gone
		}
		if status != http.StatusOK {
			continue // transient — keep polling within budget
		}
		if !bytes.Contains(data, []byte("P:nxtools-placeholder\n")) {
			return nil
		}
	}

	return &cerr.CustomError{
		Title: "Placeholder removal not confirmed",
		Message: fmt.Sprintf("the placeholder package was deleted, but its index entry was still present after %v; "+
			"it should clear on its own shortly, but an `apk update` run in the meantime may report a format error",
			alpineIndexPollAttempts*alpineIndexPollInterval),
	}
}

// deletePlaceholderAsset finds the placeholder package we just uploaded (by
// its on-disk basename, which is also the filename Nexus stores it under)
// and removes it.
func deletePlaceholderAsset(reponame, filename string) *cerr.CustomError {
	assetList, err := ListAssets(reponame, false, false)
	if err != nil {
		return err
	}
	for _, a := range assetList {
		if filepath.Base(a.Path) == filename {
			return DeleteAssets([]string{a.ID})
		}
	}
	return &cerr.CustomError{Title: "Placeholder not found", Message: filename + " was not found in " + reponame}
}

// saveAlpineSigningKeypair writes both halves of the keypair to disk, named
// after Nexus's own identifier for the key so they can be deployed exactly
// as-is: <identifier> (private, 0600) and <identifier>.rsa.pub (public,
// 0644) — the ".rsa.pub" suffix is not decorative, it's the exact filename
// pattern apk-tools looks for in /etc/apk/keys/.
func saveAlpineSigningKeypair(dir, identifier, privPEM, pubPEM string) *cerr.CustomError {
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		return &cerr.CustomError{Title: "Unable to create key save directory", Message: e.Error()}
	}

	privPath := filepath.Join(dir, identifier)
	pubPath := filepath.Join(dir, identifier+".rsa.pub")

	if e := os.WriteFile(privPath, []byte(privPEM), 0600); e != nil {
		return &cerr.CustomError{Title: "Unable to save private key", Message: e.Error()}
	}
	if e := os.WriteFile(pubPath, []byte(pubPEM), 0644); e != nil {
		return &cerr.CustomError{Title: "Unable to save public key", Message: e.Error()}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Saved signing keypair to " + hftx.Green(privPath) + " and " + hftx.Green(pubPath)))
	}
	return nil
}
