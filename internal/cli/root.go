// Package cli defines the command-line interface for the application.
package cli

import (
	"fmt"

	"github.com/Solexma/bloatjack/internal/rules"
	"github.com/spf13/cobra"
)

// Version is set by the main package
var Version string

var rootCmd = &cobra.Command{
	Use:   "bloatjack",
	Short: "BloatJack - Cyber-surgeon that slims your containers",
	Long: `BloatJack is a cyber-surgeon that slims your containers by measuring, explaining, 
and automatically right-sizing resources.`,
	// Version will be set by main
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Set the version string for the root command including the ruleset version
	rulesetVersion, err := rules.GetRulesetVersion()
	if err != nil {
		// If there's an error reading the ruleset version, just use the binary version
		rootCmd.Version = Version
	} else {
		// Version format: "bloatjack version v0.3.4 (ruleset 2025-06-01)"
		rootCmd.Version = fmt.Sprintf("%s (ruleset %s)", Version, rulesetVersion)
	}

	return rootCmd.Execute()
}

func init() {
	// Add other commands here
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(tuneCmd)
	// rootCmd.AddCommand(uiCmd) // Assuming uiCmd will be defined in ui.go
	rootCmd.AddCommand(rulesCmd)
}
