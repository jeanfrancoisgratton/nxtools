// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original filename: src/env/add_remove.go
// Original timestamp: 2023/09/15 08:23

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	hf "github.com/jeanfrancoisgratton/helperFunctions/v5"
)

func RemoveEnvFile(envfile string) *cerr.CustomError {
	if !strings.HasSuffix(envfile, ".json") {
		envfile += ".json"
	}
	if err := os.Remove(filepath.Join(os.Getenv("HOME"), ".config", "JFG", "nxtools", envfile)); err != nil {
		return &cerr.CustomError{Title: "Error removing " + envfile, Message: err.Error()}
	}

	fmt.Printf("%s removed succesfully\n", envfile)
	return nil
}

func AddEnvFile(envfile string) *cerr.CustomError {
	var env EnvironmentStruct
	if !strings.HasSuffix(envfile, ".json") {
		envfile += ".json"
	}
	env.NexusServerUrl = hf.GetStringValFromPrompt("Enter the server URL (ex: https://myserver:myport): ")
	env.Username = hf.GetStringValFromPrompt("Enter your username: ")
	env.Password = hf.GetPassword("Enter your password: ", DebugMode)
	env.Comments = hf.GetStringValFromPrompt("[OPTIONAL] Enter your comments: ")

	if env.NexusServerUrl == "" || env.Username == "" || env.Password == "" {
		return &cerr.CustomError{Title: "Error creating environment file", Message: "Nexus URL, username and password cannot be empty"}
	} else {
		env.Password = hf.EncodeString(env.Password, "")
	}
	return env.SaveEnvironmentFile(envfile)
}
