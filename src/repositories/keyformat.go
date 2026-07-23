// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/repositories/keyformat.go
//
// NxRM's Alpine (APK) repository API requires the signing key in PKCS1
// ("RSA PRIVATE KEY") PEM form, but most modern tooling (openssl genpkey,
// recent openssl genrsa, etc.) emits PKCS8 ("PRIVATE KEY") by default. This
// file detects the PEM format of a supplied RSA key and converts it to
// PKCS1 when needed, so the same -k flag used for APT's PGP key can be
// reused for Alpine without the caller having to know or care which
// container format their RSA key is in.

package repositories

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
)

// toPKCS1RSAKey normalizes a PEM-encoded RSA private key to PKCS1 form.
// Keys already in PKCS1 form are returned unchanged; PKCS8-wrapped RSA keys
// are unwrapped and re-encoded as PKCS1. Any other key type or format is
// rejected, since Alpine repos require a PKCS1 RSA key.
func toPKCS1RSAKey(pemData string) (string, *cerr.CustomError) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return "", &cerr.CustomError{Title: "invalid signing key", Message: "no PEM-encoded private key found in signing key file"}
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		// already PKCS1
		return pemData, nil

	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", &cerr.CustomError{Title: "invalid signing key", Message: fmt.Sprintf("failed to parse PKCS8 key: %s", err.Error())}
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return "", &cerr.CustomError{Title: "unsupported signing key", Message: "Alpine signing key must be an RSA key"}
		}
		out := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)})
		return string(out), nil

	default:
		return "", &cerr.CustomError{Title: "unsupported signing key format", Message: fmt.Sprintf("expected a PKCS1 or PKCS8 RSA private key, got PEM block type %q", block.Type)}
	}
}
