package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the .librarian/config.yaml structure.
type Config struct {
	Librarian    LibrarianConfig `yaml:"librarian"`
	Sources      SourceConfig    `yaml:"sources"`
	Container    ContainerConfig `yaml:"container"`
}

type LibrarianConfig struct {
	Version          string `yaml:"version"`
	Language         string `yaml:"language"`
	ReleaseTagFormat string `yaml:"release_tag_format"`
}

type SourceConfig struct {
	Googleapis string `yaml:"googleapis"`
	Discovery  string `yaml:"discovery"`
	Protobuf   string `yaml:"protobuf"`
}

type ContainerConfig struct {
	URL     string `yaml:"url"`
	Version string `yaml:"version"`
}

const (
	configDir  = ".librarian"
	configFile = "config.yaml"
)

// Load reads the config.yaml file from the .librarian directory.
func Load() (*Config, error) {
	path := filepath.Join(configDir, configFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to .librarian/config.yaml.
func (c *Config) Save() error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create .librarian directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := filepath.Join(configDir, configFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Set updates a configuration value.
func (c *Config) Set(key, value string) error {
	switch key {
	case "librarian.version":
		c.Librarian.Version = value
	case "librarian.language":
		c.Librarian.Language = value
	case "librarian.release_tag_format":
		c.Librarian.ReleaseTagFormat = value
	case "sources.googleapis":
		c.Sources.Googleapis = value
	case "sources.discovery":
		c.Sources.Discovery = value
	case "sources.protobuf":
		c.Sources.Protobuf = value
	case "container.url":
		c.Container.URL = value
	case "container.version":
		c.Container.Version = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}
