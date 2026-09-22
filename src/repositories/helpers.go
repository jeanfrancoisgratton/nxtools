// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/repositories/helpers.go
// Original timestamp : 2026.09.22 17:54:41

package repositories

import (
	"strings"

	"github.com/jeanfrancoisgratton/customError/v3"
)

// NormalizeWritePolicy validates the --writepolicy value and returns its
// canonical uppercase form ("ALLOW", "ALLOW_ONCE" or "DENY").
func NormalizeWritePolicy(policy string) (string, *customError.CustomError) {
	lower := strings.ToLower(policy)
	if lower != "allow" && lower != "deny" && lower != "allow_once" {
		return "", &customError.CustomError{Title: "Invalid --writepolicy value",
			Message: "Supported policies are ALLOW, ALLOW_ONCE and DENY; you selected " + policy}
	}
	return strings.ToUpper(lower), nil
}

// IsAlpineFormat reports whether format refers to the Alpine/APK repo format,
// which nxtools accepts under either spelling.
func IsAlpineFormat(format string) bool {
	f := strings.ToLower(format)
	return f == "alpine" || f == "apk"
}

// IsAptFormat reports whether format refers to the apt repo format.
func IsAptFormat(format string) bool {
	return strings.ToLower(format) == "apt"
}

// checkAlpineSigningRequirement enforces that --sign was passed when creating
// an Alpine-format repo (its signing keypair is generated and registered by
// nxtools, not supplied via --keyfile like other formats), and returns a
// warning when --keyfile was also set (since it's ignored for this format).
func CheckAlpineSigningRequirement(signingFile string, signFlagChanged bool) (warning string, err *customError.CustomError) {
	if signingFile != "" {
		warning = "-k/--keyfile is ignored for the Alpine format; use --sign instead"
	}
	if !signFlagChanged {
		err = &customError.CustomError{Title: "Missing --sign",
			Message: "You need to pass --sign (optionally --sign=PATH) when using the Alpine format"}
	}
	return warning, err
}

// AptSigningMode is which of --sign / -k, --keyfile a `repo create --format
// apt` invocation should use to obtain its signing key.
type AptSigningMode int

const (
	AptSignKeyfile  AptSigningMode = iota // -k/--keyfile: an existing key is handed to Nexus as-is
	AptSignGenerate                       // --sign: nxtools generates and registers a fresh key
)

// CheckAptSigningRequirement validates --sign/-k usage for `repo create
// --format apt` and reports which mode to use. Exactly one of --sign or
// -k/--keyfile is required; passing both is rejected as ambiguous.
func CheckAptSigningRequirement(signingFile string, signFlagChanged bool) (AptSigningMode, *customError.CustomError) {
	if signFlagChanged && signingFile != "" {
		return 0, &customError.CustomError{Title: "Conflicting signing flags",
			Message: "--sign and -k/--keyfile are mutually exclusive; pass one or the other"}
	}
	if signFlagChanged {
		return AptSignGenerate, nil
	}
	if signingFile == "" {
		return 0, &customError.CustomError{Title: "Missing signing key",
			Message: "You need to provide a PGP private key file (-k/--keyfile) or pass --sign to generate one when using the APT format"}
	}
	return AptSignKeyfile, nil
}
