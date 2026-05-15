package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DomainEntry struct {
	Domain   string `yaml:"domain"`
	Upstream string `yaml:"upstream"`
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
	return &cfg, nil
}

func WriteAppConfig(path string, cfg *AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
