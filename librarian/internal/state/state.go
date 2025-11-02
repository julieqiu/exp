package state

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// State represents the .librarian/state.yaml structure.
type State struct {
	Libraries map[string]*Library `yaml:"libraries"`
}

// Library represents a single library in the state file.
type Library struct {
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
			return &State{Libraries: make(map[string]*Library)}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if s.Libraries == nil {
		s.Libraries = make(map[string]*Library)
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

// AddLibrary adds or updates a library in the state.
func (s *State) AddLibrary(id string, lib *Library) {
	if s.Libraries == nil {
		s.Libraries = make(map[string]*Library)
	}
	s.Libraries[id] = lib
}

// RemoveLibrary removes a library from the state.
func (s *State) RemoveLibrary(id string) error {
	if _, exists := s.Libraries[id]; !exists {
		return fmt.Errorf("library %s not found", id)
	}
	delete(s.Libraries, id)
	return nil
}

// GetLibrary retrieves a library from the state.
func (s *State) GetLibrary(id string) (*Library, bool) {
	lib, ok := s.Libraries[id]
	return lib, ok
}
