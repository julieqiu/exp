package state

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// State represents the .librarian/state.yaml structure.
type State struct {
	Packages map[string]*Package `yaml:"packages"`
}

// Package represents a single package in the state file.
type Package struct {
	Path     string          `yaml:"path,omitempty"`
	Generate *GenerateState  `yaml:"generate,omitempty"`
	Release  *ReleaseState   `yaml:"release,omitempty"`
}

// GenerateState tracks generation metadata.
type GenerateState struct {
	APIs          []API  `yaml:"apis"`
	Commit        string `yaml:"commit"`
	Librarian     string `yaml:"librarian"`
	Image         string `yaml:"image"`
	GoogleapisSHA string `yaml:"googleapis-sha,omitempty"`
	DiscoverySHA  string `yaml:"discovery-sha,omitempty"`
}

// ReleaseState tracks release metadata.
type ReleaseState struct {
	LastReleasedAt *ReleaseInfo `yaml:"last_released_at,omitempty"`
	NextReleaseAt  *ReleaseInfo `yaml:"next_release_at,omitempty"`
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

const (
	stateDir  = ".librarian"
	stateFile = "state.yaml"
)

// Load reads the state.yaml file from the .librarian directory.
func Load() (*State, error) {
	path := filepath.Join(stateDir, stateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Packages: make(map[string]*Package)}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if s.Packages == nil {
		s.Packages = make(map[string]*Package)
	}

	return &s, nil
}

// Save writes the state to .librarian/state.yaml.
func (s *State) Save() error {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create .librarian directory: %w", err)
	}

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	path := filepath.Join(stateDir, stateFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// AddPackage adds or updates a package in the state.
func (s *State) AddPackage(id string, pkg *Package) {
	if s.Packages == nil {
		s.Packages = make(map[string]*Package)
	}
	s.Packages[id] = pkg
}

// RemovePackage removes a package from the state.
func (s *State) RemovePackage(id string) error {
	if _, exists := s.Packages[id]; !exists {
		return fmt.Errorf("package %s not found", id)
	}
	delete(s.Packages, id)
	return nil
}

// GetPackage retrieves a package from the state.
func (s *State) GetPackage(id string) (*Package, bool) {
	pkg, ok := s.Packages[id]
	return pkg, ok
}
