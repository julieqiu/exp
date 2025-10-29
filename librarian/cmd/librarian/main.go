package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	generateDescription = `The generate command is the primary tool for all code generation
tasks. It handles both the initial setup of a new library (onboarding) and the
regeneration of existing ones. Librarian works by delegating language-specific
tasks to a container, which is configured in the .librarian/state.yaml file.
Librarian is environment aware and will check if the current directory is the
root of a librarian repository. If you are not executing in such a directory the
'--repo' flag must be provided.

# Onboarding a new library

To configure and generate a new library for the first time, you must specify the
API to be generated and the library it will belong to. Librarian will invoke the
'configure' command in the language container to set up the repository, add the
new library's configuration to the '.librarian/state.yaml' file, and then
proceed with generation.

Example:
  librarian generate --library=secretmanager --api=google/cloud/secretmanager/v1

# Regenerating existing libraries

You can regenerate a single, existing library by specifying either the library
ID or the API path. If no specific library or API is provided, Librarian will
regenerate all libraries listed in '.librarian/state.yaml'. If '--library' or
'--api' is specified the whole library will be regenerated.

Examples:
  # Regenerate a single library by its ID
  librarian generate --library=secretmanager

  # Regenerate a single library by its API path
  librarian generate --api=google/cloud/secretmanager/v1

  # Regenerate all libraries in the repository
  librarian generate

# Workflow and Options:

The generation process involves delegating to the language container's
'generate' command. After the code is generated, the tool cleans the destination
directories and copies the new files into place, according to the configuration
in '.librarian/state.yaml'.

- If the '--build' flag is specified, the 'build' command is also executed in
  the container to compile and validate the generated code.
- If the '--push' flag is provided, the changes are committed to a new branch,
  and a pull request is created on GitHub. Otherwise, the changes are left in
  your local working directory for inspection. When pushing to a remote branch,
  you have the option of using HTTPS or SSH. Librarian will automatically determine
  whether to use HTTPS or SSH based on the remote URI.

Example with build and push:
  LIBRARIAN_GITHUB_TOKEN=xxx librarian generate --push --build`

	releaseStageDescription = `The 'release stage' command is the primary entry point for initiating
a new release. It automates the creation of a release pull request by parsing
conventional commits, determining the next semantic version for each library,
and generating a changelog. Librarian is environment aware and will check if the
current directory is the root of a librarian repository. If you are not
executing in such a directory the '--repo' flag must be provided.

This command scans the git history since the last release, identifies changes
(feat, fix, BREAKING CHANGE), and calculates the appropriate version bump
according to semver rules. It then delegates all language-specific file
modifications, such as updating a CHANGELOG.md or bumping the version in a pom.xml,
to the configured language-specific container.

If a specific library is configured for release via the '--library' flag, a single
releasable change is needed to automatically calculate a version bump. If there are
no releasable changes since the last release, the '--version' flag should be included
to set a new version for the library. The new version must be "SemVer" greater than the
current version.

By default, 'release stage' leaves the changes in your local working directory
for inspection. Use the '--push' flag to automatically commit the changes to
a new branch and create a pull request on GitHub. The '--commit' flag may be
used to create a local commit without creating a pull request; this flag is
ignored if '--push' is also specified. When pushing to a remote branch,
you have the option of using HTTPS or SSH. Librarian will automatically determine
whether to use HTTPS or SSH based on the remote URI.

Examples:
  # Create a release PR for all libraries with pending changes.
  librarian release stage --push

  # Create a release PR for a single library.
  librarian release stage --library=secretmanager --push

  # Manually specify a version for a single library, overriding the calculation.
  librarian release stage --library=secretmanager --library-version=2.0.0 --push`

	releaseTagDescription = `The 'tag' command is the final step in the release
process. It is designed to be run after a release pull request, created by
'release stage', has been merged.

This command's primary responsibilities are to:

- Create a Git tag for each library version included in the merged pull request.
- Create a corresponding GitHub Release for each tag, using the release notes
  from the pull request body.
- Update the pull request's label from 'release:pending' to 'release:done' to
  mark the process as complete.

You can target a specific merged pull request using the '--pr' flag. If no pull
request is specified, the command will automatically search for and process all
merged pull requests with the 'release:pending' label from the last 30 days.

Examples:
  # Tag and create a GitHub release for a specific merged PR.
  librarian release tag --repo=https://github.com/googleapis/google-cloud-go --pr=https://github.com/googleapis/google-cloud-go/pull/123

  # Find and process all pending merged release PRs in a repository.
  librarian release tag --repo=https://github.com/googleapis/google-cloud-go`
)

func main() {
	cmd := &cli.Command{
		Name:     "librarian",
		Usage:    "manages Google API client libraries by automating onboarding, regeneration, and release",
		Commands: []*cli.Command{
			newGenerateCommand(),
			newReleaseCommand(),
			newVersionCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newGenerateCommand() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Description: generateDescription,
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
	}
}

func newReleaseCommand() *cli.Command {
	return &cli.Command{
		Name:        "release",
		Description: "Manages releases of libraries.",
		Commands: []*cli.Command{
			newReleaseStageCommand(),
			newReleaseTagCommand(),
		},
	}
}

func newReleaseStageCommand() *cli.Command {
	return &cli.Command{
		Name:        "stage",
		Description: releaseStageDescription,
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
	}
}

func newReleaseTagCommand() *cli.Command {
	return &cli.Command{
		Name:        "tag",
		Description: releaseTagDescription,
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
	}
}

func newVersionCommand() *cli.Command {
	return &cli.Command{
		Name:        "version",
		Description: "Version prints version information for the librarian binary.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("librarian version 1.0.0")
			return nil
		},
	}
}
