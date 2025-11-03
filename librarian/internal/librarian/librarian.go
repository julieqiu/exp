package librarian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/julieqiu/exp/librarian/internal/config"
	"github.com/julieqiu/exp/librarian/internal/state"
	"github.com/urfave/cli/v3"
)

func NewApp() *cli.Command {
	return &cli.Command{
		Name:  "librarian",
		Usage: "A comprehensive CLI for managing software artifact lifecycle, from initialization and code generation to release automation",
		Commands: []*cli.Command{
			{
				Name:      "init",
				Usage:     "Initialize a new librarian-managed repository",
				Arguments: []cli.Argument{&cli.StringArg{Name: "language"}},
				Action:    initCommand,
				Category:  "SETUP",
			},
			{
				Name:  "config",
				Usage: "Manage configuration",
				Commands: []*cli.Command{
					{
						Name:      "get",
						Usage:     "Read a configuration value",
						Arguments: []cli.Argument{&cli.StringArg{Name: "key"}},
						Action:    configGetCommand,
					},
					{
						Name:      "set",
						Usage:     "Set a configuration value",
						Arguments: []cli.Argument{&cli.StringArg{Name: "key"}, &cli.StringArg{Name: "value"}},
						Action:    configSetCommand,
					},
					{
						Name:      "update",
						Usage:     "Update toolchain versions to latest",
						Arguments: []cli.Argument{&cli.StringArg{Name: "key"}},
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all",
								Usage: "Update all toolchain versions",
							},
						},
						Action: configUpdateCommand,
					},
				},
				Category: "SETUP",
			},
			{
				Name:      "add",
				Usage:     "Track a directory for management",
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}, &cli.StringArg{Name: "api"}},
				Action:    addCommand,
				Category:  "MANAGE",
			},
			{
				Name:      "edit",
				Usage:     "Edit artifact configuration",
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}},
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "keep",
						Usage: "Files/directories to keep (don't overwrite) during generation",
					},
					&cli.StringSliceFlag{
						Name:  "remove",
						Usage: "Files to remove after generation",
					},
					&cli.StringSliceFlag{
						Name:  "exclude",
						Usage: "Files to exclude from release",
					},
					&cli.StringSliceFlag{
						Name:  "language",
						Usage: "Language-specific metadata (format: LANG:KEY=VALUE, e.g., go:module=github.com/user/repo)",
					},
				},
				Action:   editCommand,
				Category: "MANAGE",
			},
			{
				Name:      "remove",
				Usage:     "Stop tracking a directory",
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}},
				Action:    removeCommand,
				Category:  "MANAGE",
			},
			{
				Name:  "generate",
				Usage: "Generate or regenerate code for tracked directories",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "Regenerate all artifacts",
					},
				},
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}},
				Action:    generateCommand,
				Category:  "MANAGE",
			},
			{
				Name:  "prepare",
				Usage: "Prepare a release with version updates and notes",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "Prepare all artifacts for release",
					},
				},
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}},
				Action:    prepareCommand,
				Category:  "MANAGE",
			},
			{
				Name:  "release",
				Usage: "Tag and publish a prepared release",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "Release all prepared artifacts",
					},
				},
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}},
				Action:    releaseCommand,
				Category:  "MANAGE",
			},
			{
				Name:     "list",
				Usage:    "List all tracked directories",
				Action:   listCommand,
				Category: "SHOW",
			},
			{
				Name:      "status",
				Usage:     "Show generation and release status",
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}},
				Action:    statusCommand,
				Category:  "SHOW",
			},
			{
				Name:      "history",
				Usage:     "View release history",
				Arguments: []cli.Argument{&cli.StringArg{Name: "path"}},
				Action:    historyCommand,
				Category:  "SHOW",
			},
		},
	}
}

func initCommand(ctx context.Context, cmd *cli.Command) error {
	language := cmd.StringArg("language")
	if cmd.NArg() == 0 {
		language = ""
	}
	supportedLanguages := []string{"go", "python", "rust", "dart", ""}
	isSupported := false
	for _, l := range supportedLanguages {
		if language == l {
			isSupported = true
			break
		}
	}
	if !isSupported {
		return fmt.Errorf("language must be one of: %s", strings.Join(supportedLanguages, ", "))
	}


librarianVersion, err := getLibrarianVersion()
	if err != nil {
		return err
	}

	cfg := &config.Config{
		Librarian: config.LibrarianConfig{
			Version: librarianVersion,
		},
		Release: config.ReleaseConfig{
			TagFormat: "{name}-v{version}",
		},
	}
	if language != "" {
		cfg.Librarian.Language = language
		cfg.Generate = config.GenerateConfig{
			Container: config.ContainerConfig{
				Image: fmt.Sprintf("us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/%s-librarian-generator", language),
				Tag:   "latest",
			},
			GoogleapisRepo: "github.com/googleapis/googleapis",
			GoogleapisRef:  "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0",
			DiscoveryRepo:  "github.com/googleapis/discovery-artifact-manager",
			DiscoveryRef:   "f9e8d7c6b5a4f3e2d1c0b9a8f7e6d5c4b3a2f1e0",
			Dir:            "generated/",
		}
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	runYamlFmt(".librarian/config.yaml")

	if language == "" {
		fmt.Println("Initialized release-only librarian repository")
	} else {
		fmt.Printf("Initialized librarian repository for %s\n", language)
	}
	fmt.Println("Created .librarian/config.yaml")
	return nil
}

func addCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.StringArg("path")
	api := cmd.StringArg("api")
	if path == "" {
		return fmt.Errorf("path is required")
	}
	fmt.Printf("Adding %s to librarian.\n", path)
	if api != "" {
		fmt.Printf("With API: %s\n", api)
	}
	artifact := &state.Artifact{}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Release.TagFormat != "" {
		artifact.Release = &state.ReleaseState{
			Version: "null",
		}
	}
	if api != "" && cfg.Librarian.Language != "" {
		if err := ensureGenerationConfig(cfg); err != nil {
			return err
		}
		artifact.Generate = &state.GenerateState{
			APIs:      []state.API{{Path: api}},
			Commit:    "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0", // Dummy value
			Librarian: cfg.Librarian.Version,
			Container: state.ContainerState{
				Image: cfg.Generate.Container.Image,
				Tag:   cfg.Generate.Container.Tag,
			},
			Googleapis: state.GoogleapisState{
				Repo: cfg.Generate.GoogleapisRepo,
				Ref:  cfg.Generate.GoogleapisRef,
			},
			Discovery: state.DiscoveryState{
				Repo: cfg.Generate.DiscoveryRepo,
				Ref:  cfg.Generate.DiscoveryRef,
			},
		}
	}
	if err := artifact.Save(path); err != nil {
		return err
	}
	runYamlFmt(filepath.Join(path, ".librarian.yaml"))
	fmt.Printf("Created %s/.librarian.yaml\n", path)
	return nil
}

// ensureGenerationConfig initializes generation-related config fields if they're not set.
func ensureGenerationConfig(cfg *config.Config) error {
	var updated bool

	// Initialize generator image if not set
	if cfg.Generate.Container.Image == "" {
		if cfg.Librarian.Language == "python" {
			cfg.Generate.Container.Image = "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator"
			cfg.Generate.Container.Tag = "latest"
		} else if cfg.Librarian.Language == "go" {
			cfg.Generate.Container.Image = "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/go-librarian-generator"
			cfg.Generate.Container.Tag = "latest"
		}
		updated = true
	}

	// Initialize googleapis SHA if not set
	if cfg.Generate.GoogleapisRef == "" {
		googleapisSHA, err := getLatestSHA("googleapis", "googleapis")
		if err != nil {
			return fmt.Errorf("failed to get latest googleapis SHA: %w", err)
		}
		cfg.Generate.GoogleapisRef = googleapisSHA
		updated = true
	}

	// Initialize discovery SHA if not set
	if cfg.Generate.DiscoveryRef == "" {
		discoverySHA, err := getLatestSHA("googleapis", "discovery-artifact-manager")
		if err != nil {
			return fmt.Errorf("failed to get latest discovery SHA: %w", err)
		}
		cfg.Generate.DiscoveryRef = discoverySHA
		updated = true
	}

	if updated {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		runYamlFmt(".librarian/config.yaml")
		fmt.Println("Initialized generation configuration")
	}

	return nil
}

func generateCommand(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")
	path := cmd.StringArg("path")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.Librarian.Language == "" {
		return fmt.Errorf("repository not configured for generation")
	}

	// Ensure generation config is initialized
	if err := ensureGenerationConfig(cfg); err != nil {
		return err
	}

	if all {
		// Regenerate all artifacts
		artifacts, err := state.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load artifacts: %w", err)
		}

		fmt.Printf("Regenerating all %d artifacts...\n", len(artifacts))
		for path, artifact := range artifacts {
			if artifact.Generate == nil {
				continue
			}
			fmt.Printf("  - Regenerating %s\n", path)

			// Sync artifact state with current config
			artifact.Generate.Librarian = cfg.Librarian.Version
			artifact.Generate.Container.Image = cfg.Generate.Container.Image
			artifact.Generate.Container.Tag = cfg.Generate.Container.Tag
			artifact.Generate.Googleapis.Repo = cfg.Generate.GoogleapisRepo
			artifact.Generate.Googleapis.Ref = cfg.Generate.GoogleapisRef
			artifact.Generate.Discovery.Repo = cfg.Generate.DiscoveryRepo
			artifact.Generate.Discovery.Ref = cfg.Generate.DiscoveryRef

			if err := artifact.Save(path); err != nil {
				return fmt.Errorf("failed to save artifact state: %w", err)
			}
			runYamlFmt(filepath.Join(path, ".librarian.yaml"))

			// TODO: Run generator for each artifact
		}
		fmt.Println("Generation complete")
		return nil
	}

	if path == "" {
		return fmt.Errorf("path is required (or use --all)")
	}

	// Check if artifact exists
	artifact, err := state.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load artifact: %w", err)
	}
	if artifact.Generate == nil {
		return fmt.Errorf("artifact at %s is not configured for generation", path)
	}

	// Regenerating existing artifact - sync state with current config
	fmt.Printf("Regenerating artifact at %s...\n", path)

	artifact.Generate.Librarian = cfg.Librarian.Version
	artifact.Generate.Container.Image = cfg.Generate.Container.Image
	artifact.Generate.Container.Tag = cfg.Generate.Container.Tag
	artifact.Generate.Googleapis.Repo = cfg.Generate.GoogleapisRepo
	artifact.Generate.Googleapis.Ref = cfg.Generate.GoogleapisRef
	artifact.Generate.Discovery.Repo = cfg.Generate.DiscoveryRepo
	artifact.Generate.Discovery.Ref = cfg.Generate.DiscoveryRef

	// Save artifact state
	if err := artifact.Save(path); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(filepath.Join(path, ".librarian.yaml"))

	fmt.Println("Running generator...")
	// TODO: Actually run the generator container
	fmt.Println("Generation complete")
	return nil
}

func configGetCommand(ctx context.Context, cmd *cli.Command) error {
	key := cmd.StringArg("key")
	if key == "" {
		return fmt.Errorf("key is required")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	value, err := cfg.Get(key)
	if err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}

func configSetCommand(ctx context.Context, cmd *cli.Command) error {
	key := cmd.StringArg("key")
	value := cmd.StringArg("value")

	if key == "" || value == "" {
		return fmt.Errorf("key and value are required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Set(key, value); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	runYamlFmt(".librarian/config.yaml")

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func configUpdateCommand(ctx context.Context, cmd *cli.Command) error {
	key := cmd.StringArg("key")
	all := cmd.Bool("all")

	if cmd.NArg() == 0 && !all {
		return fmt.Errorf("key or --all is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Checking for updates...")
	var updated bool

	updateAll := all || key == "all"
	updateGeneratorImage := updateAll || key == "generator.image"
	updateGoogleapis := updateAll || key == "generator.googleapis"
	updateDiscovery := updateAll || key == "generator.discovery"

	// Update librarian version
	fmt.Printf("Current librarian version: %s\n", cfg.Librarian.Version)

librarianVersion, err := getLibrarianVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest librarian version: %w", err)
	}
	if librarianVersion != cfg.Librarian.Version {
		fmt.Printf("Updating librarian version to %s\n", librarianVersion)
		cfg.Librarian.Version = librarianVersion
		updated = true
	} else {
		fmt.Println("Librarian version is up to date")
	}

	// Update googleapis SHA if generate config exists
	if cfg.Librarian.Language != "" && cfg.Generate.GoogleapisRef != "" && updateGoogleapis {
		googleapisSHA, err := getLatestSHA("googleapis", "googleapis")
		if err != nil {
			return fmt.Errorf("failed to get latest googleapis SHA: %w", err)
		}
		if googleapisSHA != cfg.Generate.GoogleapisRef {
			fmt.Printf("Updating googleapis to %s\n", googleapisSHA[:7])
			cfg.Generate.GoogleapisRef = googleapisSHA
			updated = true
		} else {
			fmt.Println("Googleapis is up to date")
		}
	}

	// Update discovery SHA if generate config exists
	if cfg.Librarian.Language != "" && cfg.Generate.DiscoveryRef != "" && updateDiscovery {
		discoverySHA, err := getLatestSHA("googleapis", "discovery-artifact-manager")
		if err != nil {
			return fmt.Errorf("failed to get latest discovery SHA: %w", err)
		}
		if discoverySHA != cfg.Generate.DiscoveryRef {
			fmt.Printf("Updating discovery to %s\n", discoverySHA[:7])
			cfg.Generate.DiscoveryRef = discoverySHA
			updated = true
		} else {
			fmt.Println("Discovery is up to date")
		}
	}

	if cfg.Librarian.Language != "" && updateGeneratorImage {
		// Dummy update for generator image
		fmt.Println("Generator image updated to latest.")
		updated = true
	}

	if updated {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		runYamlFmt(".librarian/config.yaml")
		fmt.Println("Configuration updated successfully")
	} else {
		fmt.Println("All versions are up to date")
	}

	return nil
}

func removeCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.StringArg("path")

	if path == "" {
		return fmt.Errorf("path is required")
	}

	if err := state.Remove(path); err != nil {
		return err
	}

	fmt.Printf("Removed artifact at %s\n", path)
	return nil
}

func editCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.StringArg("path")
	if path == "" {
		return fmt.Errorf("path is required")
	}

	keep := cmd.StringSlice("keep")
	remove := cmd.StringSlice("remove")
	exclude := cmd.StringSlice("exclude")
	languageFlags := cmd.StringSlice("language")

	// Load existing artifact
	artifact, err := state.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load artifact at %s: %w", path, err)
	}

	// Initialize config if needed
	if artifact.Config == nil {
		artifact.Config = &state.ConfigState{}
	}

	// Update config fields if flags were provided
	updated := false
	if len(keep) > 0 {
		artifact.Config.Keep = keep
		updated = true
		fmt.Printf("Set keep: %v\n", keep)
	}
	if len(remove) > 0 {
		artifact.Config.Remove = remove
		updated = true
		fmt.Printf("Set remove: %v\n", remove)
	}
	if len(exclude) > 0 {
		artifact.Config.Exclude = exclude
		updated = true
		fmt.Printf("Set exclude: %v\n", exclude)
	}

	// Update language-specific fields if flags were provided
	for _, flag := range languageFlags {
		lang, key, value, err := parseLanguageFlag(flag)
		if err != nil {
			return err
		}

		if artifact.Language == nil {
			artifact.Language = &state.LanguageState{}
		}

		switch lang {
		case "go":
			if artifact.Language.Go == nil {
				artifact.Language.Go = &state.GoLanguage{}
			}
			switch key {
			case "module":
				artifact.Language.Go.Module = value
				updated = true
				fmt.Printf("Set Go module: %s\n", value)
			default:
				return fmt.Errorf("unknown Go property: %s (expected 'module')", key)
			}
		case "python":
			if artifact.Language.Python == nil {
				artifact.Language.Python = &state.PythonLanguage{}
			}
			switch key {
			case "package":
				artifact.Language.Python.Package = value
				updated = true
				fmt.Printf("Set Python package: %s\n", value)
			default:
				return fmt.Errorf("unknown Python property: %s (expected 'package')", key)
			}
		case "rust":
			if artifact.Language.Rust == nil {
				artifact.Language.Rust = &state.RustLanguage{}
			}
			switch key {
			case "crate":
				artifact.Language.Rust.Crate = value
				updated = true
				fmt.Printf("Set Rust crate: %s\n", value)
			default:
				return fmt.Errorf("unknown Rust property: %s (expected 'crate')", key)
			}
		case "dart":
			if artifact.Language.Dart == nil {
				artifact.Language.Dart = &state.DartLanguage{}
			}
			switch key {
			case "package":
				artifact.Language.Dart.Package = value
				updated = true
				fmt.Printf("Set Dart package: %s\n", value)
			default:
				return fmt.Errorf("unknown Dart property: %s (expected 'package')", key)
			}
		default:
			return fmt.Errorf("unknown language: %s (expected go, python, rust, or dart)", lang)
		}
	}

	if !updated {
		// No flags provided, show current config
		fmt.Printf("Current configuration for %s:\n", path)
		hasConfig := false
		if artifact.Config != nil {
			if len(artifact.Config.Keep) > 0 {
				fmt.Printf("  Keep: %v\n", artifact.Config.Keep)
				hasConfig = true
			}
			if len(artifact.Config.Remove) > 0 {
				fmt.Printf("  Remove: %v\n", artifact.Config.Remove)
				hasConfig = true
			}
			if len(artifact.Config.Exclude) > 0 {
				fmt.Printf("  Exclude: %v\n", artifact.Config.Exclude)
				hasConfig = true
			}
		}
		if artifact.Language != nil {
			if artifact.Language.Go != nil && artifact.Language.Go.Module != "" {
				fmt.Printf("  Go module: %s\n", artifact.Language.Go.Module)
				hasConfig = true
			}
			if artifact.Language.Python != nil && artifact.Language.Python.Package != "" {
				fmt.Printf("  Python package: %s\n", artifact.Language.Python.Package)
				hasConfig = true
			}
			if artifact.Language.Rust != nil && artifact.Language.Rust.Crate != "" {
				fmt.Printf("  Rust crate: %s\n", artifact.Language.Rust.Crate)
				hasConfig = true
			}
			if artifact.Language.Dart != nil && artifact.Language.Dart.Package != "" {
				fmt.Printf("  Dart package: %s\n", artifact.Language.Dart.Package)
				hasConfig = true
			}
		}
		if !hasConfig {
			fmt.Println("  (no configuration set)")
		}
		return nil
	}

	// Save updated artifact
	if err := artifact.Save(path); err != nil {
		return fmt.Errorf("failed to save artifact state: %w", err)
	}
	runYamlFmt(filepath.Join(path, ".librarian.yaml"))

	fmt.Printf("Updated configuration for %s\n", path)
	return nil
}

func prepareCommand(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")
	path := cmd.StringArg("path")

	if !all && cmd.NArg() == 0 {
		return fmt.Errorf("either --all flag or path is required")
	}

	if all {
		artifacts, err := state.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load artifacts: %w", err)
		}

		fmt.Printf("Preparing all %d artifacts for release...\n", len(artifacts))
		for path, artifact := range artifacts {
			if artifact.Release == nil {
				continue
			}
			fmt.Printf("  - Preparing %s\n", path)
			if err := prepareRelease(artifact); err != nil {
				return fmt.Errorf("failed to prepare release for %s: %w", path, err)
			}
			if err := artifact.Save(path); err != nil {
				return fmt.Errorf("failed to save artifact state for %s: %w", path, err)
			}
			runYamlFmt(filepath.Join(path, ".librarian.yaml"))
		}
	} else {
		artifact, err := state.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load artifact at %s: %w", path, err)
		}
		if artifact.Release == nil {
			return fmt.Errorf("artifact at %s is not configured for release", path)
		}
		fmt.Printf("Preparing artifact at %s for release...\n", path)
		if err := prepareRelease(artifact); err != nil {
			return fmt.Errorf("failed to prepare release for %s: %w", path, err)
		}
		if err := artifact.Save(path); err != nil {
			return fmt.Errorf("failed to save artifact state for %s: %w", path, err)
		}
		runYamlFmt(filepath.Join(path, ".librarian.yaml"))
	}

	fmt.Println("Prepare complete")
	return nil
}

func prepareRelease(artifact *state.Artifact) error {
	nextVersion, err := bumpPatch(artifact.Release.Version)
	if err != nil {
		return err
	}
	commit, err := getCurrentCommit()
	if err != nil {
		return err
	}
	artifact.Release.Prepared = &state.ReleaseInfo{
		Tag:    nextVersion,
		Commit: commit,
	}
	return nil
}

func bumpPatch(version string) (string, error) {
	if version == "null" {
		return "v0.1.0", nil
	}
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}
	patch, err := Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid patch version: %s", parts[2])
	}
	return fmt.Sprintf("v%s.%s.%d", parts[0], parts[1], patch+1), nil
}

func getCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func Atoi(s string) (int, error) {
	i := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid digit: %c", r)
		}
		i = i*10 + int(r-'0')
	}
	return i, nil
}

func releaseCommand(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")
	path := cmd.StringArg("path")

	if !all && cmd.NArg() == 0 {
		return fmt.Errorf("either --all flag or path is required")
	}

	if all {
		artifacts, err := state.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load artifacts: %w", err)
		}

		var tagged bool
		for path, artifact := range artifacts {
			if artifact.Release != nil && artifact.Release.Prepared != nil {
				fmt.Printf("Releasing %s %s...\n", path, artifact.Release.Prepared.Tag)
				if err := createGitTag(artifact.Release.Prepared.Tag, artifact.Release.Prepared.Commit); err != nil {
					return fmt.Errorf("failed to create git tag for %s: %w", path, err)
				}
				fmt.Println("  - Creating git tag...")

			artifact.Release.Version = artifact.Release.Prepared.Tag
			artifact.Release.Prepared = nil
			tagged = true

			if err := artifact.Save(path); err != nil {
				return fmt.Errorf("failed to save artifact state: %w", err)
			}
			runYamlFmt(filepath.Join(path, ".librarian.yaml"))
			fmt.Println("  - Done.")
			}
		}

		if !tagged {
			fmt.Println("No artifacts to release.")
			return nil
		}

		fmt.Println("Release complete.")
		return nil
	}

	if path == "" {
		return fmt.Errorf("path is required (or use --all)")
	}

	artifact, err := state.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load artifact at %s: %w", path, err)
	}

	if artifact.Release == nil || artifact.Release.Prepared == nil {
		return fmt.Errorf("no release prepared for artifact at %s", path)
	}

	fmt.Printf("Releasing %s %s...\n", path, artifact.Release.Prepared.Tag)
	if err := createGitTag(artifact.Release.Prepared.Tag, artifact.Release.Prepared.Commit); err != nil {
		return fmt.Errorf("failed to create git tag for %s: %w", path, err)
	}
	fmt.Println("  - Creating git tag...")

	artifact.Release.Version = artifact.Release.Prepared.Tag
	artifact.Release.Prepared = nil

	if err := artifact.Save(path); err != nil {
		return fmt.Errorf("failed to save artifact state: %w", err)
	}
	runYamlFmt(filepath.Join(path, ".librarian.yaml"))
	fmt.Println("  - Done.")

	fmt.Println("Release complete.")
	return nil
}

func createGitTag(tag, commit string) error {
	cmd := exec.Command("git", "tag", tag, commit)
	return cmd.Run()
}

func listCommand(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Listing all tracked directories:")
	artifacts, err := state.LoadAll()
	if err != nil {
		return err
	}
	for path := range artifacts {
		fmt.Printf("- %s\n", path)
	}
	return nil
}

func statusCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.StringArg("path")
	if path == "" {
		return fmt.Errorf("path is required")
	}
	artifact, err := state.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load artifact at %s: %w", path, err)
	}

	fmt.Printf("Status for %s:\n", path)
	if artifact.Generate != nil {
		fmt.Println("  Generation:")
		fmt.Printf("    Commit: %s\n", artifact.Generate.Commit)
		fmt.Printf("    Librarian version: %s\n", artifact.Generate.Librarian)
	}
	if artifact.Release != nil {
		fmt.Println("  Release:")
		fmt.Printf("    Version: %s\n", artifact.Release.Version)
		if artifact.Release.Prepared != nil {
			fmt.Printf("    Prepared: %s\n", artifact.Release.Prepared.Tag)
		}
	}
	return nil
}

func historyCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.StringArg("path")
	if path == "" {
		return fmt.Errorf("path is required")
	}
	fmt.Printf("Showing release history for %s:\n", path)


librarianFile := filepath.Join(path, ".librarian.yaml")
	c := exec.Command("git", "log", "--pretty=format:%h %ad | %s", "--date=short", "--", librarianFile)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// parseLanguageFlag parses a string in the format "LANG:KEY=VALUE" and returns the language, key, and value.
func parseLanguageFlag(s string) (lang, key, value string, err error) {
	// Split on first ':'
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("expected format LANG:KEY=VALUE, got %q", s)
	}
	lang = parts[0]

	// Split on first '='
	kvParts := strings.SplitN(parts[1], "=", 2)
	if len(kvParts) != 2 {
		return "", "", "", fmt.Errorf("expected format LANG:KEY=VALUE, got %q", s)
	}
	key, value = kvParts[0], kvParts[1]

	return lang, key, value, nil
}

// getLatestSHA fetches the latest commit SHA for the given repo in the given
// org.
func getLatestSHA(org, repo string) (string, error) {
	repoURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", org, repo)
	resp, err := http.Get(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to get repo info: %w", err)
	}
	defer resp.Body.Close()
	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return "", fmt.Errorf("failed to decode repo info: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", org, repo, repoInfo.DefaultBranch)
	resp, err = http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get latest commit: %w", err)
	}
	defer resp.Body.Close()

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return commit.SHA, nil
}

// getLibrarianVersion returns the latest version of librarian.
func getLibrarianVersion() (string, error) {
	// Dummy version for prototype
	return "v0.1.0-dummy", nil
}

// runYamlFmt runs yamlfmt on the given file.
func runYamlFmt(file string) {
	cmd := exec.Command("yamlfmt", file)
	if err := cmd.Run(); err != nil {
		// If yamlfmt is not installed, we don't want to fail.
		// We'll just log a warning.
		log.Printf("failed to run yamlfmt on %s: %v", file, err)
	}
}
