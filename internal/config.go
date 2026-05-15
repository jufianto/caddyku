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
	Compose   string        `yaml:"compose"`
	Service   string        `yaml:"service"`
	Container string        `yaml:"container"`
	Domains   []DomainEntry `yaml:"domains"`
}

func ParseConfig(path string) (*AppConfig, error) {
	cfg, err := ReadAppConfig(path)
	if err != nil {
		return nil, err
	}
	if err := ValidateDomains(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func ReadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func ValidateDomains(cfg *AppConfig) error {
	if len(cfg.Domains) == 0 {
		return fmt.Errorf("config has no domains defined")
	}
	for _, domain := range cfg.Domains {
		if domain.Domain == "" {
			return fmt.Errorf("domain entry is missing domain")
		}
		if domain.Upstream == "" && domain.Body == "" {
			return fmt.Errorf("domain %q must define upstream or body", domain.Domain)
		}
		if domain.Upstream != "" && domain.Body != "" {
			return fmt.Errorf("domain %q cannot define both upstream and body", domain.Domain)
		}
	}
	return nil
}

func WriteAppConfig(path string, cfg *AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
