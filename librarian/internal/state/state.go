package state

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const stateFile = ".librarian.yaml"

// Artifact represents a single artifact's state.
type Artifact struct {
	Generate *GenerateState `yaml:"generate,omitempty"`
	Release  *ReleaseState  `yaml:"release,omitempty"`
	Config   *ConfigState   `yaml:"config,omitempty"`
	Language *LanguageState `yaml:"language,omitempty"`
}

// GenerateState tracks generation metadata.
type GenerateState struct {
	APIs       []API            `yaml:"apis"`
	Commit     string           `yaml:"commit"`
	Librarian  string           `yaml:"librarian"`
	Container  ContainerState   `yaml:"container"`
	Googleapis GoogleapisState  `yaml:"googleapis"`
	Discovery  DiscoveryState   `yaml:"discovery"`
}

// ContainerState tracks container metadata.
type ContainerState struct {
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
}

// GoogleapisState tracks googleapis metadata.
type GoogleapisState struct {
	Repo string `yaml:"repo"`
	Ref  string `yaml:"ref"`
}

// DiscoveryState tracks discovery metadata.
type DiscoveryState struct {
	Repo string `yaml:"repo"`
	Ref  string `yaml:"ref"`
}

// ReleaseState tracks release metadata.
type ReleaseState struct {
	Version      string       `yaml:"version"`
	Prepared     *ReleaseInfo `yaml:"prepared,omitempty"`
}

// ReleaseInfo contains information about a specific release.
type ReleaseInfo struct {
	Tag    string `yaml:"tag,omitempty"`
	Commit string `yaml:"commit,omitempty"`
}

// API represents an API path.
type API struct {
	Path          string `yaml:"path"`
	ServiceConfig string `yaml:"service_config,omitempty"`
}

// ConfigState holds artifact-specific configuration.
type ConfigState struct {
	Keep    []string `yaml:"keep,omitempty"`    // Files/directories to keep (don't overwrite) during generation
	Remove  []string `yaml:"remove,omitempty"`  // Files to remove after generation
	Exclude []string `yaml:"exclude,omitempty"` // Files to exclude from release
	Dir     string   `yaml:"dir,omitempty"`     // Where to write generated code (overrides global default)
}

// LanguageState holds language-specific metadata for the artifact.
type LanguageState struct {
	Go     *GoLanguage     `yaml:"go,omitempty"`
	Python *PythonLanguage `yaml:"python,omitempty"`
	Rust   *RustLanguage   `yaml:"rust,omitempty"`
	Dart   *DartLanguage   `yaml:"dart,omitempty"`
}

// GoLanguage holds Go-specific metadata.
type GoLanguage struct {
	Module string `yaml:"module,omitempty"` // Go module path (e.g., "github.com/user/repo")
}

// PythonLanguage holds Python-specific metadata.
type PythonLanguage struct {
	Package string `yaml:"package,omitempty"` // Python package name (e.g., "my-package")
}

// RustLanguage holds Rust-specific metadata.
type RustLanguage struct {
	Crate string `yaml:"crate,omitempty"` // Rust crate name (e.g., "my_crate")
}

// DartLanguage holds Dart-specific metadata.
type DartLanguage struct {
	Package string `yaml:"package,omitempty"` // Dart package name (e.g., "my_package")
}

// Load reads the .librarian.yaml file from the artifact's directory.
func Load(artifactPath string) (*Artifact, error) {
	path := filepath.Join(artifactPath, stateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Artifact{}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var a Artifact
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &a, nil
}

// Save writes the artifact state to .librarian.yaml in the artifact's directory.
func (a *Artifact) Save(artifactPath string) error {
	data, err := yaml.Marshal(a)
	if err != nil {
		return fmt.Errorf("failed to marshal artifact state: %w", err)
	}

	path := filepath.Join(artifactPath, stateFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// Remove deletes the .librarian.yaml file from the artifact's directory.
func Remove(artifactPath string) error {
	path := filepath.Join(artifactPath, stateFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}
	return nil
}

// LoadAll scans for all .librarian.yaml files and returns a map of artifact paths to their states.
func LoadAll() (map[string]*Artifact, error) {
	artifacts := make(map[string]*Artifact)

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == stateFile {
			// Get the directory containing the .librarian.yaml file
			artifactPath := filepath.Dir(path)

			artifact, err := Load(artifactPath)
			if err != nil {
				return fmt.Errorf("failed to load artifact at %s: %w", artifactPath, err)
			}

			artifacts[artifactPath] = artifact
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan for artifacts: %w", err)
	}

	return artifacts, nil
}
