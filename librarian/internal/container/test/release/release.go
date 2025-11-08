package release

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds the configuration for the release-stage command.
type Config struct {
	LibrarianDir string
	RepoDir      string
	OutputDir    string
}

// Request represents the release-stage-request.json structure.
type Request struct {
	Libraries []Library `json:"libraries"`
}

// Library represents a library to prepare for release.
type Library struct {
	ID               string   `json:"id"`
	Version          string   `json:"version"`
	Changes          []Change `json:"changes"`
	APIs             []API    `json:"apis"`
	SourceRoots      []string `json:"source_roots"`
	ReleaseTriggered bool     `json:"release_triggered"`
}

// Change represents a changelog entry.
type Change struct {
	Type           string `json:"type"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	PiperCLNumber  string `json:"piper_cl_number"`
	CommitHash     string `json:"commit_hash"`
}

// API represents an API in the library.
type API struct {
	Path string `json:"path"`
}

// Stage implements the release-stage command.
// It reads release-stage-request.json and creates updated version and changelog files.
func Stage(ctx context.Context, cfg *Config) error {
	slog.Info("release-stage: starting", "config", cfg)

	// Validate directories exist
	if err := validateDirs(cfg); err != nil {
		return err
	}

	// Read and parse the request
	requestPath := filepath.Join(cfg.LibrarianDir, "release-stage-request.json")
	slog.Debug("release-stage: reading request", "path", requestPath)

	data, err := os.ReadFile(requestPath)
	if err != nil {
		return fmt.Errorf("release-stage: failed to read request file: %w", err)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("release-stage: failed to parse request JSON: %w", err)
	}

	slog.Info("release-stage: parsed request", "libraries", len(req.Libraries))

	// Process each library
	for _, lib := range req.Libraries {
		slog.Info("release-stage: processing library", "id", lib.ID, "version", lib.Version)

		if err := createReleaseFiles(cfg.OutputDir, &lib); err != nil {
			return fmt.Errorf("release-stage: failed to create files for %s: %w", lib.ID, err)
		}
	}

	slog.Info("release-stage: completed successfully")
	return nil
}

// validateDirs checks that required directories exist.
func validateDirs(cfg *Config) error {
	dirs := map[string]string{
		"librarian": cfg.LibrarianDir,
		"repo":      cfg.RepoDir,
		"output":    cfg.OutputDir,
	}

	for name, path := range dirs {
		if info, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("release-stage: %s directory does not exist: %s", name, path)
			}
			return fmt.Errorf("release-stage: failed to stat %s directory: %w", name, err)
		} else if !info.IsDir() {
			return fmt.Errorf("release-stage: %s path is not a directory: %s", name, path)
		}
	}

	return nil
}

// createReleaseFiles generates version.go and CHANGES.md files.
func createReleaseFiles(outputDir string, lib *Library) error {
	// Create version.go
	versionPath := filepath.Join(outputDir, "version.go")
	versionContent := generateVersionFile(lib)
	if err := os.WriteFile(versionPath, []byte(versionContent), 0644); err != nil {
		return fmt.Errorf("failed to write version.go: %w", err)
	}
	slog.Debug("release-stage: created file", "path", versionPath)

	// Create CHANGES.md
	changesPath := filepath.Join(outputDir, "CHANGES.md")
	changesContent := generateChangesFile(lib)
	if err := os.WriteFile(changesPath, []byte(changesContent), 0644); err != nil {
		return fmt.Errorf("failed to write CHANGES.md: %w", err)
	}
	slog.Debug("release-stage: created file", "path", changesPath)

	return nil
}

// generateVersionFile creates version.go content with the new version.
func generateVersionFile(lib *Library) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", lib.ID))
	b.WriteString("// Version is the current version of this library.\n")
	b.WriteString(fmt.Sprintf("const Version = %q\n", lib.Version))

	return b.String()
}

// generateChangesFile creates CHANGES.md content with new changelog entries.
func generateChangesFile(lib *Library) string {
	var b strings.Builder

	// Header
	b.WriteString("# Changelog\n\n")

	// New version section
	b.WriteString(fmt.Sprintf("## %s\n\n", lib.Version))
	b.WriteString(fmt.Sprintf("Released: %s\n\n", time.Now().Format("2006-01-02")))

	// Group changes by type
	features := []Change{}
	fixes := []Change{}
	other := []Change{}

	for _, change := range lib.Changes {
		switch change.Type {
		case "feat":
			features = append(features, change)
		case "fix":
			fixes = append(fixes, change)
		default:
			other = append(other, change)
		}
	}

	// Features
	if len(features) > 0 {
		b.WriteString("### Features\n\n")
		for _, change := range features {
			b.WriteString(fmt.Sprintf("- %s", change.Subject))
			if change.Body != "" {
				b.WriteString(fmt.Sprintf(": %s", change.Body))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Bug Fixes
	if len(fixes) > 0 {
		b.WriteString("### Bug Fixes\n\n")
		for _, change := range fixes {
			b.WriteString(fmt.Sprintf("- %s", change.Subject))
			if change.Body != "" {
				b.WriteString(fmt.Sprintf(": %s", change.Body))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Other changes
	if len(other) > 0 {
		b.WriteString("### Other Changes\n\n")
		for _, change := range other {
			b.WriteString(fmt.Sprintf("- %s", change.Subject))
			if change.Body != "" {
				b.WriteString(fmt.Sprintf(": %s", change.Body))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("---\n\n")
	b.WriteString("*This changelog was generated by the test container.*\n")

	return b.String()
}
