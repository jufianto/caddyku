package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jufi/caddyku/internal"
	"github.com/spf13/cobra"
)

var (
	initAppDir       string
	initAppService   string
	initAppContainer string
	initAppDomain    string
	initAppUpstream  string
	initAppNoReload  bool
)

var initAppCmd = &cobra.Command{
	Use:   "init-app",
	Short: "Patch an app's docker-compose.yml to join caddy-net and generate caddyku.yaml",
	Example: `  caddyku init-app --dir ~/projects/myapp --service backend \
    --container myapp-backend --domain myapp.com --upstream myapp-backend:8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if initAppService == "" || initAppContainer == "" {
			return fmt.Errorf("--service and --container are required")
		}

		composePath := filepath.Join(initAppDir, "docker-compose.yml")
		if _, err := os.Stat(composePath); os.IsNotExist(err) {
			return fmt.Errorf("docker-compose.yml not found at %s", composePath)
		}

		if err := internal.PatchAppCompose(composePath, initAppService, initAppContainer); err != nil {
			return fmt.Errorf("patching docker-compose.yml: %w", err)
		}
		fmt.Printf("Patched %s: service %q now joins caddy-net with container_name %q\n",
			composePath, initAppService, initAppContainer)

		if initAppDomain != "" && initAppUpstream != "" {
			cfg := &internal.AppConfig{
				Domains: []internal.DomainEntry{
					{Domain: initAppDomain, Upstream: initAppUpstream},
				},
			}
			configPath := filepath.Join(initAppDir, "caddyku.yaml")
			if err := internal.WriteAppConfig(configPath, cfg); err != nil {
				return fmt.Errorf("writing caddyku.yaml: %w", err)
			}
			fmt.Printf("Created %s\n", configPath)

			label := filepath.Base(initAppDir)
			cf := caddyfilePath()
			if err := internal.AddBlock(cf, label, cfg.Domains); err != nil {
				return fmt.Errorf("updating Caddyfile: %w", err)
			}
			fmt.Printf("Added domain %s -> %s to Caddyfile\n", initAppDomain, initAppUpstream)

			if !initAppNoReload {
				return reloadCaddy()
			}
		}

		return nil
	},
}

func init() {
	home, _ := os.UserHomeDir()
	initAppCmd.Flags().StringVar(&initAppDir, "dir", ".", "path to the app project directory")
	initAppCmd.Flags().StringVar(&initAppService, "service", "", "service name in docker-compose.yml to patch")
	initAppCmd.Flags().StringVar(&initAppContainer, "container", "", "container_name to set for the service")
	initAppCmd.Flags().StringVar(&initAppDomain, "domain", "", "domain to register in Caddyfile (optional)")
	initAppCmd.Flags().StringVar(&initAppUpstream, "upstream", "", "upstream container:port (optional, required if --domain set)")
	initAppCmd.Flags().BoolVar(&initAppNoReload, "no-reload", false, "skip caddy reload")
	_ = home
	rootCmd.AddCommand(initAppCmd)
}
