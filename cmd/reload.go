package cmd

import (
	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload Caddy configuration without downtime",
	RunE: func(cmd *cobra.Command, args []string) error {
		return reloadCaddy()
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}
