// internal/rules/embed.go
package rules

import (
	"embed"
	"strings"
)

//go:embed *.yml VERSION
var EmbeddedRules embed.FS

// GetRulesetVersion returns the version of the embedded ruleset
func GetRulesetVersion() (string, error) {
	versionBytes, err := EmbeddedRules.ReadFile("VERSION")
	if err != nil {
		return "", err
	}
	// Trim spaces and convert to string
	return strings.TrimSpace(string(versionBytes)), nil
}
