// internal/rules/embed.go
package rules

import "embed"

//go:embed *.yml VERSION
var EmbeddedRules embed.FS
