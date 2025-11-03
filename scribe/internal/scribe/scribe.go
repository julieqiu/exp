package scribe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/julieqiu/exp/scribe/internal/scraper"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

var supportedLanguages = []string{
	"cpp",
	"dotnet",
	"go",
	"java",
	"nodejs",
	"php",
	"python",
	"ruby",
	"rust",
}

// Run creates and executes the scribe CLI command.
func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:      "scribe",
		Usage:     "scrape Google Cloud documentation for language libraries",
		ArgsUsage: "<language>",
		Description: `Supported languages:
  - cpp
  - dotnet
  - go
  - java
  - nodejs
  - php
  - python
  - ruby
  - rust`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "scrape all supported languages",
			},
		},
		Action: run,
	}

	return cmd.Run(ctx, args)
}

func run(ctx context.Context, cmd *cli.Command) error {
	all := cmd.Bool("all")

	var languages []string
	if all {
		languages = supportedLanguages
	} else {
		if cmd.Args().Len() < 1 {
			return fmt.Errorf("language argument required\n\nRun 'scribe --help' for usage")
		}
		languages = []string{cmd.Args().First()}
	}

	for _, language := range languages {
		fmt.Printf("\n=== Scraping %s ===\n", language)

		libraries, err := scraper.Scrape(language)
		if err != nil {
			return fmt.Errorf("failed to scrape libraries for %s: %w", language, err)
		}

		if len(libraries) == 0 {
			fmt.Printf("No libraries found for %s\n", language)
			continue
		}

		if err := writeYAML(language, libraries); err != nil {
			return fmt.Errorf("failed to write YAML for %s: %w", language, err)
		}

		if !all {
			printTable(libraries)
		}
	}

	return nil
}

func writeYAML(language string, libraries []scraper.Library) error {
	dir := filepath.Join("testdata", "reference")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(dir, language+".yaml")
	data, err := yaml.Marshal(libraries)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Wrote %d libraries to %s\n", len(libraries), filePath)
	return nil
}

func printTable(libraries []scraper.Library) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tPackage\tLink")
	fmt.Fprintln(w, "----\t-------\t----")

	for _, lib := range libraries {
		for i, pkg := range lib.Packages {
			if i == 0 {
				fmt.Fprintf(w, "%s\t%s\t%s\n", lib.Name, pkg.Name, pkg.Link)
			} else {
				fmt.Fprintf(w, "\t%s\t%s\n", pkg.Name, pkg.Link)
			}
		}
	}

	w.Flush()
}
