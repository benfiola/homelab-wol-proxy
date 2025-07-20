package wolproxy

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var versionBytes []byte

// Get the proxy version via an embedded version file
func GetVersion() string {
	version := strings.TrimSpace(string(versionBytes))
	if version == "" {
		version = "0.0.0+undefined"
	}
	return version
}