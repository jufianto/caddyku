package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	proxyDir    string
	projectsDir string
	caddyService string
)

var rootCmd = &cobra.Command{
	Use:   "caddyku",
	Short: "Manage Caddy reverse proxy for multiple Docker projects",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	home, _ := os.UserHomeDir()
	defaultProxyDir := filepath.Join(home, "projects", "caddy-proxy")
	defaultProjectsDir := filepath.Join(home, "projects")

	rootCmd.PersistentFlags().StringVar(&proxyDir, "proxy-dir", defaultProxyDir, "path to caddy-proxy project directory")
	rootCmd.PersistentFlags().StringVar(&projectsDir, "projects-dir", defaultProjectsDir, "root directory to scan for caddyku.yaml files")
	rootCmd.PersistentFlags().StringVar(&caddyService, "caddy-service", "caddy", "caddy service name in docker compose")
}

func caddyfilePath() string {
	return filepath.Join(proxyDir, "Caddyfile")
}

func reloadCaddy() error {
	fmt.Println("Reloading Caddy...")
	cmd := exec.Command("docker", "compose", "exec", caddyService, "caddy", "reload", "--config", "/etc/caddy/Caddyfile")
	cmd.Dir = proxyDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("caddy reload failed: %w", err)
	}
	fmt.Println("Caddy reloaded successfully.")
	return nil
}
