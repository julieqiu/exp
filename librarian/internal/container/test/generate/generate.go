package generate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the configuration for the generate command.
type Config struct {
	LibrarianDir string
	InputDir     string
	OutputDir    string
	SourceDir    string
}

// Request represents the generate-request.json structure.
type Request struct {
	ID            string   `json:"id"`
	Version       string   `json:"version"`
	APIs          []API    `json:"apis"`
	SourceRoots   []string `json:"source_roots"`
	PreserveRegex []string `json:"preserve_regex"`
	RemoveRegex   []string `json:"remove_regex"`
}

// API represents an API to generate.
type API struct {
	Path          string `json:"path"`
	ServiceConfig string `json:"service_config"`
}

// Generate implements the generate command.
// It reads generate-request.json and creates placeholder files in the output directory.
func Generate(ctx context.Context, cfg *Config) error {
	slog.Info("generate: starting", "config", cfg)

	// Validate directories exist
	if err := validateDirs(cfg); err != nil {
		return err
	}

	// Read and parse the request
	requestPath := filepath.Join(cfg.LibrarianDir, "generate-request.json")
	slog.Debug("generate: reading request", "path", requestPath)

	data, err := os.ReadFile(requestPath)
	if err != nil {
		return fmt.Errorf("generate: failed to read request file: %w", err)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("generate: failed to parse request JSON: %w", err)
	}

	slog.Info("generate: parsed request", "id", req.ID, "version", req.Version, "apis", len(req.APIs))

	// Create output files
	if err := createOutputFiles(cfg.OutputDir, &req); err != nil {
		return fmt.Errorf("generate: failed to create output files: %w", err)
	}

	slog.Info("generate: completed successfully")
	return nil
}

// validateDirs checks that required directories exist.
func validateDirs(cfg *Config) error {
	dirs := map[string]string{
		"librarian": cfg.LibrarianDir,
		"input":     cfg.InputDir,
		"output":    cfg.OutputDir,
		"source":    cfg.SourceDir,
	}

	for name, path := range dirs {
		if info, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("generate: %s directory does not exist: %s", name, path)
			}
			return fmt.Errorf("generate: failed to stat %s directory: %w", name, err)
		} else if !info.IsDir() {
			return fmt.Errorf("generate: %s path is not a directory: %s", name, path)
		}
	}

	return nil
}

// createOutputFiles generates placeholder files in the output directory.
func createOutputFiles(outputDir string, req *Request) error {
	// Create client.go
	clientPath := filepath.Join(outputDir, "client.go")
	clientContent := generateClientFile(req)
	if err := os.WriteFile(clientPath, []byte(clientContent), 0644); err != nil {
		return fmt.Errorf("failed to write client.go: %w", err)
	}
	slog.Debug("generate: created file", "path", clientPath)

	// Create doc.go
	docPath := filepath.Join(outputDir, "doc.go")
	docContent := generateDocFile(req)
	if err := os.WriteFile(docPath, []byte(docContent), 0644); err != nil {
		return fmt.Errorf("failed to write doc.go: %w", err)
	}
	slog.Debug("generate: created file", "path", docPath)

	// Create version.go
	versionPath := filepath.Join(outputDir, "version.go")
	versionContent := generateVersionFile(req)
	if err := os.WriteFile(versionPath, []byte(versionContent), 0644); err != nil {
		return fmt.Errorf("failed to write version.go: %w", err)
	}
	slog.Debug("generate: created file", "path", versionPath)

	// Create README.md
	readmePath := filepath.Join(outputDir, "README.md")
	readmeContent := generateReadmeFile(req)
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}
	slog.Debug("generate: created file", "path", readmePath)

	return nil
}

// generateClientFile creates placeholder client.go content.
func generateClientFile(req *Request) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("// Package %s provides a client for the %s API.\n", req.ID, req.ID))
	b.WriteString(fmt.Sprintf("package %s\n\n", req.ID))
	b.WriteString("// Client is a client for the API.\n")
	b.WriteString("type Client struct {\n")
	b.WriteString("\t// This is a test container placeholder.\n")
	b.WriteString("}\n\n")
	b.WriteString("// NewClient creates a new client.\n")
	b.WriteString("func NewClient() *Client {\n")
	b.WriteString("\treturn &Client{}\n")
	b.WriteString("}\n\n")

	// Add a method for each API
	for _, api := range req.APIs {
		parts := strings.Split(api.Path, "/")
		apiName := parts[len(parts)-1]
		b.WriteString(fmt.Sprintf("// %sService provides access to the %s API.\n", capitalize(apiName), api.Path))
		b.WriteString(fmt.Sprintf("func (c *Client) %sService() *%sService {\n", capitalize(apiName), capitalize(apiName)))
		b.WriteString(fmt.Sprintf("\treturn &%sService{}\n", capitalize(apiName)))
		b.WriteString("}\n\n")
		b.WriteString(fmt.Sprintf("// %sService is a placeholder service.\n", capitalize(apiName)))
		b.WriteString(fmt.Sprintf("type %sService struct{}\n\n", capitalize(apiName)))
	}

	return b.String()
}

// generateDocFile creates placeholder doc.go content.
func generateDocFile(req *Request) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("// Package %s provides access to the following APIs:\n", req.ID))
	for _, api := range req.APIs {
		b.WriteString(fmt.Sprintf("//   - %s\n", api.Path))
	}
	b.WriteString("//\n")
	b.WriteString("// This is a test container placeholder.\n")
	b.WriteString(fmt.Sprintf("package %s\n", req.ID))

	return b.String()
}

// generateVersionFile creates placeholder version.go content.
func generateVersionFile(req *Request) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", req.ID))
	b.WriteString("// Version is the current version of this library.\n")
	b.WriteString(fmt.Sprintf("const Version = %q\n", req.Version))

	return b.String()
}

// generateReadmeFile creates placeholder README.md content.
func generateReadmeFile(req *Request) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", req.ID))
	b.WriteString(fmt.Sprintf("This library provides access to the %s API.\n\n", req.ID))
	b.WriteString("## Installation\n\n")
	b.WriteString("```bash\n")
	b.WriteString(fmt.Sprintf("go get example.com/%s\n", req.ID))
	b.WriteString("```\n\n")
	b.WriteString("## Usage\n\n")
	b.WriteString("```go\n")
	b.WriteString(fmt.Sprintf("import \"%s\"\n\n", req.ID))
	b.WriteString(fmt.Sprintf("client := %s.NewClient()\n", req.ID))
	b.WriteString("```\n\n")
	b.WriteString("## APIs\n\n")
	for _, api := range req.APIs {
		b.WriteString(fmt.Sprintf("- %s\n", api.Path))
	}
	b.WriteString("\n")
	b.WriteString("---\n\n")
	b.WriteString("*This is a test container placeholder generated for testing purposes.*\n")

	return b.String()
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
