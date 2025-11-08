package build

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Config holds the configuration for the build command.
type Config struct {
	LibrarianDir string
	RepoDir      string
}

// Request represents the build-request.json structure.
type Request struct {
	Libraries []Library `json:"libraries"`
}

// Library represents a library to build and test.
type Library struct {
	ID          string   `json:"id"`
	SourceRoots []string `json:"source_roots"`
}

// Build implements the build command.
// It reads build-request.json and simulates building and testing libraries.
func Build(ctx context.Context, cfg *Config) error {
	slog.Info("build: starting", "config", cfg)

	// Validate directories exist
	if err := validateDirs(cfg); err != nil {
		return err
	}

	// Read and parse the request
	requestPath := filepath.Join(cfg.LibrarianDir, "build-request.json")
	slog.Debug("build: reading request", "path", requestPath)

	data, err := os.ReadFile(requestPath)
	if err != nil {
		return fmt.Errorf("build: failed to read request file: %w", err)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("build: failed to parse request JSON: %w", err)
	}

	slog.Info("build: parsed request", "libraries", len(req.Libraries))

	// Process each library
	for _, lib := range req.Libraries {
		slog.Info("build: processing library", "id", lib.ID)

		if err := buildLibrary(cfg.RepoDir, &lib); err != nil {
			return fmt.Errorf("build: failed to build %s: %w", lib.ID, err)
		}
	}

	slog.Info("build: completed successfully")
	return nil
}

// validateDirs checks that required directories exist.
func validateDirs(cfg *Config) error {
	dirs := map[string]string{
		"librarian": cfg.LibrarianDir,
		"repo":      cfg.RepoDir,
	}

	for name, path := range dirs {
		if info, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("build: %s directory does not exist: %s", name, path)
			}
			return fmt.Errorf("build: failed to stat %s directory: %w", name, err)
		} else if !info.IsDir() {
			return fmt.Errorf("build: %s path is not a directory: %s", name, path)
		}
	}

	return nil
}

// buildLibrary simulates building and testing a library.
func buildLibrary(repoDir string, lib *Library) error {
	slog.Info("build: simulating build", "library", lib.ID)

	// Simulate checking source roots exist
	for _, root := range lib.SourceRoots {
		rootPath := filepath.Join(repoDir, root)
		if _, err := os.Stat(rootPath); err != nil {
			slog.Warn("build: source root not found (continuing anyway)", "path", rootPath)
		} else {
			slog.Debug("build: verified source root", "path", rootPath)
		}
	}

	// Simulate running tests
	slog.Info("build: simulating test execution", "library", lib.ID)
	slog.Info("build: all tests passed", "library", lib.ID)

	// Simulate running build
	slog.Info("build: simulating compilation", "library", lib.ID)
	slog.Info("build: build succeeded", "library", lib.ID)

	return nil
}
