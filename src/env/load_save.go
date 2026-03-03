// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/env/load_save.go
// Original timestamp: 2023/08/19 10:02

package env

import (
	"encoding/json"
	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hf "github.com/jeanfrancoisgratton/helperFunctions/v4"
	"os"
	"path/filepath"
	"strings"
)

func safeDecodePassword(s string) string {
	// helperFunctions.DecodeString() panics if the input is not a valid
	// ciphertext produced by helperFunctions.EncodeString(). Since users may
	// still create env files manually with a cleartext password, fall back to
	// the original string.
	if strings.TrimSpace(s) == "" {
		return ""
	}

	decoded := s
	func() {
		defer func() {
			if recover() != nil {
				decoded = s
			}
		}()
		decoded = hf.DecodeString(s, "")
	}()

	return decoded
}

// Load the JSON environment file in the user's config directory, and store it into a data type (struct)
func LoadEnvironmentFile() (EnvironmentStruct, *cerr.CustomError) {
	var payload EnvironmentStruct
	var err error

	if !strings.HasSuffix(EnvConfigFile, ".json") {
		EnvConfigFile += ".json"
	}
	rcFile := filepath.Join(os.Getenv("HOME"), ".config", "JFG", "nxtools", EnvConfigFile)
	jFile, err := os.ReadFile(rcFile)
	if err != nil {
		return EnvironmentStruct{}, &cerr.CustomError{Title: err.Error(), Fatality: cerr.Fatal}
	}
	err = json.Unmarshal(jFile, &payload)
	if err != nil {
		return EnvironmentStruct{}, &cerr.CustomError{Title: err.Error(), Fatality: cerr.Fatal}
	} else {
		payload.Password = safeDecodePassword(payload.Password)
		return payload, nil
	}
}

// Save the above structure into a JSON file in the user's config directory
func (e EnvironmentStruct) SaveEnvironmentFile(outputfile string) *cerr.CustomError {
	if outputfile == "" {
		outputfile = EnvConfigFile
	}
	jStream, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return &cerr.CustomError{Title: err.Error(), Fatality: cerr.Fatal}
	}
	rcFile := filepath.Join(os.Getenv("HOME"), ".config", "JFG", "nxtools", outputfile)
	if err = os.WriteFile(rcFile, jStream, 0600); err != nil {
		return &cerr.CustomError{Title: err.Error(), Fatality: cerr.Fatal}
	}

	return nil
}
