// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/03
// Original filename: src/repositories/client.go

package repositories

import (
	"strings"

	cerr "github.com/jeanfrancoisgratton/customError/v3"
	"nxtools/env"
	"nxtools/rest"
)

func newClientFromEnvFile(envFile string) (*rest.Client, *cerr.CustomError) {
	oldEnvFile := env.EnvConfigFile
	if strings.TrimSpace(envFile) == "" {
		envFile = "defaultEnv.json"
	}

	env.EnvConfigFile = envFile
	e, err := env.LoadEnvironmentFile()
	env.EnvConfigFile = oldEnvFile
	if err != nil {
		return nil, err
	}

	// If the env file doesn't specify a host, rest.NewClient() falls back to NEXUS_HOST.
	cfg := rest.Config{
		Host:     strings.TrimSpace(e.NexusServerUrl),
		Username: strings.TrimSpace(e.Username),
		Password: strings.TrimSpace(e.Password),
	}

	c, e2 := rest.NewClient(cfg)
	if e2 != nil {
		return nil, &cerr.CustomError{Title: "Unable to create REST client", Message: e2.Error(), Fatality: cerr.Fatal}
	}

	return c, nil
}
