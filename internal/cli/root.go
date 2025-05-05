// Package cli defines the command-line interface for the application.
package cli

import (
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
	// Set the version string for the root command
	rootCmd.Version = Version
	return rootCmd.Execute()
}

func init() {
	// Add other commands here
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(tuneCmd)
	// rootCmd.AddCommand(uiCmd) // Assuming uiCmd will be defined in ui.go
	rootCmd.AddCommand(rulesCmd)
}
