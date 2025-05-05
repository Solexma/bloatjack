package cli

import (
	"fmt"
	"strings"

	"github.com/Solexma/bloatjack/internal/rules"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Rule represents a single optimization rule
type Rule struct {
	ID       string            `yaml:"id"`
	Priority int               `yaml:"priority"`
	Match    map[string]string `yaml:"match"`
	If       string            `yaml:"if,omitempty"`
	Set      map[string]string `yaml:"set,omitempty"`
	SetEnv   map[string]string `yaml:"set_env,omitempty"`
	Action   string            `yaml:"action,omitempty"`
	Note     string            `yaml:"note,omitempty"`
}

// RuleFile represents a YAML file containing rules
type RuleFile struct {
	Rules []Rule `yaml:"rules"`
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List and parse embedded rules",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing embedded rules:")

		// Read all YAML files from the embedded filesystem
		entries, err := rules.EmbeddedRules.ReadDir(".")
		if err != nil {
			fmt.Printf("Error reading rules directory: %v\n", err)
			return
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") {
				fmt.Printf("\n=== Rules from %s ===\n", entry.Name())

				// Read the file content
				content, err := rules.EmbeddedRules.ReadFile(entry.Name())
				if err != nil {
					fmt.Printf("Error reading %s: %v\n", entry.Name(), err)
					continue
				}

				// Parse the YAML content
				var ruleFile RuleFile
				if err := yaml.Unmarshal(content, &ruleFile); err != nil {
					fmt.Printf("Error parsing %s: %v\n", entry.Name(), err)
					continue
				}

				// Print each rule
				for _, rule := range ruleFile.Rules {
					fmt.Printf("\nRule: %s (Priority: %d)\n", rule.ID, rule.Priority)
					fmt.Printf("Match: %v\n", rule.Match)
					if rule.If != "" {
						fmt.Printf("Condition: %s\n", rule.If)
					}
					if rule.Note != "" {
						fmt.Printf("Note: %s\n", rule.Note)
					}
					fmt.Println("---")
				}
			}
		}
	},
}
