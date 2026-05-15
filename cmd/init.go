package cmd

import (
	"fmt"
	"os"

	"github.com/jufianto/caddyku/internal"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the caddy-proxy project (one-time VPS setup)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(proxyDir, 0755); err != nil {
			return fmt.Errorf("creating proxy dir: %w", err)
		}

		if err := internal.WriteProxyCompose(proxyDir); err != nil {
			return err
		}
		fmt.Println("Created docker-compose.yml")

		if err := internal.WriteEmptyCaddyfile(proxyDir); err != nil {
			return err
		}
		fmt.Println("Created Caddyfile")

		fmt.Printf("\nCaddy proxy initialized at %s\n", proxyDir)
		fmt.Println("Start it with:")
		fmt.Printf("  cd %s && docker compose up -d\n", proxyDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
