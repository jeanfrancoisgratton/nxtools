// nxtools
// Unit tests for APT PGP signing keypair generation. These don't need a live
// server: they generate a keypair and verify it parses back as a valid
// OpenPGP entity with the expected properties.

package assets

import (
	"crypto/rsa"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
)

func TestGenerateAptSigningKeypair_NoPassphrase(t *testing.T) {
	priv, pub, err := generateAptSigningKeypair("test-apt-repo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(priv, "BEGIN PGP PRIVATE KEY BLOCK") {
		t.Errorf("private key is not armored as expected: %q", priv[:min(60, len(priv))])
	}
	if !strings.Contains(pub, "BEGIN PGP PUBLIC KEY BLOCK") {
		t.Errorf("public key is not armored as expected: %q", pub[:min(60, len(pub))])
	}

	el, e := openpgp.ReadArmoredKeyRing(strings.NewReader(priv))
	if e != nil {
		t.Fatalf("generated private key does not parse as a valid OpenPGP keyring: %v", e)
	}
	if len(el) != 1 {
		t.Fatalf("expected exactly one entity, got %d", len(el))
	}
	entity := el[0]

	if entity.PrivateKey == nil {
		t.Fatal("entity has no primary private key")
	}
	if entity.PrivateKey.Encrypted {
		t.Error("private key should not be marked encrypted when no passphrase was given")
	}

	rsaPub, ok := entity.PrimaryKey.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("primary key is not RSA: %T", entity.PrimaryKey.PublicKey)
	}
	if bits := rsaPub.N.BitLen(); bits != aptSignKeyBits {
		t.Errorf("RSA key size = %d bits, want %d", bits, aptSignKeyBits)
	}

	foundUID := false
	for uid := range entity.Identities {
		if strings.Contains(uid, "test-apt-repo") {
			foundUID = true
		}
	}
	if !foundUID {
		t.Errorf("no identity contains the repo name, got identities: %v", entity.Identities)
	}

	// The public export must not leak any private key material.
	if strings.Contains(pub, "PRIVATE KEY") {
		t.Error("public key export contains a private key block")
	}
}

func TestGenerateAptSigningKeypair_WithPassphrase(t *testing.T) {
	priv, _, err := generateAptSigningKeypair("test-apt-repo", "correct-horse")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	el, e := openpgp.ReadArmoredKeyRing(strings.NewReader(priv))
	if e != nil {
		t.Fatalf("generated private key does not parse: %v", e)
	}
	entity := el[0]

	if !entity.PrivateKey.Encrypted {
		t.Fatal("private key should be marked encrypted when a passphrase was given")
	}

	if err := entity.PrivateKey.Decrypt([]byte("wrong-passphrase")); err == nil {
		t.Error("decrypting with the wrong passphrase should fail")
	}
	if err := entity.PrivateKey.Decrypt([]byte("correct-horse")); err != nil {
		t.Errorf("decrypting with the correct passphrase should succeed, got: %v", err)
	}
}
