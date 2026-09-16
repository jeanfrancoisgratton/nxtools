// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/09/16
// Original filename: src/assets/apt_signing.go

package assets

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hftx "github.com/jeanfrancoisgratton/helperFunctions/v5/terminalfx"
	"nxtools/repositories"
	"nxtools/shared"
)

const aptSignKeyBits = 4096

// CreateSignedAptRepo is the entry point for `--sign` on the APT format:
// generate a fresh RSA-4096 PGP signing keypair, register it with Nexus as
// the repo's aptSigning key, and save both halves locally. Unlike Alpine,
// Nexus does not need to be asked what identifier it assigned the key —
// createApt's existing bring-your-own-keyfile path already hands Nexus a
// caller-supplied PGP key as-is with no discovery step, so the same is
// assumed to hold for a generated one.
func CreateSignedAptRepo(reponame, blobname, distro, saveDir string) *cerr.CustomError {
	reponame = strings.TrimSpace(reponame)
	blobname = strings.TrimSpace(blobname)
	if reponame == "" || blobname == "" {
		return &cerr.CustomError{Title: "Missing parameters", Message: "repository name and blob store name are required"}
	}
	if strings.TrimSpace(distro) == "" {
		distro = "nexus"
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign(fmt.Sprintf("Generating RSA-%d PGP signing keypair", aptSignKeyBits)))
	}
	privArmored, pubArmored, err := generateAptSigningKeypair(reponame, repositories.RepoSigningPassphrase)
	if err != nil {
		return err
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.InProgressSign("Creating APT hosted repository " + reponame))
	}
	if err = repositories.CreateSignedAptHostedRepo(reponame, blobname, distro, privArmored, repositories.RepoSigningPassphrase); err != nil {
		return err
	}

	if err = saveAptSigningKeypair(saveDir, reponame, privArmored, pubArmored); err != nil {
		return &cerr.CustomError{
			Title: "Repository created but signing key could not be saved",
			Message: reponame + " was created and signed, but nxtools could not save the generated keypair to disk: " +
				err.Message + ". Recover the public key from the private key material before distributing it to APT clients.",
		}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Repository " + hftx.Green(reponame) + " is signed with a freshly generated PGP key"))
	}
	return nil
}

// generateAptSigningKeypair returns a fresh RSA-4096 OpenPGP keypair,
// ASCII-armored, as the format Nexus's aptSigning.keypair field expects (an
// exported PGP private key block — this is real GPG signing, unlike
// Alpine's homegrown raw-RSA scheme). If passphrase is non-empty, the
// exported private key material is actually encrypted with it, matching
// what the aptSigning.passphrase field tells Nexus to expect.
func generateAptSigningKeypair(reponame, passphrase string) (privArmored, pubArmored string, cErr *cerr.CustomError) {
	config := &packet.Config{RSABits: aptSignKeyBits}

	// NewEntity assembles these into "name (comment) <email>" and rejects '(', ')', '<', '>'
	// in any individual field (they're the assembled line's own structural separators), so
	// the comment has to carry the descriptive text rather than the name.
	entity, err := openpgp.NewEntity(reponame, "nxtools-generated APT signing key", "", config)
	if err != nil {
		return "", "", &cerr.CustomError{Title: "Failed to generate PGP key", Message: err.Error()}
	}

	if passphrase != "" {
		if err = entity.PrivateKey.Encrypt([]byte(passphrase)); err != nil {
			return "", "", &cerr.CustomError{Title: "Failed to encrypt private key", Message: err.Error()}
		}
		for _, sub := range entity.Subkeys {
			if sub.PrivateKey == nil {
				continue
			}
			if err = sub.PrivateKey.Encrypt([]byte(passphrase)); err != nil {
				return "", "", &cerr.CustomError{Title: "Failed to encrypt subkey", Message: err.Error()}
			}
		}
	}

	var privBuf bytes.Buffer
	privWriter, err := armor.Encode(&privBuf, openpgp.PrivateKeyType, nil)
	if err != nil {
		return "", "", &cerr.CustomError{Title: "Failed to armor-encode private key", Message: err.Error()}
	}
	// SerializePrivate (not WithoutSigning) re-signs identities using the primary key as a
	// signer, which requires it to still be decrypted — fine before Encrypt() above, but not
	// after. The self-signatures NewEntity already produced are valid and don't need redoing.
	if err = entity.SerializePrivateWithoutSigning(privWriter, nil); err != nil {
		return "", "", &cerr.CustomError{Title: "Failed to serialize private key", Message: err.Error()}
	}
	if err = privWriter.Close(); err != nil {
		return "", "", &cerr.CustomError{Title: "Failed to finalize private key", Message: err.Error()}
	}

	var pubBuf bytes.Buffer
	pubWriter, err := armor.Encode(&pubBuf, openpgp.PublicKeyType, nil)
	if err != nil {
		return "", "", &cerr.CustomError{Title: "Failed to armor-encode public key", Message: err.Error()}
	}
	if err = entity.Serialize(pubWriter); err != nil {
		return "", "", &cerr.CustomError{Title: "Failed to serialize public key", Message: err.Error()}
	}
	if err = pubWriter.Close(); err != nil {
		return "", "", &cerr.CustomError{Title: "Failed to finalize public key", Message: err.Error()}
	}

	return privBuf.String(), pubBuf.String(), nil
}

// saveAptSigningKeypair writes both halves of the keypair to disk under the
// repo's name: <reponame>.private.asc (0600) and <reponame>.public.asc
// (0644). Unlike Alpine, APT/dpkg trust doesn't require any particular
// filename convention — sources.list's signed-by= just points at whatever
// path the public key lives at — so the names are chosen for operator
// clarity rather than being load-bearing.
func saveAptSigningKeypair(dir, reponame, privArmored, pubArmored string) *cerr.CustomError {
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		return &cerr.CustomError{Title: "Unable to create key save directory", Message: e.Error()}
	}

	privPath := filepath.Join(dir, reponame+".private.asc")
	pubPath := filepath.Join(dir, reponame+".public.asc")

	if e := os.WriteFile(privPath, []byte(privArmored), 0600); e != nil {
		return &cerr.CustomError{Title: "Unable to save private key", Message: e.Error()}
	}
	if e := os.WriteFile(pubPath, []byte(pubArmored), 0644); e != nil {
		return &cerr.CustomError{Title: "Unable to save public key", Message: e.Error()}
	}

	if !shared.QuietOutput {
		fmt.Println(hftx.EnabledSign("Saved signing keypair to " + hftx.Green(privPath) + " and " + hftx.Green(pubPath)))
	}
	return nil
}
