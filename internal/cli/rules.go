package cli

import (
	"fmt"

	"github.com/Solexma/bloatjack/internal/rules"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List embedded rules",
	RunE: func(cmd *cobra.Command, _ []string) error {
		rs, err := rules.Parse(rules.EmbeddedRules)
		if err != nil {
			return err
		}

		for _, r := range rs {
			fmt.Printf("• %-20s  prio:%2d  match:%v\n",
				r.ID, r.Priority, r.Match)
			if r.If != "" {
				fmt.Printf("   if:   %s\n", r.If)
			}
			if r.Note != "" {
				fmt.Printf("   note: %s\n", r.Note)
			}
		}
		return nil
	},
}
