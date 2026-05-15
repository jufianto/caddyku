package cmd

import (
	"fmt"

	"github.com/jufianto/caddyku/internal"
	"github.com/spf13/cobra"
)

var (
	removeDomain   string
	removeNoReload bool
)

var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove a domain block from the Caddyfile",
	Example: `  caddyku remove --domain app.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if removeDomain == "" {
			return fmt.Errorf("--domain is required")
		}

		label := sanitizeLabel(removeDomain)
		cf := caddyfilePath()

		if err := internal.RemoveBlock(cf, label); err != nil {
			return err
		}
		fmt.Printf("Removed domain block for %s\n", removeDomain)

		if !removeNoReload {
			return reloadCaddy()
		}
		return nil
	},
}

func init() {
	removeCmd.Flags().StringVar(&removeDomain, "domain", "", "domain to remove")
	removeCmd.Flags().BoolVar(&removeNoReload, "no-reload", false, "skip caddy reload after removing")
	rootCmd.AddCommand(removeCmd)
}
