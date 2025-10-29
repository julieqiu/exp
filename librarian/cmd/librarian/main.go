package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "librarian",
		Usage: "manages Google API client libraries by automating onboarding, regeneration, and release",
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "The generate command is the primary tool for all code generation tasks",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "api",
						Usage: "Relative path to the API to be configured/generated (e.g., google/cloud/functions/v2)",
					},
					&cli.StringFlag{
						Name:    "api-source",
						Value:   "https://github.com/googleapis/googleapis",
						Usage:   "The location of an API specification repository",
					},
					&cli.StringFlag{
						Name:    "branch",
						Value:   "main",
						Usage:   "The branch to use with remote code repositories",
					},
					&cli.BoolFlag{
						Name:  "build",
						Usage: "If true, Librarian will build each generated library by invoking the language-specific container",
					},
					&cli.StringFlag{
						Name:  "host-mount",
						Usage: "For use when librarian is running in a container. A mapping of a directory from the host to the container, in the format <host-mount>:<local-mount>",
					},
					&cli.StringFlag{
						Name:  "image",
						Usage: "Language specific image used to invoke code generation and releasing",
					},
					&cli.StringFlag{
						Name:  "language",
						Usage: "The language of the library to generate",
					},
					&cli.StringFlag{
						Name:  "library",
						Usage: "The library ID to generate or release (e.g. google-cloud-secretmanager-v1)",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Working directory root",
					},
					&cli.BoolFlag{
						Name:  "push",
						Usage: "If true, Librarian will create a commit and a pull request for the changes",
					},
					&cli.StringFlag{
						Name:  "repo",
						Usage: "Code repository where the generated code will reside",
					},
					&cli.BoolFlag{
						Name:  "v",
						Usage: "enables verbose logging",
					},
				},
			},
			{
				Name:  "release",
				Usage: "Manages releases of libraries",
				Commands: []*cli.Command{
					{
						Name:  "stage",
						Usage: "stages a release by creating a release pull request",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "branch",
								Value:   "main",
								Usage:   "The branch to use with remote code repositories",
							},
							&cli.BoolFlag{
								Name:  "commit",
								Usage: "If true, librarian will create a commit for the release but not create a pull request",
							},
							&cli.StringFlag{
								Name:  "image",
								Usage: "Language specific image used to invoke code generation and releasing",
							},
							&cli.StringFlag{
								Name:  "library",
								Usage: "The library ID to generate or release (e.g. google-cloud-secretmanager-v1)",
							},
							&cli.StringFlag{
								Name:  "library-version",
								Usage: "Overrides the automatic semantic version calculation and forces a specific version for a library",
							},
							&cli.StringFlag{
								Name:  "output",
								Usage: "Working directory root",
							},
							&cli.BoolFlag{
								Name:  "push",
								Usage: "If true, Librarian will create a commit and a pull request for the changes",
							},
							&cli.StringFlag{
								Name:  "repo",
								Usage: "Code repository where the generated code will reside",
							},
							&cli.BoolFlag{
								Name:  "v",
								Usage: "enables verbose logging",
							},
						},
					},
					{
						Name:  "tag",
						Usage: "tags and creates a GitHub release",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "github-api-endpoint",
								Usage: "The GitHub API endpoint to use for all GitHub API operations",
							},
							&cli.StringFlag{
								Name:  "pr",
								Usage: "The URL of a pull request to operate on",
							},
							&cli.StringFlag{
								Name:  "repo",
								Usage: "Code repository where the generated code will reside",
							},
							&cli.BoolFlag{
								Name:  "v",
								Usage: "enables verbose logging",
							},
						},
					},
				},
			},
			{
				Name:  "version",
				Usage: "Version prints version information for the librarian binary",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("librarian version 1.0.0")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
