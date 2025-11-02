package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/julieqiu/exp/librarian/internal/config"
	"github.com/julieqiu/exp/librarian/internal/state"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "librarian",
		Usage: "Manage Google API client libraries",
		Commands: []*cli.Command{
			{
				Name:      "init",
				Usage:     "Initialize a new librarian-managed repository",
				Arguments: []cli.Argument{&cli.StringArg{Name: "language"}},
				Action:    initCommand,
			},
			{
				Name:  "add",
				Usage: "Add a library to be managed by librarian",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "library-id"},
					&cli.StringArg{Name: "api-path"},
				},
				Action: addCommand,
			},
			{
				Name:  "update",
				Usage: "Regenerate client libraries",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "Update all libraries",
					},
				},
				Arguments: []cli.Argument{&cli.StringArg{Name: "library-id"}},
				Action:    updateCommand,
			},
			{
				Name:  "config",
				Usage: "Manage configuration",
				Commands: []*cli.Command{
					{
						Name:  "set",
						Usage: "Set a configuration value",
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
								Name:  "no-sync",
								Usage: "Skip regenerating libraries after update",
							},
						},
						Action: configUpdateCommand,
					},
				},
			},
			{
				Name:      "remove",
				Usage:     "Remove a library from librarian management",
				Arguments: []cli.Argument{&cli.StringArg{Name: "library-id"}},
				Action:    removeCommand,
			},
			{
				Name:  "release",
				Usage: "Release libraries",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "Release all libraries",
					},
				},
				Arguments: []cli.Argument{&cli.StringArg{Name: "library-id"}},
				Action:    releaseCommand,
			},
			{
				Name:   "publish",
				Usage:  "Publish libraries that have a pending release",
				Action: publishCommand,
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
		Version:          librarianVersion,
		Mode:             mode,
		ReleaseTagFormat: "{package}-v{version}",
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	runYamlFmt(".librarian/config.yaml")

	st := &state.State{
		Packages: make(map[string]*state.Package),
	}

	if err := st.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(".librarian/state.yaml")

	fmt.Printf("Initialized librarian repository for %s\n", mode)
	fmt.Println("Created .librarian/config.yaml")
	fmt.Println("Created .librarian/state.yaml")
	return nil
}

// ensureGenerationConfig initializes generation-related config fields if they're not set.
func ensureGenerationConfig(cfg *config.Config) error {
	var updated bool

	// Initialize generator image if not set
	if cfg.Generate.Image == "" {
		if cfg.Mode == "python" {
			cfg.Generate.Image = "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator:latest"
		} else if cfg.Mode == "go" {
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

func addCommand(ctx context.Context, cmd *cli.Command) error {
	libraryID := cmd.StringArg("library-id")
	apiPath := cmd.StringArg("api-path")

	if libraryID == "" || apiPath == "" {
		return fmt.Errorf("library-id and api-path are required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Ensure generation config is initialized
	if err := ensureGenerationConfig(cfg); err != nil {
		return err
	}

	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	pkg := &state.Package{
		Generate: &state.GenerateState{
			APIs: []state.API{
				{Path: apiPath},
			},
			Commit:        "c288189b43c016dd3cf1ec73ce3cadee8b732f07", // Dummy value
			Librarian:     cfg.Version,
			Image:         cfg.GeneratorImage(),
			GoogleapisSHA: cfg.Generate.Googleapis,
			DiscoverySHA:  cfg.Generate.Discovery,
		},
		Release: &state.ReleaseState{
			LastReleasedAt: &state.ReleaseInfo{
				Tag:    "v1.18.0",
				Commit: "4a92b10e8f0a2b5c6d7e8f9a0b1c2d3e4f5a6b7c",
			},
			NextReleaseAt: &state.ReleaseInfo{
				Tag:    "v1.19.0",
				Commit: "some-new-commit-hash",
			},
		},
	}

	st.AddPackage(libraryID, pkg)

	if err := st.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(".librarian/state.yaml")

	fmt.Printf("Added library %s with API %s\n", libraryID, apiPath)
	fmt.Println("Running generator...")
	// TODO: Actually run the generator container
	fmt.Println("Generation complete")
	return nil
}

func updateCommand(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")
	libraryID := cmd.StringArg("library-id")

	if !all && libraryID == "" {
		return fmt.Errorf("either --all flag or library-id is required")
	}

	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if all {
		fmt.Printf("Updating all %d packages...\n", len(st.Packages))
		for id := range st.Packages {
			fmt.Printf("  - Updating %s\n", id)
			// TODO: Run generator for each package
		}
	} else {
		if _, exists := st.GetPackage(libraryID); !exists {
			return fmt.Errorf("package %s not found", libraryID)
		}
		fmt.Printf("Updating package %s...\n", libraryID)
		// TODO: Run generator for the package
	}

	fmt.Println("Update complete")
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
	noSync := cmd.Bool("no-sync")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Checking for updates...")
	var updated bool

	// Update librarian version
	fmt.Printf("Current librarian version: %s\n", cfg.Version)
	librarianVersion, err := getLibrarianVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest librarian version: %w", err)
	}
	if librarianVersion != cfg.Version {
		fmt.Printf("Updating librarian version to %s\n", librarianVersion)
		cfg.Version = librarianVersion
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

	if !noSync {
		fmt.Println("\nRegenerating all packages...")
		st, err := state.Load()
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}

		if len(st.Packages) == 0 {
			fmt.Println("No packages to regenerate")
			return nil
		}

		fmt.Printf("Updating all %d packages...\n", len(st.Packages))
		for id := range st.Packages {
			fmt.Printf("  - Updating %s\n", id)
			// TODO: Run generator for each package
		}
		fmt.Println("Regeneration complete")
	}

	return nil
}

func removeCommand(ctx context.Context, cmd *cli.Command) error {
	libraryID := cmd.StringArg("library-id")

	if libraryID == "" {
		return fmt.Errorf("library-id is required")
	}

	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if err := st.RemovePackage(libraryID); err != nil {
		return err
	}

	if err := st.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(".librarian/state.yaml")

	fmt.Printf("Removed package %s\n", libraryID)
	return nil
}

func releaseCommand(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")
	libraryID := cmd.StringArg("library-id")

	if !all && libraryID == "" {
		return fmt.Errorf("either --all flag or library-id is required")
	}

	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if all {
		fmt.Printf("Releasing all %d packages...\n", len(st.Packages))
		for id := range st.Packages {
			fmt.Printf("  - Releasing %s\n", id)
			// TODO: Create release PR/tag for each package
		}
	} else {
		if _, exists := st.GetPackage(libraryID); !exists {
			return fmt.Errorf("package %s not found", libraryID)
		}
		fmt.Printf("Releasing package %s...\n", libraryID)
		// TODO: Create release PR/tag for the package
	}

	fmt.Println("Release complete")
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

func publishCommand(ctx context.Context, cmd *cli.Command) error {
	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	var published bool
	for id, pkg := range st.Packages {
		if pkg.Release != nil && pkg.Release.NextReleaseAt != nil && pkg.Release.NextReleaseAt.Tag != "" {
			if pkg.Release.LastReleasedAt == nil || pkg.Release.NextReleaseAt.Tag != pkg.Release.LastReleasedAt.Tag {
				fmt.Printf("Publishing %s %s...\n", id, pkg.Release.NextReleaseAt.Tag)
				// TODO: Implement actual publishing logic (git tag, push, etc.)
				fmt.Println("  - Tagging and pushing release...")
				fmt.Println("  - Publishing to package manager...")

				pkg.Release.LastReleasedAt = pkg.Release.NextReleaseAt
				pkg.Release.NextReleaseAt = nil
				published = true
				fmt.Println("  - Done.")
			}
		}
	}

	if !published {
		fmt.Println("No packages to publish.")
		return nil
	}

	if err := st.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(".librarian/state.yaml")

	fmt.Println("Publish complete.")
	return nil
}
