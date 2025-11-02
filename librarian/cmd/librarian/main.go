package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/julieqiu/exp/librarian/internal/config"
	"github.com/julieqiu/exp/librarian/internal/state"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "librarian",
		Usage: "Manage release artifacts",
		Commands: []*cli.Command{
			{
				Name:      "init",
				Usage:     "Initialize a new librarian-managed repository",
				Arguments: []cli.Argument{&cli.StringArg{Name: "language"}},
				Action:    initCommand,
			},
			{
				Name:  "config",
				Usage: "Manage configuration",
				Commands: []*cli.Command{
					{
						Name:  "set",
						Usage: "Set a configuration value",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "sync",
								Usage: "Regenerate all artifacts after setting config",
							},
						},
						Arguments: []cli.Argument{
							&cli.StringArg{Name: "key"},
							&cli.StringArg{Name: "value"},
						},
						Action: configSetCommand,
					},
					{
						Name:  "update",
						Usage: "Update librarian and container versions",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "sync",
								Usage: "Regenerate all artifacts after update",
							},
						},
						Action: configUpdateCommand,
					},
				},
			},
			{
				Name:  "generate",
				Usage: "Generate client libraries",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "Regenerate all artifacts",
					},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "artifact-path"},
					&cli.StringArg{Name: "api-path"},
				},
				Action: generateCommand,
			},
			{
				Name:      "remove",
				Usage:     "Remove an artifact from librarian management",
				Arguments: []cli.Argument{&cli.StringArg{Name: "artifact-path"}},
				Action:    removeCommand,
			},
			{
				Name:  "release",
				Usage: "Manage releases",
				Commands: []*cli.Command{
					{
						Name:  "prepare",
						Usage: "Prepare artifacts for release",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all",
								Usage: "Prepare all artifacts",
							},
						},
						Arguments: []cli.Argument{&cli.StringArg{Name: "artifact-path"}},
						Action:    releasePrepareCommand,
					},
					{
						Name:  "tag",
						Usage: "Tag artifacts for release",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all",
								Usage: "Tag all artifacts",
							},
						},
						Arguments: []cli.Argument{&cli.StringArg{Name: "artifact-path"}},
						Action:    releaseTagCommand,
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func initCommand(ctx context.Context, cmd *cli.Command) error {
	mode := cmd.StringArg("language")
	if mode == "" {
		return fmt.Errorf("mode is required")
	}
	if mode != "python" && mode != "go" && mode != "release-only" {
		return fmt.Errorf("mode must be python, go, or release-only")
	}

	librarianVersion, err := getLibrarianVersion()
	if err != nil {
		return err
	}

	cfg := &config.Config{
		Librarian: config.LibrarianConfig{
			Version: librarianVersion,
			Mode:    mode,
		},
		Release: config.ReleaseConfig{
			TagFormat: "{package}-v{version}",
		},
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	runYamlFmt(".librarian/config.yaml")

	fmt.Printf("Initialized librarian repository for %s\n", mode)
	fmt.Println("Created .librarian/config.yaml")
	return nil
}

// ensureGenerationConfig initializes generation-related config fields if they're not set.
func ensureGenerationConfig(cfg *config.Config) error {
	var updated bool

	// Initialize generator image if not set
	if cfg.Generate.Image == "" {
		if cfg.Librarian.Mode == "python" {
			cfg.Generate.Image = "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator:latest"
		} else if cfg.Librarian.Mode == "go" {
			cfg.Generate.Image = "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/go-librarian-generator:latest"
		}
		updated = true
	}

	// Initialize googleapis SHA if not set
	if cfg.Generate.Googleapis == "" {
		googleapisSHA, err := getLatestSHA("googleapis", "googleapis")
		if err != nil {
			return fmt.Errorf("failed to get latest googleapis SHA: %w", err)
		}
		cfg.Generate.Googleapis = googleapisSHA
		updated = true
	}

	// Initialize discovery SHA if not set
	if cfg.Generate.Discovery == "" {
		discoverySHA, err := getLatestSHA("googleapis", "discovery-artifact-manager")
		if err != nil {
			return fmt.Errorf("failed to get latest discovery SHA: %w", err)
		}
		cfg.Generate.Discovery = discoverySHA
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
	artifactPath := cmd.StringArg("artifact-path")
	apiPath := cmd.StringArg("api-path")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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
			fmt.Printf("  - Regenerating %s\n", path)

			// Sync artifact state with current config
			if artifact.Generate != nil {
				artifact.Generate.Librarian = cfg.Librarian.Version
				artifact.Generate.Image = cfg.GeneratorImage()
				artifact.Generate.GoogleapisSHA = cfg.Generate.Googleapis
				artifact.Generate.DiscoverySHA = cfg.Generate.Discovery

				if err := artifact.Save(path); err != nil {
					return fmt.Errorf("failed to save artifact state: %w", err)
				}
				runYamlFmt(filepath.Join(path, ".librarian.yaml"))
			}

			// TODO: Run generator for each artifact
		}
		fmt.Println("Generation complete")
		return nil
	}

	if artifactPath == "" {
		return fmt.Errorf("artifact-path is required (or use --all)")
	}

	// Check if artifact exists
	artifact, err := state.Load(artifactPath)
	if err != nil {
		return fmt.Errorf("failed to load artifact: %w", err)
	}

	// If apiPath is provided, this is a new artifact
	if apiPath != "" {
		artifact = &state.Artifact{
			Generate: &state.GenerateState{
				APIs: []state.API{
					{Path: apiPath},
				},
				Commit:        "c288189b43c016dd3cf1ec73ce3cadee8b732f07", // Dummy value
				Librarian:     cfg.Librarian.Version,
				Image:         cfg.GeneratorImage(),
				GoogleapisSHA: cfg.Generate.Googleapis,
				DiscoverySHA:  cfg.Generate.Discovery,
			},
		}

		fmt.Printf("Created artifact at %s with API %s\n", artifactPath, apiPath)
	} else {
		// Regenerating existing artifact - sync state with current config
		fmt.Printf("Regenerating artifact at %s...\n", artifactPath)

		if artifact.Generate != nil {
			artifact.Generate.Librarian = cfg.Librarian.Version
			artifact.Generate.Image = cfg.GeneratorImage()
			artifact.Generate.GoogleapisSHA = cfg.Generate.Googleapis
			artifact.Generate.DiscoverySHA = cfg.Generate.Discovery
		}
	}

	// Save artifact state
	if err := artifact.Save(artifactPath); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(filepath.Join(artifactPath, ".librarian.yaml"))

	fmt.Println("Running generator...")
	// TODO: Actually run the generator container
	fmt.Println("Generation complete")
	return nil
}

func configSetCommand(ctx context.Context, cmd *cli.Command) error {
	sync := cmd.Bool("sync")
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

	if sync {
		fmt.Println("\nRegenerating all artifacts...")
		artifacts, err := state.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load artifacts: %w", err)
		}

		if len(artifacts) == 0 {
			fmt.Println("No artifacts to regenerate")
			return nil
		}

		fmt.Printf("Regenerating all %d artifacts...\n", len(artifacts))
		for path, artifact := range artifacts {
			fmt.Printf("  - Regenerating %s\n", path)

			// Update artifact state with current config values
			if artifact.Generate != nil {
				artifact.Generate.Librarian = cfg.Librarian.Version
				artifact.Generate.Image = cfg.GeneratorImage()
				artifact.Generate.GoogleapisSHA = cfg.Generate.Googleapis
				artifact.Generate.DiscoverySHA = cfg.Generate.Discovery

				if err := artifact.Save(path); err != nil {
					return fmt.Errorf("failed to save artifact state: %w", err)
				}
				runYamlFmt(filepath.Join(path, ".librarian.yaml"))
			}

			// TODO: Run generator for the artifact
		}
		fmt.Println("Regeneration complete")
	}

	return nil
}

func configUpdateCommand(ctx context.Context, cmd *cli.Command) error {
	sync := cmd.Bool("sync")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Checking for updates...")
	var updated bool

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
	if cfg.Generate.Googleapis != "" {
		googleapisSHA, err := getLatestSHA("googleapis", "googleapis")
		if err != nil {
			return fmt.Errorf("failed to get latest googleapis SHA: %w", err)
		}
		if googleapisSHA != cfg.Generate.Googleapis {
			fmt.Printf("Updating googleapis to %s\n", googleapisSHA[:7])
			cfg.Generate.Googleapis = googleapisSHA
			updated = true
		} else {
			fmt.Println("Googleapis is up to date")
		}
	}

	// Update discovery SHA if generate config exists
	if cfg.Generate.Discovery != "" {
		discoverySHA, err := getLatestSHA("googleapis", "discovery-artifact-manager")
		if err != nil {
			return fmt.Errorf("failed to get latest discovery SHA: %w", err)
		}
		if discoverySHA != cfg.Generate.Discovery {
			fmt.Printf("Updating discovery to %s\n", discoverySHA[:7])
			cfg.Generate.Discovery = discoverySHA
			updated = true
		} else {
			fmt.Println("Discovery is up to date")
		}
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

	if sync {
		fmt.Println("\nRegenerating all artifacts...")
		artifacts, err := state.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load artifacts: %w", err)
		}

		if len(artifacts) == 0 {
			fmt.Println("No artifacts to regenerate")
			return nil
		}

		fmt.Printf("Regenerating all %d artifacts...\n", len(artifacts))
		for path, artifact := range artifacts {
			fmt.Printf("  - Regenerating %s\n", path)

			// Update artifact state with current config values
			if artifact.Generate != nil {
				artifact.Generate.Librarian = cfg.Librarian.Version
				artifact.Generate.Image = cfg.GeneratorImage()
				artifact.Generate.GoogleapisSHA = cfg.Generate.Googleapis
				artifact.Generate.DiscoverySHA = cfg.Generate.Discovery

				if err := artifact.Save(path); err != nil {
					return fmt.Errorf("failed to save artifact state: %w", err)
				}
				runYamlFmt(filepath.Join(path, ".librarian.yaml"))
			}

			// TODO: Run generator for the artifact
		}
		fmt.Println("Regeneration complete")
	}

	return nil
}

func removeCommand(ctx context.Context, cmd *cli.Command) error {
	artifactPath := cmd.StringArg("artifact-path")

	if artifactPath == "" {
		return fmt.Errorf("artifact-path is required")
	}

	if err := state.Remove(artifactPath); err != nil {
		return err
	}

	fmt.Printf("Removed artifact at %s\n", artifactPath)
	return nil
}

func releasePrepareCommand(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")
	artifactPath := cmd.StringArg("artifact-path")

	if !all && artifactPath == "" {
		return fmt.Errorf("either --all flag or artifact-path is required")
	}

	if all {
		artifacts, err := state.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load artifacts: %w", err)
		}

		fmt.Printf("Preparing all %d artifacts for release...\n", len(artifacts))
		for path := range artifacts {
			fmt.Printf("  - Preparing %s\n", path)
			// TODO: Create release metadata and CHANGELOG.md
		}
	} else {
		_, err := state.Load(artifactPath)
		if err != nil {
			return fmt.Errorf("failed to load artifact at %s: %w", artifactPath, err)
		}
		fmt.Printf("Preparing artifact at %s for release...\n", artifactPath)
		// TODO: Create release metadata and CHANGELOG.md
	}

	fmt.Println("Prepare complete")
	return nil
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
	out, err := exec.Command("go", "list", "-m", "-json", "github.com/googleapis/librarian@latest").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get librarian version: %w", err)
	}
	var mod struct {
		Version string `json:"Version"`
	}
	if err := json.Unmarshal(out, &mod); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return mod.Version, nil
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

func releaseTagCommand(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")
	artifactPath := cmd.StringArg("artifact-path")

	if all {
		artifacts, err := state.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load artifacts: %w", err)
		}

		var tagged bool
		for path, artifact := range artifacts {
			if artifact.Release != nil && artifact.Release.NextReleaseAt != nil && artifact.Release.NextReleaseAt.Tag != "" {
				if artifact.Release.LastReleasedAt == nil || artifact.Release.NextReleaseAt.Tag != artifact.Release.LastReleasedAt.Tag {
					fmt.Printf("Tagging %s %s...\n", path, artifact.Release.NextReleaseAt.Tag)
					// TODO: Implement actual git tagging logic
					fmt.Println("  - Creating git tag...")

					artifact.Release.LastReleasedAt = artifact.Release.NextReleaseAt
					artifact.Release.NextReleaseAt = nil
					tagged = true

					if err := artifact.Save(path); err != nil {
						return fmt.Errorf("failed to save artifact state: %w", err)
					}
					runYamlFmt(filepath.Join(path, ".librarian.yaml"))
					fmt.Println("  - Done.")
				}
			}
		}

		if !tagged {
			fmt.Println("No artifacts to tag.")
			return nil
		}

		fmt.Println("Tag complete.")
		return nil
	}

	if artifactPath == "" {
		return fmt.Errorf("artifact-path is required (or use --all)")
	}

	artifact, err := state.Load(artifactPath)
	if err != nil {
		return fmt.Errorf("failed to load artifact at %s: %w", artifactPath, err)
	}

	if artifact.Release == nil || artifact.Release.NextReleaseAt == nil {
		return fmt.Errorf("no release prepared for artifact at %s", artifactPath)
	}

	if artifact.Release.LastReleasedAt != nil && artifact.Release.NextReleaseAt.Tag == artifact.Release.LastReleasedAt.Tag {
		fmt.Printf("Artifact at %s already released at %s\n", artifactPath, artifact.Release.LastReleasedAt.Tag)
		return nil
	}

	fmt.Printf("Tagging %s %s...\n", artifactPath, artifact.Release.NextReleaseAt.Tag)
	// TODO: Implement actual git tagging logic
	fmt.Println("  - Creating git tag...")

	artifact.Release.LastReleasedAt = artifact.Release.NextReleaseAt
	artifact.Release.NextReleaseAt = nil

	if err := artifact.Save(artifactPath); err != nil {
		return fmt.Errorf("failed to save artifact state: %w", err)
	}
	runYamlFmt(filepath.Join(artifactPath, ".librarian.yaml"))
	fmt.Println("  - Done.")

	fmt.Println("Tag complete.")
	return nil
}
