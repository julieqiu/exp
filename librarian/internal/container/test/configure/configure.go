package configure

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Config holds the configuration for the configure command.
type Config struct {
	LibrarianDir string
	InputDir     string
	RepoDir      string
	OutputDir    string
	SourceDir    string
}

// Request represents the configure-request.json structure.
type Request struct {
	Libraries []Library `json:"libraries"`
}

// Library represents a library to configure.
type Library struct {
	ID          string   `json:"id"`
	APIs        []API    `json:"apis"`
	SourceRoots []string `json:"source_roots"`
}

// API represents an API with its configuration status.
type API struct {
	Path          string `json:"path"`
	ServiceConfig string `json:"service_config"`
	Status        string `json:"status"` // "new" or "existing"
}

// Configure implements the configure command.
// It reads configure-request.json and creates configuration metadata files.
func Configure(ctx context.Context, cfg *Config) error {
	slog.Info("configure: starting", "config", cfg)

	// Validate directories exist
	if err := validateDirs(cfg); err != nil {
		return err
	}

	// Read and parse the request
	requestPath := filepath.Join(cfg.LibrarianDir, "configure-request.json")
	slog.Debug("configure: reading request", "path", requestPath)

	data, err := os.ReadFile(requestPath)
	if err != nil {
		return fmt.Errorf("configure: failed to read request file: %w", err)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("configure: failed to parse request JSON: %w", err)
	}

	slog.Info("configure: parsed request", "libraries", len(req.Libraries))

	// Create response with all library configurations
	response := map[string]interface{}{
		"libraries": []map[string]interface{}{},
	}

	for _, lib := range req.Libraries {
		slog.Info("configure: processing library", "id", lib.ID, "apis", len(lib.APIs))

		libConfig := map[string]interface{}{
			"library_id": lib.ID,
			"apis":       lib.APIs,
			"validated":  true,
			"status":     "ready",
		}
		response["libraries"] = append(response["libraries"].([]map[string]interface{}), libConfig)
	}

	// Write configure-response.json
	responsePath := filepath.Join(cfg.OutputDir, "configure-response.json")
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("configure: failed to marshal response: %w", err)
	}

	if err := os.WriteFile(responsePath, responseJSON, 0644); err != nil {
		return fmt.Errorf("configure: failed to write response: %w", err)
	}
	slog.Debug("configure: created file", "path", responsePath)

	slog.Info("configure: completed successfully")
	return nil
}

// validateDirs checks that required directories exist.
func validateDirs(cfg *Config) error {
	dirs := map[string]string{
		"librarian": cfg.LibrarianDir,
		"input":     cfg.InputDir,
		"repo":      cfg.RepoDir,
		"output":    cfg.OutputDir,
		"source":    cfg.SourceDir,
	}

	for name, path := range dirs {
		if info, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("configure: %s directory does not exist: %s", name, path)
			}
			return fmt.Errorf("configure: failed to stat %s directory: %w", name, err)
		} else if !info.IsDir() {
			return fmt.Errorf("configure: %s path is not a directory: %s", name, path)
		}
	}

	return nil
}
