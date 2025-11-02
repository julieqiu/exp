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
	language := cmd.StringArg("language")
	if language == "" {
		return fmt.Errorf("language is required")
	}
	if language != "python" {
		return fmt.Errorf("language must be python")
	}

	googleapisSHA, err := getLatestSHA("googleapis", "googleapis")
	if err != nil {
		return err
	}
	discoverySHA, err := getLatestSHA("googleapis", "discovery-artifact-manager")
	if err != nil {
		return err
	}
	librarianVersion, err := getLibrarianVersion()
	if err != nil {
		return err
	}

	cfg := &config.Config{
		Librarian: config.LibrarianConfig{
			Version:          librarianVersion,
			Language:         language,
			ReleaseTagFormat: "{id}-v{version}",
		},
		Sources: config.SourceConfig{
			Googleapis: fmt.Sprintf("https://github.com/googleapis/googleapis/archive/%s.tar.gz", googleapisSHA),
			Discovery:  fmt.Sprintf("https://github.com/googleapis/discovery-artifact-manager/archive/%s.tar.gz", discoverySHA),
			Protobuf:   "https://github.com/protocolbuffers/protobuf/releases/download/v29.3/protobuf-29.3.tar.gz",
		},
		Container: config.ContainerConfig{
			URL:     "us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator",
			Version: "latest",
		},
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	runYamlFmt(".librarian/config.yaml")

	st := &state.State{
		Libraries: make(map[string]*state.Library),
	}

	if err := st.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(".librarian/state.yaml")

	fmt.Printf("Initialized librarian repository for %s\n", language)
	fmt.Println("Created .librarian/config.yaml")
	fmt.Println("Created .librarian/state.yaml")
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

	st, err := state.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	lib := &state.Library{
		APIs: []state.API{
			{Path: apiPath},
		},
		GeneratedAt: state.Generated{
			Commit:    "c288189b43c016dd3cf1ec73ce3cadee8b732f07", // Dummy value
			Librarian: cfg.Librarian.Version,
			Image:     fmt.Sprintf("%s:%s", cfg.Container.URL, cfg.Container.Version),
		},
		LastReleasedAt: state.Release{
			Version: "v1.18.0",
			Commit:  "4a92b10e8f0a2b5c6d7e8f9a0b1c2d3e4f5a6b7c",
		},
		NextReleaseAt: state.Release{
			Version: "v1.19.0",
			Commit:  "some-new-commit-hash",
		},
	}

	st.AddLibrary(libraryID, lib)

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
		fmt.Printf("Updating all %d libraries...\n", len(st.Libraries))
		for id := range st.Libraries {
			fmt.Printf("  - Updating %s\n", id)
			// TODO: Run generator for each library
		}
	} else {
		if _, exists := st.GetLibrary(libraryID); !exists {
			return fmt.Errorf("library %s not found", libraryID)
		}
		fmt.Printf("Updating library %s...\n", libraryID)
		// TODO: Run generator for the library
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

	// TODO: Fetch latest librarian version and container image
	fmt.Println("Checking for updates...")
	fmt.Printf("Current librarian version: %s\n", cfg.Librarian.Version)
	fmt.Println("No updates available")

	if !noSync {
		fmt.Println("Regenerating all libraries...")
		// TODO: Run update --all
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

	if err := st.RemoveLibrary(libraryID); err != nil {
		return err
	}

	if err := st.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(".librarian/state.yaml")

	fmt.Printf("Removed library %s\n", libraryID)
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
		fmt.Printf("Releasing all %d libraries...\n", len(st.Libraries))
		for id := range st.Libraries {
			fmt.Printf("  - Releasing %s\n", id)
			// TODO: Create release PR/tag for each library
		}
	} else {
		if _, exists := st.GetLibrary(libraryID); !exists {
			return fmt.Errorf("library %s not found", libraryID)
		}
		fmt.Printf("Releasing library %s...\n", libraryID)
		// TODO: Create release PR/tag for the library
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
	for id, lib := range st.Libraries {
		if lib.NextReleaseAt.Version != "" && lib.NextReleaseAt.Version != lib.LastReleasedAt.Version {
			fmt.Printf("Publishing %s %s...\n", id, lib.NextReleaseAt.Version)
			// TODO: Implement actual publishing logic (git tag, push, etc.)
			fmt.Println("  - Tagging and pushing release...")
			fmt.Println("  - Publishing to package manager...")

			lib.LastReleasedAt = lib.NextReleaseAt
			lib.NextReleaseAt = state.Release{}
			published = true
			fmt.Println("  - Done.")
		}
	}

	if !published {
		fmt.Println("No libraries to publish.")
		return nil
	}

	if err := st.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	runYamlFmt(".librarian/state.yaml")

	fmt.Println("Publish complete.")
	return nil
}
