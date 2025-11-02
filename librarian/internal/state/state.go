package state

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// State represents the .librarian/state.yaml structure.
type State struct {
	Artifacts map[string]*Artifact `yaml:"artifacts"`
}

// Artifact represents a single artifact in the state file.
type Artifact struct {
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
			return &State{Artifacts: make(map[string]*Artifact)}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if s.Artifacts == nil {
		s.Artifacts = make(map[string]*Artifact)
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

// AddArtifact adds or updates an artifact in the state.
func (s *State) AddArtifact(id string, artifact *Artifact) {
	if s.Artifacts == nil {
		s.Artifacts = make(map[string]*Artifact)
	}
	s.Artifacts[id] = artifact
}

// RemoveArtifact removes an artifact from the state.
func (s *State) RemoveArtifact(id string) error {
	if _, exists := s.Artifacts[id]; !exists {
		return fmt.Errorf("artifact %s not found", id)
	}
	delete(s.Artifacts, id)
	return nil
}

// GetArtifact retrieves an artifact from the state.
func (s *State) GetArtifact(id string) (*Artifact, bool) {
	artifact, ok := s.Artifacts[id]
	return artifact, ok
}
