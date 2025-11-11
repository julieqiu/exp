// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete librarian.yaml configuration file.
type Config struct {
	Librarian Librarian `yaml:"librarian"`
	Sources   Sources   `yaml:"sources,omitempty"`
	Generate  Generate  `yaml:"generate,omitempty"`
	Release   Release   `yaml:"release,omitempty"`
	Editions  []Edition `yaml:"editions,omitempty"`
}

// Librarian contains the core metadata about the librarian configuration.
type Librarian struct {
	// Version is the version of librarian that created this config.
	Version string `yaml:"version"`
	// Language is the primary language for this repository (go, python, rust, dart).
	Language string `yaml:"language"`
}

// Sources contains references to external source repositories.
type Sources struct {
	Googleapis *Source `yaml:"googleapis,omitempty"`
	Discovery  *Source `yaml:"discovery,omitempty"`
}

// Source represents an external source repository.
type Source struct {
	// URL is the download URL for the source tarball.
	URL string `yaml:"url"`
	// SHA256 is the hash for integrity verification.
	SHA256 string `yaml:"sha256"`
}

// Generate contains repository-wide generation configuration.
type Generate struct {
	Container  *Container       `yaml:"container,omitempty"`
	OutputDir  string           `yaml:"output_dir,omitempty"`
	Defaults   *GenerateDefaults `yaml:"defaults,omitempty"`
}

// Container contains container image configuration.
type Container struct {
	// Image is the container registry path (without tag).
	Image string `yaml:"image"`
	// Tag is the container image tag.
	Tag string `yaml:"tag"`
}

// GenerateDefaults contains default values applied to all editions.
type GenerateDefaults struct {
	// Transport is the default transport protocol (e.g., grpc+rest).
	Transport string `yaml:"transport,omitempty"`
	// RestNumericEnums is the default for using numeric enums in REST.
	RestNumericEnums *bool `yaml:"rest_numeric_enums,omitempty"`
	// ReleaseLevel is the default release level (stable, preview).
	ReleaseLevel string `yaml:"release_level,omitempty"`
}

// Release contains repository-wide release configuration.
type Release struct {
	// TagFormat is the template for git tags (e.g., '{id}/v{version}').
	// Supported placeholders: {id}, {name}, {version}
	TagFormat string `yaml:"tag_format"`
}

// Edition represents a single edition (library, package, artifact).
type Edition struct {
	// Name is the name of the edition.
	Name string `yaml:"name"`
	// Path is the directory path relative to repository root.
	// If empty, derived from name and generate.output_dir.
	Path string `yaml:"path,omitempty"`
	// Version is the current version (null if never released).
	// Use pointer to distinguish between "0.0.0" and null.
	Version *string `yaml:"version,omitempty"`
	// APIs is a short-form list of API paths for this edition.
	// Each string is an API path like "google/cloud/secretmanager/v1".
	APIs []string `yaml:"apis,omitempty"`
	// Generate contains edition-specific generation configuration.
	Generate *EditionGenerate `yaml:"generate,omitempty"`
}

// EditionGenerate contains edition-specific generation configuration.
type EditionGenerate struct {
	// APIs is the detailed list of API configurations.
	// This is used when you need to override defaults or specify additional fields.
	APIs []API `yaml:"apis,omitempty"`
	// Metadata contains human-readable metadata about the edition.
	Metadata *Metadata `yaml:"metadata,omitempty"`
	// Language contains language-specific configuration.
	Language *LanguageConfig `yaml:"language,omitempty"`
	// Keep is a list of file patterns to preserve during regeneration.
	Keep []string `yaml:"keep,omitempty"`
	// Remove is a list of file patterns to delete after generation.
	Remove []string `yaml:"remove,omitempty"`
	// ReleaseLevel overrides the default release level for this edition.
	ReleaseLevel string `yaml:"release_level,omitempty"`
	// Transport overrides the default transport for this edition.
	Transport string `yaml:"transport,omitempty"`
}

// API represents a single API configuration.
type API struct {
	// Path is the API path within googleapis (e.g., google/cloud/secretmanager/v1).
	Path string `yaml:"path"`
	// GRPCServiceConfig is the path to the gRPC service config file.
	GRPCServiceConfig string `yaml:"grpc_service_config,omitempty"`
	// ServiceYAML is the path to the service YAML file.
	ServiceYAML string `yaml:"service_yaml,omitempty"`
	// Transport is the transport protocol (grpc, grpc+rest).
	Transport string `yaml:"transport,omitempty"`
	// RestNumericEnums indicates whether to use numeric enums in REST.
	RestNumericEnums *bool `yaml:"rest_numeric_enums,omitempty"`
	// OptArgs are additional generator-specific options.
	OptArgs []string `yaml:"opt_args,omitempty"`
}

// Metadata contains human-readable metadata about an edition.
type Metadata struct {
	// NamePretty is the human-readable name (e.g., "Secret Manager").
	NamePretty string `yaml:"name_pretty,omitempty"`
	// ProductDocumentation is the URL to product documentation.
	ProductDocumentation string `yaml:"product_documentation,omitempty"`
	// ClientDocumentation is the URL to client library documentation.
	ClientDocumentation string `yaml:"client_documentation,omitempty"`
	// IssueTracker is the URL to the issue tracker.
	IssueTracker string `yaml:"issue_tracker,omitempty"`
	// ReleaseLevel is the release level (stable, preview).
	ReleaseLevel string `yaml:"release_level,omitempty"`
	// LibraryType is the library type (GAPIC_AUTO, GAPIC_COMBO).
	LibraryType string `yaml:"library_type,omitempty"`
	// APIID is the API identifier (e.g., secretmanager.googleapis.com).
	APIID string `yaml:"api_id,omitempty"`
	// APIShortname is the short API name (e.g., secretmanager).
	APIShortname string `yaml:"api_shortname,omitempty"`
	// APIDescription is a description of the API.
	APIDescription string `yaml:"api_description,omitempty"`
	// DefaultVersion is the default API version (e.g., v1).
	DefaultVersion string `yaml:"default_version,omitempty"`
}

// LanguageConfig contains language-specific configuration.
// Only one language field should be set based on the repository's language.
type LanguageConfig struct {
	Go     *GoConfig     `yaml:"go,omitempty"`
	Python *PythonConfig `yaml:"python,omitempty"`
	Rust   *RustConfig   `yaml:"rust,omitempty"`
	Dart   *DartConfig   `yaml:"dart,omitempty"`
}

// GoConfig contains Go-specific configuration.
type GoConfig struct {
	// Module is the Go module path (e.g., cloud.google.com/go/secretmanager).
	Module string `yaml:"module,omitempty"`
}

// PythonConfig contains Python-specific configuration.
type PythonConfig struct {
	// Package is the Python package name (e.g., google-cloud-secret-manager).
	Package string `yaml:"package,omitempty"`
}

// RustConfig contains Rust-specific configuration.
type RustConfig struct {
	// Crate is the Rust crate name (e.g., cloud-storage-v1).
	Crate string `yaml:"crate,omitempty"`
}

// DartConfig contains Dart-specific configuration.
type DartConfig struct {
	// Package is the Dart package name.
	Package string `yaml:"package,omitempty"`
}

// Load reads and parses a librarian.yaml configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to a file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetEdition returns the edition with the given name, or nil if not found.
func (c *Config) GetEdition(name string) *Edition {
	for i := range c.Editions {
		if c.Editions[i].Name == name {
			return &c.Editions[i]
		}
	}
	return nil
}

// Validate performs basic validation on the configuration.
func (c *Config) Validate() error {
	if c.Librarian.Version == "" {
		return fmt.Errorf("librarian.version is required")
	}
	if c.Librarian.Language == "" {
		return fmt.Errorf("librarian.language is required")
	}

	validLanguages := map[string]bool{
		"go":     true,
		"python": true,
		"rust":   true,
		"dart":   true,
	}
	if !validLanguages[c.Librarian.Language] {
		return fmt.Errorf("invalid language: %s (must be one of: go, python, rust, dart)", c.Librarian.Language)
	}

	// Validate edition names are unique
	names := make(map[string]bool)
	for i, edition := range c.Editions {
		if edition.Name == "" {
			return fmt.Errorf("edition at index %d has empty name", i)
		}
		if names[edition.Name] {
			return fmt.Errorf("duplicate edition name: %s", edition.Name)
		}
		names[edition.Name] = true
	}

	return nil
}
