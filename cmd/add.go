package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jufianto/caddyku/internal"
	"github.com/spf13/cobra"
)

var (
	addDomain   string
	addUpstream string
	addConfig   string
	addNoReload bool
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add domain(s) to the Caddyfile",
	Example: `  caddyku add --domain app.com --upstream myapp-backend:8080
  caddyku add --config ~/projects/myapp/caddyku.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cf := caddyfilePath()

		if addConfig != "" {
			cfg, err := internal.ParseConfig(addConfig)
			if err != nil {
				return err
			}
			label := labelFromPath(addConfig)
			entries := toInternalEntries(cfg.Domains)
			if err := internal.AddBlock(cf, label, entries); err != nil {
				return err
			}
			fmt.Printf("Added %d domain(s) from %s (label: %s)\n", len(entries), addConfig, label)
		} else {
			if addDomain == "" || addUpstream == "" {
				return fmt.Errorf("--domain and --upstream are required when not using --config")
			}
			label := sanitizeLabel(addDomain)
			entries := []internal.DomainEntry{{Domain: addDomain, Upstream: addUpstream}}
			if err := internal.AddBlock(cf, label, entries); err != nil {
				return err
			}
			fmt.Printf("Added %s -> %s\n", addDomain, addUpstream)
		}

		if !addNoReload {
			return reloadCaddy()
		}
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addDomain, "domain", "", "domain name (e.g. app.com)")
	addCmd.Flags().StringVar(&addUpstream, "upstream", "", "upstream container:port (e.g. myapp-backend:8080)")
	addCmd.Flags().StringVar(&addConfig, "config", "", "path to caddyku.yaml file")
	addCmd.Flags().BoolVar(&addNoReload, "no-reload", false, "skip caddy reload after adding")
	rootCmd.AddCommand(addCmd)
}

func labelFromPath(path string) string {
	dir := filepath.Dir(path)
	return sanitizeLabel(filepath.Base(dir))
}

func sanitizeLabel(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func toInternalEntries(domains []internal.DomainEntry) []internal.DomainEntry {
	return domains
}
