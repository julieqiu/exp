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
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete librarian.yaml configuration file.
type Config struct {
	// Version is the version of librarian that created this config.
	Version string `yaml:"version"`

	// Language is the primary language for this repository (go, python, rust).
	Language string `yaml:"language"`

	// Container contains the container image configuration.
	Container *Container `yaml:"container,omitempty"`

	// Sources contains references to external source repositories.
	Sources Sources `yaml:"sources,omitempty"`

	// Generate contains generation configuration.
	Generate *Generate `yaml:"generate,omitempty"`

	// Release contains release configuration.
	Release *Release `yaml:"release,omitempty"`

	// Editions is the list of editions in this repository.
	Editions []Edition `yaml:"editions,omitempty"`
}

// Sources contains references to external source repositories.
type Sources struct {
	// Googleapis is the googleapis source repository.
	Googleapis *Source `yaml:"googleapis,omitempty"`

	// Discovery is the discovery-artifact-manager source repository.
	Discovery *Source `yaml:"discovery,omitempty"`
}

// Source represents an external source repository.
type Source struct {
	// URL is the download URL for the source tarball.
	URL string `yaml:"url"`

	// SHA256 is the hash for integrity verification.
	SHA256 string `yaml:"sha256"`
}

// Generate contains generation configuration.
type Generate struct {
	// OutputDir is the directory where generated code is written (relative to repository root).
	OutputDir string `yaml:"output_dir,omitempty"`

	// Defaults contains default values applied to all editions.
	Defaults *GenerateDefaults `yaml:"defaults,omitempty"`
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

// Release contains release configuration.
type Release struct {
	// TagFormat is the template for git tags (e.g., '{id}/v{version}').
	// Supported placeholders: {id}, {name}, {version}
	TagFormat string `yaml:"tag_format,omitempty"`
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

	// Generate contains edition-specific generation configuration.
	// If nil, this is a handwritten edition (release-only).
	Generate *EditionGenerate `yaml:"generate,omitempty"`

	// Go contains Go-specific configuration for this edition.
	Go *GoModule `yaml:"go,omitempty"`
}

// EditionGenerate contains edition-specific generation configuration.
type EditionGenerate struct {
	// APIs is the list of APIs to generate for this edition.
	// Can be specified as simple strings or detailed API configurations.
	APIs []API `yaml:"apis,omitempty"`

	// Keep is a list of file patterns to preserve during regeneration.
	Keep []string `yaml:"keep,omitempty"`

	// Remove is a list of file patterns to delete after generation.
	Remove []string `yaml:"remove,omitempty"`

	// Delete is a list of file paths to delete after generation.
	// These are typically generated files that should be removed.
	Delete []string `yaml:"delete,omitempty"`
}

// GoModule contains Go-specific configuration for an edition.
type GoModule struct {
	// ImportPath is the Go import path for this module.
	// If empty, derived from edition name and repository conventions.
	ImportPath string `yaml:"import_path,omitempty"`

	// ModulePathVersion is the module path version suffix (e.g., "v2", "v3").
	// If empty, derived from the edition Version field.
	ModulePathVersion string `yaml:"module_path_version,omitempty"`
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

	// Go contains Go-specific overrides for this API.
	Go *GoOverrides `yaml:"go,omitempty"`

	// EditionName is the name of the edition this API belongs to.
	// This is populated when the API is accessed through Edition.GetAPIConfig().
	EditionName string `yaml:"-"`
}

// GoOverrides contains Go-specific overrides for an API.
type GoOverrides struct {
	// ProtoPackage is the Go package name for generated proto code.
	// If empty, derived from the API path.
	ProtoPackage string `yaml:"proto_package,omitempty"`

	// ClientDirectory is the directory name for generated client code.
	// If empty, derived from the API version (e.g., "apiv1").
	ClientDirectory string `yaml:"client_directory,omitempty"`

	// DisableGapic disables GAPIC client generation for this API.
	DisableGapic bool `yaml:"disable_gapic,omitempty"`

	// NestedProtos is a list of additional proto files to include in generation.
	NestedProtos []string `yaml:"nested_protos,omitempty"`
}

// GetModulePath returns the full Go module import path.
// If Go.ImportPath is set, returns that value.
// Otherwise, constructs from the edition name assuming cloud.google.com/go/{name}.
func (e *Edition) GetModulePath() string {
	if e.Go != nil && e.Go.ImportPath != "" {
		return e.Go.ImportPath
	}
	// Default: cloud.google.com/go/{name}
	prefix := "cloud.google.com/go/" + e.Name
	version := e.GetModulePathVersion()
	if version != "" {
		return prefix + "/" + version
	}
	return prefix
}

// GetModulePathVersion returns the module path version suffix (e.g., "/v2").
// If Go.ModulePathVersion is set, returns that value.
// Otherwise, derives from the Version field.
func (e *Edition) GetModulePathVersion() string {
	if e.Go != nil && e.Go.ModulePathVersion != "" {
		return e.Go.ModulePathVersion
	}
	// Derive from Version field
	if e.Version == nil || *e.Version == "" {
		return ""
	}
	// TODO(https://github.com/julieqiu/xlibrarian/issues/XXX): Implement getMajorVersion
	// For now, return empty string
	return ""
}

// GetProtoPackage returns the proto package name.
// If Go.ProtoPackage is set, returns that value.
// Otherwise, derives from the API path by replacing "/" with ".".
func (a *API) GetProtoPackage() string {
	if a.Go != nil && a.Go.ProtoPackage != "" {
		return a.Go.ProtoPackage
	}
	// Default: replace "/" with "." in path
	return strings.ReplaceAll(a.Path, "/", ".")
}

// GetAPIConfig returns the configuration for the API identified by its path.
// If no API-specific configuration is found, returns nil.
func (e *Edition) GetAPIConfig(path string) *API {
	if e.Generate == nil {
		return nil
	}
	for i := range e.Generate.APIs {
		if e.Generate.APIs[i].Path == path {
			// Populate EditionName for use by GetClientDirectory
			e.Generate.APIs[i].EditionName = e.Name
			return &e.Generate.APIs[i]
		}
	}
	// Return empty config with EditionName set
	return &API{
		Path:        path,
		EditionName: e.Name,
	}
}

// GetClientDirectory returns the client directory name.
// If Go.ClientDirectory is set, returns that value.
// Otherwise, derives from the API path and edition name.
func (a *API) GetClientDirectory() (string, error) {
	if a.Go != nil && a.Go.ClientDirectory != "" {
		return a.Go.ClientDirectory, nil
	}

	// No override: derive from path
	// google/spanner/v1 => ["google", "spanner", "v1"]
	// google/spanner/admin/instance/v1 => ["google", "spanner", "admin", "instance", "v1"]
	parts := strings.Split(a.Path, "/")
	moduleIndex := slices.Index(parts, a.EditionName)
	if moduleIndex == -1 {
		return "", fmt.Errorf("edition name '%s' not found in API path '%s'", a.EditionName, a.Path)
	}

	// Remove everything up to and include the edition name.
	// google/spanner/v1 => ["v1"]
	// google/spanner/admin/instance/v1 => ["admin", "instance", "v1"]
	parts = parts[moduleIndex+1:]
	parts[len(parts)-1] = "api" + parts[len(parts)-1]
	return strings.Join(parts, "/"), nil
}

// HasDisableGapic returns true if GAPIC generation is disabled for this API.
func (a *API) HasDisableGapic() bool {
	return a.Go != nil && a.Go.DisableGapic
}

// GetNestedProtos returns the list of nested proto files for this API.
func (a *API) GetNestedProtos() []string {
	if a.Go != nil {
		return a.Go.NestedProtos
	}
	return nil
}

// Read reads and parses a librarian.yaml configuration file.
func Read(path string) (*Config, error) {
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

// Write writes the configuration to a file.
func (c *Config) Write(path string) error {
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
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.Language == "" {
		return fmt.Errorf("language is required")
	}

	validLanguages := map[string]bool{
		"go":     true,
		"python": true,
		"rust":   true,
	}
	if !validLanguages[c.Language] {
		return fmt.Errorf("invalid language: %s (must be one of: go, python, rust)", c.Language)
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
