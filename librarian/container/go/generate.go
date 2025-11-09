// Package gogenerator implements the Go client library generator.
// It generates Go GAPIC client libraries from googleapis proto definitions.
package gogenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Config holds configuration for the generator.
type Config struct {
	// LibrarianDir is the path to the librarian input directory
	// (contains generate-request.json and repo-config.yaml)
	LibrarianDir string

	// SourceDir is the path to the complete googleapis repository checkout
	SourceDir string

	// OutputDir is the path where generated code will be written
	OutputDir string

	// InputDir is the path to generator input templates/config
	InputDir string

	// DisablePostProcessor controls whether post-processing runs
	DisablePostProcessor bool
}

// GenerateRequest represents the JSON request file for generation.
type GenerateRequest struct {
	ID            string   `json:"id"`             // Module name (e.g., "functions")
	Version       string   `json:"version"`        // Semantic version (e.g., "1.2.0")
	APIs          []API    `json:"apis"`           // APIs to generate
	SourceRoots   []string `json:"source_roots"`   // Directories receiving generated code
	PreserveRegex []string `json:"preserve_regex"` // Files to leave untouched
	RemoveRegex   []string `json:"remove_regex"`   // Files to remove
	Status        string   `json:"status"`         // "new" or "existing"
}

// API represents a single API to generate.
type API struct {
	Path          string `json:"path"`           // Path in googleapis (e.g., "google/cloud/functions/v2")
	ServiceConfig string `json:"service_config"` // Service config filename
	Status        string `json:"status"`         // "new" or "existing"
}

// RepoConfig represents the repo-config.yaml file.
type RepoConfig struct {
	Modules []*ModuleConfig `yaml:"modules"`
}

// ModuleConfig holds per-module configuration overrides.
type ModuleConfig struct {
	Name                        string       `yaml:"name"`                           // Module name
	ModulePathVersion           string       `yaml:"module_path_version"`            // Major version (e.g., "v2")
	APIs                        []*APIConfig `yaml:"apis"`                           // Per-API overrides
	DeleteGenerationOutputPaths []string     `yaml:"delete_generation_output_paths"` // Paths to delete
}

// APIConfig holds per-API configuration overrides.
type APIConfig struct {
	Path            string   `yaml:"path"`             // googleapis API path
	ProtoPackage    string   `yaml:"proto_package"`    // Override proto package name
	ClientDirectory string   `yaml:"client_directory"` // Override client dir for snippets
	DisableGAPIC    bool     `yaml:"disable_gapic"`    // Disable GAPIC for this API
	NestedProtos    []string `yaml:"nested_protos"`    // Nested proto files to include
	ModuleName      string   // Populated at runtime
}

// BazelConfig represents configuration extracted from BUILD.bazel files.
type BazelConfig struct {
	grpcServiceConfig string
	gapicImportPath   string
	metadata          bool
	releaseLevel      string
	restNumericEnums  bool
	serviceYAML       string
	transport         string
	diregapic         bool
	hasGoGRPC         bool
	hasGAPIC          bool
	hasLegacyGRPC     bool
}

// Generate runs the complete generation workflow: generate code, build, and validate.
func Generate(ctx context.Context, cfg *Config) error {
	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Read generate request
	req, err := readGenerateRequest(cfg.LibrarianDir)
	if err != nil {
		return fmt.Errorf("failed to read generate request: %w", err)
	}

	// Load repo config (optional)
	repoConfig, err := loadRepoConfig(cfg.LibrarianDir)
	if err != nil {
		return fmt.Errorf("failed to load repo config: %w", err)
	}

	// Get module config for this library
	moduleConfig := getModuleConfig(repoConfig, req.ID)

	// Invoke protoc for each API
	if err := invokeProtoc(ctx, cfg, req, moduleConfig); err != nil {
		return fmt.Errorf("protoc generation failed: %w", err)
	}

	// Fix file permissions
	if err := fixPermissions(cfg.OutputDir); err != nil {
		return fmt.Errorf("failed to fix permissions: %w", err)
	}

	// Flatten output directory structure
	if err := flattenOutput(cfg.OutputDir); err != nil {
		return fmt.Errorf("failed to flatten output: %w", err)
	}

	// Apply module version (move versioned dirs to correct location)
	modulePath := getModulePath(moduleConfig, req.ID)
	if err := applyModuleVersion(cfg.OutputDir, req.ID, modulePath); err != nil {
		return fmt.Errorf("failed to apply module version: %w", err)
	}

	// Post-process if enabled
	if !cfg.DisablePostProcessor {
		if err := postProcess(ctx, req, cfg.OutputDir, moduleConfig); err != nil {
			return fmt.Errorf("post-processing failed: %w", err)
		}
	}

	// Delete unwanted paths
	if err := deleteOutputPaths(cfg.OutputDir, moduleConfig.DeleteGenerationOutputPaths); err != nil {
		return fmt.Errorf("failed to delete output paths: %w", err)
	}

	// Build and test (validation)
	if err := build(ctx, cfg.OutputDir, req.ID); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if err := test(ctx, cfg.OutputDir, req.ID); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	return nil
}

// validateConfig checks that all required configuration is present.
func validateConfig(cfg *Config) error {
	if cfg.LibrarianDir == "" {
		return fmt.Errorf("LibrarianDir is required")
	}
	if cfg.SourceDir == "" {
		return fmt.Errorf("SourceDir is required")
	}
	if cfg.OutputDir == "" {
		return fmt.Errorf("OutputDir is required")
	}
	return nil
}

// readGenerateRequest reads and parses the generate-request.json file.
func readGenerateRequest(librarianDir string) (*GenerateRequest, error) {
	path := filepath.Join(librarianDir, "generate-request.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var req GenerateRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &req, nil
}

// loadRepoConfig loads the repo-config.yaml file if it exists.
func loadRepoConfig(librarianDir string) (*RepoConfig, error) {
	// For now, return empty config - YAML parsing would require gopkg.in/yaml.v3
	// This is a placeholder for the actual implementation
	return &RepoConfig{}, nil
}

// getModuleConfig finds the module config for the given library ID.
func getModuleConfig(repoConfig *RepoConfig, libraryID string) *ModuleConfig {
	for _, mod := range repoConfig.Modules {
		if mod.Name == libraryID {
			return mod
		}
	}
	// Return empty config if not found
	return &ModuleConfig{Name: libraryID}
}

// invokeProtoc runs protoc for each API in the request.
func invokeProtoc(ctx context.Context, cfg *Config, req *GenerateRequest, moduleConfig *ModuleConfig) error {
	for _, api := range req.APIs {
		apiDir := filepath.Join(cfg.SourceDir, api.Path)

		// Parse BUILD.bazel to get configuration
		bazelConfig, err := parseBazelConfig(apiDir)
		if err != nil {
			return fmt.Errorf("failed to parse BUILD.bazel for %s: %w", api.Path, err)
		}

		// Check if GAPIC is disabled for this API
		apiConfig := getAPIConfig(moduleConfig, api.Path)
		if apiConfig.DisableGAPIC {
			bazelConfig.hasGAPIC = false
		}

		// Build protoc command arguments
		args, err := buildProtocArgs(cfg, req, &api, bazelConfig, apiConfig)
		if err != nil {
			return fmt.Errorf("failed to build protoc args for %s: %w", api.Path, err)
		}

		// Execute protoc
		if err := runCommand(ctx, args, ""); err != nil {
			return fmt.Errorf("protoc failed for %s: %w", api.Path, err)
		}

		// Generate .repo-metadata.json for this API
		if err := generateRepoMetadata(cfg.OutputDir, req, &api); err != nil {
			return fmt.Errorf("failed to generate repo metadata for %s: %w", api.Path, err)
		}
	}

	return nil
}

// parseBazelConfig parses a BUILD.bazel file to extract configuration.
func parseBazelConfig(dir string) (*BazelConfig, error) {
	path := filepath.Join(dir, "BUILD.bazel")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read BUILD.bazel: %w", err)
	}

	content := string(data)
	cfg := &BazelConfig{}

	// Check for go_gapic_library rule
	cfg.hasGAPIC = strings.Contains(content, "go_gapic_library(")

	// Check for go_grpc_library (modern) or go_proto_library (legacy)
	cfg.hasGoGRPC = strings.Contains(content, "go_grpc_library(")
	if !cfg.hasGoGRPC {
		// Check for legacy go_proto_library with gRPC plugin
		cfg.hasLegacyGRPC = strings.Contains(content, "go_proto_library(") &&
			strings.Contains(content, "@io_bazel_rules_go//proto:go_grpc")
	}

	if cfg.hasGAPIC {
		// Extract GAPIC configuration
		cfg.gapicImportPath = findString(content, "importpath")
		cfg.serviceYAML = findString(content, "service_yaml")
		cfg.grpcServiceConfig = findString(content, "grpc_service_config")
		cfg.transport = findString(content, "transport")
		cfg.releaseLevel = findString(content, "release_level")
		cfg.metadata, _ = findBool(content, "metadata")
		cfg.diregapic, _ = findBool(content, "diregapic")
		cfg.restNumericEnums, _ = findBool(content, "rest_numeric_enums")
	}

	return cfg, nil
}

// findString extracts a string value from Bazel configuration.
func findString(content, name string) string {
	pattern := fmt.Sprintf(`%s\s*=\s*"([^"]+)"`, name)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// findBool extracts a boolean value from Bazel configuration.
func findBool(content, name string) (bool, error) {
	pattern := fmt.Sprintf(`%s\s*=\s*(\w+)`, name)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strconv.ParseBool(matches[1])
	}
	return false, nil
}

// getAPIConfig finds the API config for the given API path.
func getAPIConfig(moduleConfig *ModuleConfig, apiPath string) *APIConfig {
	for _, api := range moduleConfig.APIs {
		if api.Path == apiPath {
			return api
		}
	}
	return &APIConfig{Path: apiPath}
}

// buildProtocArgs constructs the protoc command arguments.
func buildProtocArgs(cfg *Config, req *GenerateRequest, api *API, bazelConfig *BazelConfig, apiConfig *APIConfig) ([]string, error) {
	apiDir := filepath.Join(cfg.SourceDir, api.Path)

	// Gather proto files
	protoFiles, err := gatherProtoFiles(apiDir, apiConfig.NestedProtos)
	if err != nil {
		return nil, err
	}

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no proto files found in %s", apiDir)
	}

	args := []string{"protoc", "--experimental_allow_proto3_optional"}

	// Add proto/gRPC output plugins
	if bazelConfig.hasGoGRPC {
		// Modern: go_grpc_library
		args = append(args, fmt.Sprintf("--go_out=%s", cfg.OutputDir))
		args = append(args, fmt.Sprintf("--go-grpc_out=%s", cfg.OutputDir))
		args = append(args, "--go-grpc_opt=require_unimplemented_servers=false")
	} else {
		// Legacy: go_proto_library
		args = append(args, fmt.Sprintf("--go_v1_out=%s", cfg.OutputDir))
		if bazelConfig.hasLegacyGRPC {
			args = append(args, "--go_v1_opt=plugins=grpc")
		}
	}

	// Add GAPIC plugin if enabled
	if bazelConfig.hasGAPIC {
		args = append(args, fmt.Sprintf("--go_gapic_out=%s", cfg.OutputDir))

		// Add GAPIC options
		if bazelConfig.gapicImportPath != "" {
			args = append(args, fmt.Sprintf("--go_gapic_opt=go-gapic-package=%s", bazelConfig.gapicImportPath))
		}
		if bazelConfig.serviceYAML != "" {
			args = append(args, fmt.Sprintf("--go_gapic_opt=api-service-config=%s/%s",
				apiDir, bazelConfig.serviceYAML))
		}
		if bazelConfig.grpcServiceConfig != "" {
			args = append(args, fmt.Sprintf("--go_gapic_opt=grpc-service-config=%s/%s",
				apiDir, bazelConfig.grpcServiceConfig))
		}
		if bazelConfig.transport != "" {
			args = append(args, fmt.Sprintf("--go_gapic_opt=transport=%s", bazelConfig.transport))
		}
		if bazelConfig.releaseLevel != "" {
			args = append(args, fmt.Sprintf("--go_gapic_opt=release-level=%s", bazelConfig.releaseLevel))
		}
		if bazelConfig.metadata {
			args = append(args, "--go_gapic_opt=metadata")
		}
		if bazelConfig.diregapic {
			args = append(args, "--go_gapic_opt=diregapic")
		}
		if bazelConfig.restNumericEnums {
			args = append(args, "--go_gapic_opt=rest-numeric-enums")
		}
	}

	// Add import path
	args = append(args, fmt.Sprintf("-I=%s", cfg.SourceDir))

	// Add proto files
	args = append(args, protoFiles...)

	return args, nil
}

// gatherProtoFiles collects all .proto files from the API directory.
func gatherProtoFiles(apiDir string, nestedProtos []string) ([]string, error) {
	var protoFiles []string

	// Read top-level .proto files
	entries, err := os.ReadDir(apiDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".proto") {
			protoFiles = append(protoFiles, filepath.Join(apiDir, entry.Name()))
		}
	}

	// Add nested proto files
	for _, nested := range nestedProtos {
		protoFiles = append(protoFiles, filepath.Join(apiDir, nested))
	}

	return protoFiles, nil
}

// generateRepoMetadata creates a .repo-metadata.json file for the API.
func generateRepoMetadata(outputDir string, req *GenerateRequest, api *API) error {
	// This is a placeholder - actual implementation would generate proper metadata
	// For now, we'll skip this as it's not critical for the core generation flow
	return nil
}

// fixPermissions sets all .go files to 0644.
func fixPermissions(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			if err := os.Chmod(path, 0644); err != nil {
				return fmt.Errorf("failed to chmod %s: %w", path, err)
			}
		}
		return nil
	})
}

// flattenOutput moves cloud.google.com/go/* to the top level.
func flattenOutput(outputDir string) error {
	cloudPath := filepath.Join(outputDir, "cloud.google.com", "go")
	if _, err := os.Stat(cloudPath); os.IsNotExist(err) {
		// Nothing to flatten
		return nil
	}

	// List contents of cloud.google.com/go/
	entries, err := os.ReadDir(cloudPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", cloudPath, err)
	}

	// Move each entry to output root
	for _, entry := range entries {
		src := filepath.Join(cloudPath, entry.Name())
		dst := filepath.Join(outputDir, entry.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
		}
	}

	// Remove empty cloud.google.com directory
	if err := os.RemoveAll(filepath.Join(outputDir, "cloud.google.com")); err != nil {
		return fmt.Errorf("failed to remove cloud.google.com: %w", err)
	}

	return nil
}

// applyModuleVersion reorganizes versioned modules.
func applyModuleVersion(outputDir, libraryID, modulePath string) error {
	// Extract version from module path (e.g., cloud.google.com/go/dataproc/v2 -> v2)
	parts := strings.Split(modulePath, "/")
	if len(parts) == 0 {
		return nil
	}

	lastPart := parts[len(parts)-1]
	if !strings.HasPrefix(lastPart, "v") {
		// Not a versioned module
		return nil
	}

	version := lastPart

	// Move {id}/{version}/* to {id}/*
	versionedPath := filepath.Join(outputDir, libraryID, version)
	if _, err := os.Stat(versionedPath); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(versionedPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", versionedPath, err)
	}

	targetPath := filepath.Join(outputDir, libraryID)
	for _, entry := range entries {
		src := filepath.Join(versionedPath, entry.Name())
		dst := filepath.Join(targetPath, entry.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
		}
	}

	// Remove empty version directory
	if err := os.RemoveAll(versionedPath); err != nil {
		return fmt.Errorf("failed to remove %s: %w", versionedPath, err)
	}

	// Do the same for snippets
	snippetsVersionedPath := filepath.Join(outputDir, "internal", "generated", "snippets", libraryID, version)
	if _, err := os.Stat(snippetsVersionedPath); err == nil {
		entries, err := os.ReadDir(snippetsVersionedPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", snippetsVersionedPath, err)
		}

		snippetsTargetPath := filepath.Join(outputDir, "internal", "generated", "snippets", libraryID)
		for _, entry := range entries {
			src := filepath.Join(snippetsVersionedPath, entry.Name())
			dst := filepath.Join(snippetsTargetPath, entry.Name())
			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
			}
		}

		if err := os.RemoveAll(snippetsVersionedPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", snippetsVersionedPath, err)
		}
	}

	return nil
}

// postProcess runs post-processing steps: goimports, go mod init/tidy.
func postProcess(ctx context.Context, req *GenerateRequest, outputDir string, moduleConfig *ModuleConfig) error {
	if len(req.APIs) == 0 {
		// Proto-only package, skip post-processing
		return nil
	}

	if req.Version == "" {
		return fmt.Errorf("version is required for post-processing")
	}

	// Update snippet metadata
	if err := updateSnippetsMetadata(outputDir, req); err != nil {
		return fmt.Errorf("failed to update snippets metadata: %w", err)
	}

	// Run goimports
	if err := runCommand(ctx, []string{"goimports", "-w", "."}, outputDir); err != nil {
		return fmt.Errorf("goimports failed: %w", err)
	}

	// Run go mod init/tidy only for new modules
	if req.Status == "new" && len(req.APIs) > 0 {
		modulePath := getModulePath(moduleConfig, req.ID)
		moduleDir := filepath.Join(outputDir, req.ID)

		if err := runCommand(ctx, []string{"go", "mod", "init", modulePath}, moduleDir); err != nil {
			return fmt.Errorf("go mod init failed: %w", err)
		}

		if err := runCommand(ctx, []string{"go", "mod", "tidy"}, moduleDir); err != nil {
			return fmt.Errorf("go mod tidy failed: %w", err)
		}
	}

	return nil
}

// updateSnippetsMetadata updates snippet metadata files with the new version.
func updateSnippetsMetadata(outputDir string, req *GenerateRequest) error {
	snippetsDir := filepath.Join(outputDir, "internal", "generated", "snippets", req.ID)
	if _, err := os.Stat(snippetsDir); os.IsNotExist(err) {
		// No snippets directory, skip
		return nil
	}

	// Walk through snippet metadata files and replace $VERSION
	return filepath.WalkDir(snippetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasPrefix(d.Name(), "snippet_metadata.") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Replace $VERSION placeholder with actual version
		updated := strings.ReplaceAll(string(data), "$VERSION", req.Version)

		return os.WriteFile(path, []byte(updated), 0644)
	})
}

// getModulePath returns the Go module path for the library.
func getModulePath(moduleConfig *ModuleConfig, libraryID string) string {
	if moduleConfig.ModulePathVersion != "" {
		return fmt.Sprintf("cloud.google.com/go/%s/%s", libraryID, moduleConfig.ModulePathVersion)
	}
	return fmt.Sprintf("cloud.google.com/go/%s", libraryID)
}

// deleteOutputPaths removes unwanted paths from the output directory.
func deleteOutputPaths(outputDir string, paths []string) error {
	for _, p := range paths {
		fullPath := filepath.Join(outputDir, p)
		if err := os.RemoveAll(fullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete %s: %w", fullPath, err)
		}
	}
	return nil
}

// build runs go build on the generated code.
func build(ctx context.Context, outputDir, libraryID string) error {
	moduleDir := filepath.Join(outputDir, libraryID)
	return runCommand(ctx, []string{"go", "build", "./..."}, moduleDir)
}

// test runs go test on the generated code.
func test(ctx context.Context, outputDir, libraryID string) error {
	moduleDir := filepath.Join(outputDir, libraryID)
	return runCommand(ctx, []string{"go", "test", "./...", "-short"}, moduleDir)
}

// runCommand executes a command in the given working directory.
func runCommand(ctx context.Context, args []string, workingDir string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s\nOutput: %s", err, string(output))
	}

	return nil
}
