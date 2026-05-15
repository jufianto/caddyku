package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DomainEntry struct {
	Domain   string `yaml:"domain"`
	Upstream string `yaml:"upstream"`
	Body     string `yaml:"body"`
}

type AppConfig struct {
	Domains []DomainEntry `yaml:"domains"`
}

func ParseConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if len(cfg.Domains) == 0 {
		return nil, fmt.Errorf("config has no domains defined")
	}
	for _, domain := range cfg.Domains {
		if domain.Domain == "" {
			return nil, fmt.Errorf("domain entry is missing domain")
		}
		if domain.Upstream == "" && domain.Body == "" {
			return nil, fmt.Errorf("domain %q must define upstream or body", domain.Domain)
		}
		if domain.Upstream != "" && domain.Body != "" {
			return nil, fmt.Errorf("domain %q cannot define both upstream and body", domain.Domain)
		}
	}
	return &cfg, nil
}

func WriteAppConfig(path string, cfg *AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
