// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/01 15:41
// Original filename: src/env/types.go

package env

var EnvConfigFile string
var DebugMode = false

// This structure holds the basic software config
type EnvironmentStruct struct {
	NexusServerUrl string `json:"NexusServerUrl,omitempty"`
	Username       string `json:"Username"`
	Password       string `json:"Password"`
}
