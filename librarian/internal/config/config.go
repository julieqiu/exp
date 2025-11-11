package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the .librarian/config.yaml structure.
type Config struct {
	Librarian LibrarianConfig  `yaml:"librarian"`
	Generate  *GenerateConfig  `yaml:"generate,omitempty"`
	Release   *ReleaseConfig   `yaml:"release,omitempty"`
}

type LibrarianConfig struct {
	Version  string `yaml:"version"`
	Language string `yaml:"language,omitempty"`
}

type GenerateConfig struct {
	Container  *ContainerConfig `yaml:"container,omitempty"`
	Googleapis *RepoConfig      `yaml:"googleapis,omitempty"`
	Discovery  *RepoConfig      `yaml:"discovery,omitempty"`
	Dir        string           `yaml:"dir,omitempty"`
}

type ContainerConfig struct {
	Image string `yaml:"image,omitempty"`
	Tag   string `yaml:"tag,omitempty"`
}

type RepoConfig struct {
	Repo string `yaml:"repo"`
	Ref  string `yaml:"ref,omitempty"`
}

type ReleaseConfig struct {
	TagFormat      string          `yaml:"tag_format"`
	BranchPatterns []BranchPattern `yaml:"branch_patterns,omitempty"`
}

type BranchPattern struct {
	Pattern    string `yaml:"pattern"`     // "main", "release/*", etc.
	Prerelease string `yaml:"prerelease"`  // "", "rc", "alpha", etc.
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
	case "release.tag_format":
		if c.Release == nil {
			c.Release = &ReleaseConfig{}
		}
		c.Release.TagFormat = value
	case "generate.container.image":
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		if c.Generate.Container == nil {
			c.Generate.Container = &ContainerConfig{}
		}
		c.Generate.Container.Image = value
	case "generate.container.tag":
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		if c.Generate.Container == nil {
			c.Generate.Container = &ContainerConfig{}
		}
		c.Generate.Container.Tag = value
	case "generate.container":
		// Syntactic sugar: parse "image:tag"
		parts := strings.Split(value, ":")
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		if c.Generate.Container == nil {
			c.Generate.Container = &ContainerConfig{}
		}
		c.Generate.Container.Image = parts[0]
		if len(parts) > 1 {
			c.Generate.Container.Tag = parts[1]
		}
	case "generate.googleapis.repo":
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		if c.Generate.Googleapis == nil {
			c.Generate.Googleapis = &RepoConfig{}
		}
		c.Generate.Googleapis.Repo = value
	case "generate.googleapis.ref":
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		if c.Generate.Googleapis == nil {
			c.Generate.Googleapis = &RepoConfig{}
		}
		c.Generate.Googleapis.Ref = value
	case "generate.discovery.repo":
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		if c.Generate.Discovery == nil {
			c.Generate.Discovery = &RepoConfig{}
		}
		c.Generate.Discovery.Repo = value
	case "generate.discovery.ref":
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		if c.Generate.Discovery == nil {
			c.Generate.Discovery = &RepoConfig{}
		}
		c.Generate.Discovery.Ref = value
	case "generate.dir":
		if c.Generate == nil {
			c.Generate = &GenerateConfig{}
		}
		c.Generate.Dir = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// Get retrieves a configuration value.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "librarian.version":
		return c.Librarian.Version, nil
	case "librarian.language":
		return c.Librarian.Language, nil
	case "release.tag_format":
		if c.Release != nil {
			return c.Release.TagFormat, nil
		}
		return "", nil
	case "generate.container.image":
		if c.Generate != nil && c.Generate.Container != nil {
			return c.Generate.Container.Image, nil
		}
		return "", nil
	case "generate.container.tag":
		if c.Generate != nil && c.Generate.Container != nil {
			return c.Generate.Container.Tag, nil
		}
		return "", nil
	case "generate.googleapis.repo":
		if c.Generate != nil && c.Generate.Googleapis != nil {
			return c.Generate.Googleapis.Repo, nil
		}
		return "", nil
	case "generate.googleapis.ref":
		if c.Generate != nil && c.Generate.Googleapis != nil {
			return c.Generate.Googleapis.Ref, nil
		}
		return "", nil
	case "generate.discovery.repo":
		if c.Generate != nil && c.Generate.Discovery != nil {
			return c.Generate.Discovery.Repo, nil
		}
		return "", nil
	case "generate.discovery.ref":
		if c.Generate != nil && c.Generate.Discovery != nil {
			return c.Generate.Discovery.Ref, nil
		}
		return "", nil
	case "generate.dir":
		if c.Generate != nil {
			return c.Generate.Dir, nil
		}
		return "", nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// GoogleapisURL returns the full URL for the googleapis archive.
func (c *Config) GoogleapisURL() string {
	if c.Generate == nil || c.Generate.Googleapis == nil {
		return ""
	}
	return fmt.Sprintf("https://%s/archive/%s.tar.gz", c.Generate.Googleapis.Repo, c.Generate.Googleapis.Ref)
}

// DiscoveryURL returns the full URL for the discovery archive.
func (c *Config) DiscoveryURL() string {
	if c.Generate == nil || c.Generate.Discovery == nil {
		return ""
	}
	return fmt.Sprintf("https://%s/archive/%s.tar.gz", c.Generate.Discovery.Repo, c.Generate.Discovery.Ref)
}

// ContainerImage returns the full container image with tag.
func (c *Config) ContainerImage() string {
	if c.Generate == nil || c.Generate.Container == nil {
		return ""
	}
	if c.Generate.Container.Tag != "" {
		return fmt.Sprintf("%s:%s", c.Generate.Container.Image, c.Generate.Container.Tag)
	}
	return c.Generate.Container.Image
}

// Dir returns the generation directory with a default of "generated".
func (c *Config) Dir() string {
	if c.Generate != nil && c.Generate.Dir != "" {
		return c.Generate.Dir
	}
	return "generated"
}
