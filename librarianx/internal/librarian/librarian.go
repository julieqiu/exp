package librarian

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/julieqiu/xlibrarian/internal/config"
	"github.com/urfave/cli/v3"
)

// Sentinel errors for validation.
var (
	errLanguageRequired      = errors.New("language argument required (go, python, rust)")
	errArtifactPathRequired  = errors.New("artifact path required")
	errArtifactOrAllRequired = errors.New("artifact path required (or use --all)")
	errUpdateFlagRequired    = errors.New("one of --all, --googleapis, or --discovery required")
	errShaWithAll            = errors.New("--sha cannot be used with --all")
)

// Run executes the librarian command with the given arguments.
func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:    "librarianx",
		Usage:   "manage Google Cloud client libraries",
		Version: "v0.1.0",
		Commands: []*cli.Command{
			initCommand(),
			installCommand(),
			newCommand(),
			generateCommand(),
			testCommand(),
			updateCommand(),
			releaseCommand(),
		},
	}

	return cmd.Run(ctx, args)
}

// initCommand creates a new repository configuration.
func initCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "initialize a repository configuration",
		ArgsUsage: "<language>",
		Description: `Initialize a repository configuration for the specified language.

   Creates .librarian/config.yaml with default settings for the language.

   Supported languages: go, python, rust

   Example:
     librarianx init go
     librarianx init python`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() < 1 {
				return errLanguageRequired
			}
			language := cmd.Args().Get(0)
			return runInit(ctx, language)
		},
	}
}

// installCommand installs language-specific generator dependencies.
func installCommand() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "install language-specific generator dependencies",
		ArgsUsage: "<language>",
		Description: `Install generator dependencies for the specified language.

   Can install dependencies locally or use a Docker container.

   Example:
     librarianx install go
     librarianx install python --use-container`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "use-container",
				Usage: "use Docker container for code generation",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() < 1 {
				return errLanguageRequired
			}
			language := cmd.Args().Get(0)
			useContainer := cmd.Bool("use-container")
			return runInstall(ctx, language, useContainer)
		},
	}
}

// newCommand creates a new library and generates code.
func newCommand() *cli.Command {
	return &cli.Command{
		Name:      "new",
		Usage:     "create a new library and generate code",
		ArgsUsage: "<artifact-path> [api-paths...]",
		Description: `Create a new library and generate code in one step.

   For generated libraries: Parses BUILD.bazel files, creates configuration, and generates code.
   For handwritten libraries: Creates release-only configuration (no API paths needed).

   This combines the previous 'add' and 'generate' steps into one command.

   Examples:
     # Create new generated library (Go)
     librarianx new secretmanager google/cloud/secretmanager/v1

     # Create library with multiple API versions
     librarianx new secretmanager google/cloud/secretmanager/v1 google/cloud/secretmanager/v1beta2

     # Create handwritten library (no generation)
     librarianx new custom-tool

     # Create new generated library (Python)
     librarianx new google-cloud-secret-manager google/cloud/secretmanager/v1`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() < 1 {
				return errArtifactPathRequired
			}
			artifactPath := cmd.Args().Get(0)
			apiPaths := cmd.Args().Slice()[1:]
			return runNew(ctx, artifactPath, apiPaths)
		},
	}
}

// generateCommand generates code for an artifact.
func generateCommand() *cli.Command {
	return &cli.Command{
		Name:      "generate",
		Usage:     "regenerate code for an existing artifact",
		ArgsUsage: "[artifact-path]",
		Description: `Regenerate code for an existing artifact or all artifacts.

   Uses the configuration in .librarian.yaml to run code generation.
   For new libraries, use 'librarianx new' instead.

   Examples:
     # Regenerate specific artifact
     librarianx generate secretmanager

     # Regenerate all artifacts
     librarianx generate --all`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "regenerate all artifacts in the repository",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			all := cmd.Bool("all")
			if all {
				return runGenerateAll(ctx)
			}
			if cmd.NArg() < 1 {
				return errArtifactOrAllRequired
			}
			artifactPath := cmd.Args().Get(0)
			return runGenerate(ctx, artifactPath)
		},
	}
}

// testCommand runs tests for an artifact.
func testCommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "run tests for an artifact",
		ArgsUsage: "<artifact-path>",
		Description: `Run language-specific tests for an artifact.

   Examples:
     # Run tests for Go library
     librarianx test secretmanager

     # Run tests for Python library
     librarianx test google-cloud-secret-manager

     # Run tests for all artifacts
     librarianx test --all`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "run tests for all artifacts in the repository",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			all := cmd.Bool("all")
			if all {
				return runTestAll(ctx)
			}
			if cmd.NArg() < 1 {
				return errArtifactOrAllRequired
			}
			artifactPath := cmd.Args().Get(0)
			return runTest(ctx, artifactPath)
		},
	}
}

// updateCommand updates source references (googleapis/discovery).
func updateCommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "update source references to latest versions",
		Description: `Update googleapis or discovery source references to latest versions.

   This updates the sources section in .librarian/config.yaml.
   After updating sources, run 'librarianx generate --all' to regenerate libraries.

   Examples:
     # Update both googleapis and discovery sources
     librarianx update --all

     # Update only googleapis source
     librarianx update --googleapis

     # Update only discovery source
     librarianx update --discovery

     # Pin to specific commit
     librarianx update --googleapis --sha abc123def456`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "update all sources (googleapis and discovery)",
			},
			&cli.BoolFlag{
				Name:  "googleapis",
				Usage: "update googleapis source",
			},
			&cli.BoolFlag{
				Name:  "discovery",
				Usage: "update discovery source",
			},
			&cli.StringFlag{
				Name:  "sha",
				Usage: "pin to specific commit SHA (only with --googleapis or --discovery)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			all := cmd.Bool("all")
			googleapis := cmd.Bool("googleapis")
			discovery := cmd.Bool("discovery")
			sha := cmd.String("sha")

			if !all && !googleapis && !discovery {
				return errUpdateFlagRequired
			}

			if sha != "" && all {
				return errShaWithAll
			}

			return runUpdate(ctx, all, googleapis, discovery, sha)
		},
	}
}

// releaseCommand releases libraries (version bump, tag, publish).
func releaseCommand() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "release libraries (dry-run by default)",
		ArgsUsage: "[artifact-path]",
		Description: `Release one or more libraries by bumping versions, creating tags, and publishing.

   By default, runs in DRY-RUN mode to show what would happen without making changes.
   Use --execute to actually perform the release.

   This command does everything:
   1. Analyze conventional commits to determine version bump
   2. Update version files and changelogs
   3. Create git commit
   4. Create and push git tags
   5. Publish to package registries (PyPI, crates.io, pkg.go.dev)

   Examples:
     # Dry-run: show what would be released
     librarianx release secretmanager

     # Actually release a specific library
     librarianx release secretmanager --execute

     # Release all changed libraries (dry-run)
     librarianx release --all

     # Release all changed libraries (execute)
     librarianx release --all --execute

     # Skip tests during release
     librarianx release secretmanager --execute --skip-tests

     # Skip publishing to registries (only tag)
     librarianx release secretmanager --execute --skip-publish`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "release all libraries with pending changes",
			},
			&cli.BoolFlag{
				Name:  "execute",
				Usage: "actually perform the release (default is dry-run)",
			},
			&cli.BoolFlag{
				Name:  "skip-tests",
				Usage: "skip running tests before release",
			},
			&cli.BoolFlag{
				Name:  "skip-publish",
				Usage: "create tags but don't publish to package registries",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			all := cmd.Bool("all")
			execute := cmd.Bool("execute")
			skipTests := cmd.Bool("skip-tests")
			skipPublish := cmd.Bool("skip-publish")

			var artifactPath string
			if !all {
				if cmd.NArg() < 1 {
					return errArtifactOrAllRequired
				}
				artifactPath = cmd.Args().Get(0)
			}

			return runRelease(ctx, artifactPath, all, execute, skipTests, skipPublish)
		},
	}
}

// Placeholder implementations for each command.
// These will be implemented in separate files.

func runInit(ctx context.Context, language string) error {
	// Check if librarian.yaml already exists
	const configPath = "librarian.yaml"
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("librarian.yaml already exists in current directory")
	}

	// Create default config based on language
	cfg := createDefaultConfig(language)

	// Write config to librarian.yaml
	if err := cfg.Write(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Created librarian.yaml\n")
	return nil
}

func createDefaultConfig(language string) *config.Config {
	cfg := &config.Config{
		Version:  "v0.1.0",
		Language: language,
		Release: &config.Release{
			TagFormat: "{id}/v{version}",
		},
	}

	// Add language-specific defaults
	switch language {
	case "go":
		cfg.Container = &config.Container{
			Image: "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/go-librarian-generator",
			Tag:   "latest",
		}
		cfg.Sources = config.Sources{
			Googleapis: &config.Source{
				URL:    "https://github.com/googleapis/googleapis/archive/9fcfbea0aa5b50fa22e190faceb073d74504172b.tar.gz",
				SHA256: "81e6057ffd85154af5268c2c3c8f2408745ca0f7fa03d43c68f4847f31eb5f98",
			},
		}
		cfg.Generate = &config.Generate{
			OutputDir: "./",
			Defaults: &config.GenerateDefaults{
				Transport:        "grpc+rest",
				RestNumericEnums: boolPtr(true),
				ReleaseLevel:     "stable",
			},
		}

	case "python":
		cfg.Container = &config.Container{
			Image: "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator",
			Tag:   "latest",
		}
		cfg.Sources = config.Sources{
			Googleapis: &config.Source{
				URL:    "https://github.com/googleapis/googleapis/archive/9fcfbea0aa5b50fa22e190faceb073d74504172b.tar.gz",
				SHA256: "81e6057ffd85154af5268c2c3c8f2408745ca0f7fa03d43c68f4847f31eb5f98",
			},
		}
		cfg.Generate = &config.Generate{
			OutputDir: "packages/",
			Defaults: &config.GenerateDefaults{
				Transport:        "grpc+rest",
				RestNumericEnums: boolPtr(true),
				ReleaseLevel:     "stable",
			},
		}
		cfg.Release.TagFormat = "{name}/v{version}"

	case "rust":
		cfg.Container = &config.Container{
			Image: "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/rust-librarian-generator",
			Tag:   "latest",
		}
		cfg.Sources = config.Sources{
			Googleapis: &config.Source{
				URL:    "https://github.com/googleapis/googleapis/archive/9fcfbea0aa5b50fa22e190faceb073d74504172b.tar.gz",
				SHA256: "81e6057ffd85154af5268c2c3c8f2408745ca0f7fa03d43c68f4847f31eb5f98",
			},
		}
		cfg.Generate = &config.Generate{
			OutputDir: "generated/",
			Defaults: &config.GenerateDefaults{
				Transport:    "grpc+rest",
				ReleaseLevel: "preview",
			},
		}
		cfg.Release.TagFormat = "{name}/v{version}"
	}

	return cfg
}

func boolPtr(b bool) *bool {
	return &b
}

func runInstall(ctx context.Context, language string, useContainer bool) error {
	return fmt.Errorf("install command not yet implemented for language: %s (container: %v)", language, useContainer)
}

func runNew(ctx context.Context, artifactPath string, apiPaths []string) error {
	return fmt.Errorf("new command not yet implemented for artifact: %s with APIs: %v", artifactPath, apiPaths)
}

func runGenerate(ctx context.Context, artifactPath string) error {
	return fmt.Errorf("generate command not yet implemented for artifact: %s", artifactPath)
}

func runGenerateAll(ctx context.Context) error {
	return fmt.Errorf("generate --all command not yet implemented")
}

func runTest(ctx context.Context, artifactPath string) error {
	return fmt.Errorf("test command not yet implemented for artifact: %s", artifactPath)
}

func runTestAll(ctx context.Context) error {
	return fmt.Errorf("test --all command not yet implemented")
}

func runUpdate(ctx context.Context, all, googleapis, discovery bool, sha string) error {
	return fmt.Errorf("update command not yet implemented (all: %v, googleapis: %v, discovery: %v, sha: %s)", all, googleapis, discovery, sha)
}

func runRelease(ctx context.Context, artifactPath string, all, execute, skipTests, skipPublish bool) error {
	if !execute {
		return fmt.Errorf("release command not yet implemented (DRY-RUN mode - artifact: %s, all: %v)", artifactPath, all)
	}
	return fmt.Errorf("release command not yet implemented (EXECUTE mode - artifact: %s, all: %v, skip-tests: %v, skip-publish: %v)", artifactPath, all, skipTests, skipPublish)
}
