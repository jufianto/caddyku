package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jufi/caddyku/internal"
	"github.com/spf13/cobra"
)

var syncNoReload bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Scan projects directory for caddyku.yaml files and rebuild Caddyfile",
	Example: `  caddyku sync
  caddyku sync --projects-dir ~/projects`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Scanning %s for caddyku.yaml files...\n", projectsDir)

		entries, err := os.ReadDir(projectsDir)
		if err != nil {
			return fmt.Errorf("reading projects dir: %w", err)
		}

		blocks := map[string][]internal.DomainEntry{}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			configPath := filepath.Join(projectsDir, entry.Name(), "caddyku.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				continue
			}
			cfg, err := internal.ParseConfig(configPath)
			if err != nil {
				fmt.Printf("  warning: skipping %s: %v\n", configPath, err)
				continue
			}
			label := entry.Name()
			blocks[label] = cfg.Domains
			fmt.Printf("  found: %s (%d domain(s))\n", configPath, len(cfg.Domains))
		}

		if len(blocks) == 0 {
			fmt.Println("No caddyku.yaml files found.")
			return nil
		}

		cf := caddyfilePath()
		if err := internal.SyncBlocks(cf, blocks); err != nil {
			return err
		}
		fmt.Printf("Synced %d project(s) into Caddyfile\n", len(blocks))

		if !syncNoReload {
			return reloadCaddy()
		}
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncNoReload, "no-reload", false, "skip caddy reload after sync")
	rootCmd.AddCommand(syncCmd)
}
