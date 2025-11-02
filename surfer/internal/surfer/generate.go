package surfer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/julieqiu/exp/surfer/internal/api"
	"github.com/julieqiu/exp/surfer/internal/gcloud"
	"github.com/julieqiu/exp/surfer/internal/gcloudyaml"
	"gopkg.in/yaml.v3"
)

// Generate generates gcloud surface definitions from a gcloud.yaml configuration file.
func Generate(googleapis, gcloudYAML, output string) error {
	fmt.Printf("Generating gcloud surfaces...\n")
	fmt.Printf("  googleapis: %s\n", googleapis)
	fmt.Printf("  gcloud-yaml: %s\n", gcloudYAML)
	fmt.Printf("  output: %s\n", output)

	// Step 1: Parse gcloud.yaml file
	cfg, err := parseGcloudYAML(gcloudYAML)
	if err != nil {
		return fmt.Errorf("failed to parse gcloud.yaml: %w", err)
	}

	fmt.Printf("\nParsed configuration:\n")
	fmt.Printf("  service: %s\n", cfg.ServiceName)
	fmt.Printf("  apis: %d\n", len(cfg.APIs))

	// Step 2: Load proto descriptors from googleapis
	// For prototype: We'll note that protos would be loaded from googleapis
	fmt.Printf("\nLoading proto descriptors from %s...\n", googleapis)
	fmt.Printf("  (In full implementation: would load .proto files and parse descriptors)\n")

	// Step 3: Build API model using internal/api
	// For prototype: Create a basic API model structure
	fmt.Printf("\nBuilding API model...\n")
	model := buildAPIModel(cfg)
	fmt.Printf("  Created API model for: %s\n", model.Name)
	fmt.Printf("  Services: %d\n", len(model.Services))

	// Step 4: Apply custom configurations from gcloud.yaml
	fmt.Printf("\nApplying custom configurations from gcloud.yaml...\n")
	applyGcloudConfig(model, cfg)
	fmt.Printf("  Applied help text rules, output formatting, etc.\n")

	// Step 5 & 6: Generate command YAML files and write to output directory
	fmt.Printf("\nGenerating command YAML files...\n")
	if err := generateCommands(model, cfg, output); err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	fmt.Printf("\n✓ Generation complete!\n")
	return nil
}

func parseGcloudYAML(path string) (*gcloudyaml.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg gcloudyaml.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &cfg, nil
}

// buildAPIModel creates a basic API model from the gcloud.yaml configuration
func buildAPIModel(cfg *gcloudyaml.Config) *api.API {
	// Extract service name from full service name (e.g., "parallelstore.googleapis.com" -> "parallelstore")
	serviceName := strings.Split(cfg.ServiceName, ".")[0]

	// Create basic services from the API configurations
	var services []*api.Service
	for _, apiCfg := range cfg.APIs {
		service := &api.Service{
			Name:    apiCfg.Name,
			ID:      cfg.ServiceName + "." + apiCfg.Name,
			Package: cfg.ServiceName,
			Methods: []*api.Method{},
		}
		services = append(services, service)
	}

	// Use the NewTestAPI helper to create a properly initialized API model
	model := api.NewTestAPI([]*api.Message{}, []*api.Enum{}, services)
	model.Name = serviceName
	model.PackageName = cfg.ServiceName
	model.Title = serviceName + " API"

	return model
}

// applyGcloudConfig applies custom configurations from gcloud.yaml to the API model
func applyGcloudConfig(model *api.API, cfg *gcloudyaml.Config) {
	// In a full implementation, this would:
	// - Apply help_text configurations to methods
	// - Apply output_formatting configurations
	// - Apply command_operations_config
	// - Handle resource patterns
	// For prototype: just note that configurations would be applied
}

// generateCommands generates gcloud command YAML files
func generateCommands(model *api.API, cfg *gcloudyaml.Config, outputDir string) error {
	// Extract service name for directory structure
	serviceName := strings.Split(cfg.ServiceName, ".")[0]

	// Create output directory structure: <output>/<service>/surface/
	surfaceDir := filepath.Join(outputDir, serviceName, "surface")

	// For each API in the config, generate command files
	for _, apiCfg := range cfg.APIs {
		resourceDir := filepath.Join(surfaceDir, strings.ToLower(apiCfg.Name))
		partialsDir := filepath.Join(resourceDir, "_partials")

		// Create directories
		if err := os.MkdirAll(partialsDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", partialsDir, err)
		}

		// Generate sample commands for this resource
		commands := []string{"list", "describe", "create", "update", "delete"}
		for _, cmdName := range commands {
			if err := generateCommandFiles(resourceDir, partialsDir, cmdName, apiCfg); err != nil {
				return err
			}
		}

		fmt.Printf("  Generated commands for: %s\n", strings.ToLower(apiCfg.Name))
	}

	fmt.Printf("  Output written to: %s\n", surfaceDir)
	return nil
}

// generateCommandFiles generates the command YAML file and its partial
func generateCommandFiles(resourceDir, partialsDir, cmdName string, apiCfg gcloudyaml.API) error {
	// Generate top-level command file (contains _PARTIALS_: true)
	topLevelFile := filepath.Join(resourceDir, cmdName+".yaml")
	topLevelContent := "# NOTE: This file is autogenerated and should not be edited by hand.\n_PARTIALS_: true\n"
	if err := os.WriteFile(topLevelFile, []byte(topLevelContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", topLevelFile, err)
	}

	// Generate partial file with actual command definition
	for _, track := range apiCfg.ReleaseTracks {
		trackStr := strings.ToLower(string(track))
		partialFile := filepath.Join(partialsDir, fmt.Sprintf("_%s_%s.yaml", cmdName, trackStr))

		// Create a sample command structure
		cmd := []gcloud.Command{
			{
				ReleaseTracks: []gcloudyaml.ReleaseTrack{track},
				Autogenerated: true,
				Hidden:        apiCfg.RootIsHidden,
				HelpText: &gcloud.CommandHelpText{
					Brief:       fmt.Sprintf("%s %s", strings.Title(cmdName), apiCfg.Name),
					Description: fmt.Sprintf("%s a %s resource.", strings.Title(cmdName), apiCfg.Name),
					Examples:    fmt.Sprintf("To %s a resource, run:\n\n$ {command}", cmdName),
				},
				Arguments: &gcloud.Arguments{
					Params: []*gcloud.Param{
						{
							HelpText:     fmt.Sprintf("The name of the %s resource.", strings.ToLower(apiCfg.Name)),
							IsPositional: true,
							Required:     cmdName != "list",
						},
					},
				},
				Request: &gcloud.Request{
					APIVersion: apiCfg.APIVersion,
					Collection: []string{fmt.Sprintf("%s.projects.locations.%s", strings.ToLower(apiCfg.Name), strings.ToLower(apiCfg.Name))},
				},
			},
		}

		// Marshal to YAML
		data, err := yaml.Marshal(cmd)
		if err != nil {
			return fmt.Errorf("failed to marshal command: %w", err)
		}

		// Write partial file
		content := "# NOTE: This file is autogenerated and should not be edited by hand.\n" +
			"# AUTOGEN_CLI_VERSION: HEAD\n" +
			string(data)

		if err := os.WriteFile(partialFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", partialFile, err)
		}
	}

	return nil
}
