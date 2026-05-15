package cmd

import (
	"fmt"

	"github.com/jufianto/caddyku/internal"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domains in the current Caddyfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		domains, err := internal.ListDomains(caddyfilePath())
		if err != nil {
			return err
		}
		if len(domains) == 0 {
			fmt.Println("No domains found in Caddyfile.")
			return nil
		}
		fmt.Printf("Domains in %s:\n", caddyfilePath())
		for _, d := range domains {
			fmt.Printf("  %s\n", d)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
