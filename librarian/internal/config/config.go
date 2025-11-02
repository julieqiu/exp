package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the .librarian/config.yaml structure.
type Config struct {
	Librarian LibrarianConfig `yaml:"librarian"`
	Generate  GenerateConfig  `yaml:"generate,omitempty"`
	Release   ReleaseConfig   `yaml:"release,omitempty"`
}

type LibrarianConfig struct {
	Version string `yaml:"version"`
	Mode    string `yaml:"mode"`
}

type GenerateConfig struct {
	Image      string                   `yaml:"image,omitempty"`
	Googleapis string                   `yaml:"googleapis,omitempty"`
	Discovery  string                   `yaml:"discovery,omitempty"`
	Custom     []map[string]interface{} `yaml:"custom,omitempty"`
}

type ReleaseConfig struct {
	TagFormat string `yaml:"tag_format"`
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
	case "version":
		c.Librarian.Version = value
	case "mode":
		c.Librarian.Mode = value
	case "release.tag_format":
		c.Release.TagFormat = value
	case "generate.image":
		c.Generate.Image = value
	case "generate.googleapis":
		c.Generate.Googleapis = value
	case "generate.discovery":
		c.Generate.Discovery = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// GoogleapisURL returns the full URL for the googleapis archive.
func (c *Config) GoogleapisURL() string {
	return fmt.Sprintf("https://github.com/googleapis/googleapis/archive/%s.tar.gz", c.Generate.Googleapis)
}

// DiscoveryURL returns the full URL for the discovery archive.
func (c *Config) DiscoveryURL() string {
	return fmt.Sprintf("https://github.com/googleapis/discovery-artifact-manager/archive/%s.tar.gz", c.Generate.Discovery)
}

// GeneratorImage returns the full generator image.
func (c *Config) GeneratorImage() string {
	return c.Generate.Image
}
