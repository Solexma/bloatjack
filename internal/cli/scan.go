package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan containers and generate optimization report",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scanning containers...")
		// TODO: Implement scan logic
	},
}
