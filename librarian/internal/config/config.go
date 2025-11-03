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
	Version  string `yaml:"version"`
	Language string `yaml:"language,omitempty"`
}

type GenerateConfig struct {
	Container      ContainerConfig `yaml:"container,omitempty"`
	GoogleapisRepo string          `yaml:"googleapis_repo,omitempty"`
	GoogleapisRef  string          `yaml:"googleapis_ref,omitempty"`
	DiscoveryRepo  string          `yaml:"discovery_repo,omitempty"`
	DiscoveryRef   string          `yaml:"discovery_ref,omitempty"`
	Dir            string          `yaml:"dir,omitempty"`
	Custom         []map[string]interface{} `yaml:"custom,omitempty"`
}

type ContainerConfig struct {
	Image string `yaml:"image,omitempty"`
	Tag   string `yaml:"tag,omitempty"`
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
	case "language":
		c.Librarian.Language = value
	case "release.tag_format":
		c.Release.TagFormat = value
	case "generate.container.image":
		c.Generate.Container.Image = value
	case "generate.container.tag":
		c.Generate.Container.Tag = value
	case "generate.googleapis_repo":
		c.Generate.GoogleapisRepo = value
	case "generate.googleapis_ref":
		c.Generate.GoogleapisRef = value
	case "generate.discovery_repo":
		c.Generate.DiscoveryRepo = value
	case "generate.discovery_ref":
		c.Generate.DiscoveryRef = value
	case "generate.dir":
		c.Generate.Dir = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// Get retrieves a configuration value.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "version":
		return c.Librarian.Version, nil
	case "language":
		return c.Librarian.Language, nil
	case "release.tag_format":
		return c.Release.TagFormat, nil
	case "generate.container.image":
		return c.Generate.Container.Image, nil
	case "generate.container.tag":
		return c.Generate.Container.Tag, nil
	case "generate.googleapis_repo":
		return c.Generate.GoogleapisRepo, nil
	case "generate.googleapis_ref":
		return c.Generate.GoogleapisRef, nil
	case "generate.discovery_repo":
		return c.Generate.DiscoveryRepo, nil
	case "generate.discovery_ref":
		return c.Generate.DiscoveryRef, nil
	case "generate.dir":
		return c.Generate.Dir, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// GoogleapisURL returns the full URL for the googleapis archive.
func (c *Config) GoogleapisURL() string {
	return fmt.Sprintf("https://%s/archive/%s.tar.gz", c.Generate.GoogleapisRepo, c.Generate.GoogleapisRef)
}

// DiscoveryURL returns the full URL for the discovery archive.
func (c *Config) DiscoveryURL() string {
	return fmt.Sprintf("https://%s/archive/%s.tar.gz", c.Generate.DiscoveryRepo, c.Generate.DiscoveryRef)
}

// GeneratorImage returns the full generator image.
func (c *Config) GeneratorImage() string {
	return fmt.Sprintf("%s:%s", c.Generate.Container.Image, c.Generate.Container.Tag)
}

// Dir returns the generation directory with a default of "generated".
func (c *Config) Dir() string {
	if c.Generate.Dir == "" {
		return "generated"
	}
	return c.Generate.Dir
}
