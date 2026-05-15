package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jufianto/caddyku/internal"
	"github.com/spf13/cobra"
)

var (
	initAppDir            string
	initAppService        string
	initAppContainer      string
	initAppCompose        string
	initAppDomain         string
	initAppUpstream       string
	initAppNoReload       bool
	initAppForceContainer bool
)

var initAppCmd = &cobra.Command{
	Use:   "init-app",
	Short: "Patch an app's Docker Compose file to join caddy-net and generate caddyku.yaml",
	Example: `  caddyku init-app --dir ~/projects/myapp
  caddyku init-app --dir ~/projects/myapp --service backend --container myapp-backend
  caddyku init-app --dir ~/projects/myapp --compose-file docker-compose.prod.yml --service backend
  caddyku init-app --dir ~/projects/myapp --service backend --domain myapp.com --upstream myapp-backend:8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(initAppDir)
		if err != nil {
			return fmt.Errorf("resolving dir: %w", err)
		}

		configPath := filepath.Join(absDir, "caddyku.yaml")
		cfg := &internal.AppConfig{}
		if _, err := os.Stat(configPath); err == nil {
			parsed, err := internal.ReadAppConfig(configPath)
			if err != nil {
				return fmt.Errorf("reading caddyku.yaml: %w", err)
			}
			cfg = parsed
		}

		service := firstNonEmpty(initAppService, cfg.Service)
		containerName := firstNonEmpty(initAppContainer, cfg.Container)
		composeFile := firstNonEmpty(initAppCompose, cfg.Compose, "docker-compose.yml")
		if service == "" {
			return fmt.Errorf("service is required; pass --service or define service: in caddyku.yaml")
		}

		composePath := composeFile
		if !filepath.IsAbs(composePath) {
			composePath = filepath.Join(absDir, composeFile)
		}
		if _, err := os.Stat(composePath); os.IsNotExist(err) {
			return fmt.Errorf("compose file not found at %s", composePath)
		}

		actualContainerName, err := internal.PatchAppCompose(composePath, service, containerName, initAppForceContainer)
		if err != nil {
			return fmt.Errorf("patching docker-compose.yml: %w", err)
		}
		if actualContainerName == "" {
			fmt.Printf("Patched %s: service %q now joins caddy-net\n", composePath, service)
		} else {
			fmt.Printf("Patched %s: service %q now joins caddy-net with container_name %q\n",
				composePath, service, actualContainerName)
		}

		if cfg.Service == "" {
			cfg.Service = service
		}
		if cfg.Compose == "" && composeFile != "docker-compose.yml" {
			cfg.Compose = composeFile
		}
		if cfg.Container == "" && actualContainerName != "" {
			cfg.Container = actualContainerName
		}

		if initAppDomain != "" && initAppUpstream != "" {
			cfg.Domains = []internal.DomainEntry{
				{Domain: initAppDomain, Upstream: initAppUpstream},
			}
		} else if initAppDomain != "" || initAppUpstream != "" {
			return fmt.Errorf("--domain and --upstream must be provided together")
		}

		if len(cfg.Domains) > 0 {
			if err := internal.ValidateDomains(cfg); err != nil {
				return err
			}
			appCfg := &internal.AppConfig{
				Compose:   cfg.Compose,
				Service:   cfg.Service,
				Container: cfg.Container,
				Domains: []internal.DomainEntry{
					cfg.Domains[0],
				},
			}
			if len(cfg.Domains) > 1 {
				appCfg.Domains = cfg.Domains
			}
			if err := internal.WriteAppConfig(configPath, appCfg); err != nil {
				return fmt.Errorf("writing caddyku.yaml: %w", err)
			}
			fmt.Printf("Wrote %s\n", configPath)

			label := filepath.Base(absDir)
			cf := caddyfilePath()
			if err := internal.AddBlock(cf, label, appCfg.Domains); err != nil {
				return fmt.Errorf("updating Caddyfile: %w", err)
			}
			fmt.Printf("Added %d domain(s) to Caddyfile\n", len(appCfg.Domains))

			if !initAppNoReload {
				return reloadCaddy()
			}
		}

		return nil
	},
}

func init() {
	initAppCmd.Flags().StringVar(&initAppDir, "dir", ".", "path to the app project directory")
	initAppCmd.Flags().StringVar(&initAppCompose, "compose-file", "", "compose file to patch, relative to --dir unless absolute")
	initAppCmd.Flags().StringVar(&initAppService, "service", "", "service name in the compose file to patch")
	initAppCmd.Flags().StringVar(&initAppContainer, "container", "", "container_name to set if the service does not already have one")
	initAppCmd.Flags().StringVar(&initAppDomain, "domain", "", "domain to register in Caddyfile (optional)")
	initAppCmd.Flags().StringVar(&initAppUpstream, "upstream", "", "upstream container:port (optional, required if --domain set)")
	initAppCmd.Flags().BoolVar(&initAppNoReload, "no-reload", false, "skip caddy reload")
	initAppCmd.Flags().BoolVar(&initAppForceContainer, "force-container", false, "overwrite existing container_name if different from --container")
	rootCmd.AddCommand(initAppCmd)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
