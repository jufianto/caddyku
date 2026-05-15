package internal

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const proxyComposeTemplate = `services:
  caddy:
    image: caddy:latest
    container_name: caddy-proxy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    networks:
      - caddy-net

networks:
  caddy-net:
    name: caddy-net

volumes:
  caddy_data:
  caddy_config:
`

func WriteProxyCompose(dir string) error {
	path := dir + "/docker-compose.yml"
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("docker-compose.yml already exists at %s", path)
	}
	return os.WriteFile(path, []byte(proxyComposeTemplate), 0644)
}

func WriteEmptyCaddyfile(dir string) error {
	path := dir + "/Caddyfile"
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("Caddyfile already exists at %s", path)
	}
	return os.WriteFile(path, []byte(""), 0644)
}

func PatchAppCompose(path, service, containerName string, forceContainer bool) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}

	services, ok := raw["services"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no services found in %s", path)
	}

	svc, ok := services[service].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("service %q not found in %s", service, path)
	}

	actualContainerName := containerName
	if existing, ok := svc["container_name"].(string); ok && existing != "" {
		actualContainerName = existing
		if containerName != "" && containerName != existing && !forceContainer {
			return "", fmt.Errorf("service %q already has container_name %q; use --force-container to overwrite it", service, existing)
		}
	}

	if containerName != "" && (actualContainerName == containerName || forceContainer) {
		actualContainerName = containerName
		svc["container_name"] = containerName
	}

	networks := toStringSlice(svc["networks"])
	if !contains(networks, "caddy-net") {
		networks = append(networks, "caddy-net")
	}
	if !contains(networks, "default") {
		networks = append(networks, "default")
	}
	svc["networks"] = networks
	services[service] = svc

	topNetworks, _ := raw["networks"].(map[string]interface{})
	if topNetworks == nil {
		topNetworks = map[string]interface{}{}
	}
	topNetworks["caddy-net"] = map[string]interface{}{
		"external": true,
		"name":     "caddy-net",
	}
	raw["networks"] = topNetworks

	out, err := yaml.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("marshaling compose: %w", err)
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		return "", err
	}
	return actualContainerName, nil
}

func toStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []interface{}:
		var result []string
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return val
	}
	return nil
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}
