// internal/rules/parser.go
package rules

import (
	"fmt"
	"io/fs"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse loads and validates rules from the embedded filesystem
func Parse(embeddedFS fs.FS) ([]Rule, error) {
	var allRules []Rule

	// Read all YAML files from the embedded filesystem
	entries, err := fs.ReadDir(embeddedFS, ".")
	if err != nil {
		return nil, fmt.Errorf("error reading rules directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yml") || strings.HasSuffix(entry.Name(), ".yaml")) {
			// Skip VERSION file
			if entry.Name() == "VERSION" {
				continue
			}

			// Read the file content
			content, err := fs.ReadFile(embeddedFS, entry.Name())
			if err != nil {
				return nil, fmt.Errorf("error reading %s: %w", entry.Name(), err)
			}

			// Parse the YAML content
			var ruleFile RuleFile
			if err := yaml.Unmarshal(content, &ruleFile); err != nil {
				return nil, fmt.Errorf("error parsing %s: %w", entry.Name(), err)
			}

			// Validate rules
			ids := map[string]bool{}
			for _, rule := range ruleFile.Rules {
				if ids[rule.ID] {
					return nil, fmt.Errorf("duplicate id %s in %s", rule.ID, entry.Name())
				}
				ids[rule.ID] = true
			}

			allRules = append(allRules, ruleFile.Rules...)
		}
	}

	return allRules, nil
}
