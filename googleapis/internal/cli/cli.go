// Package cli provides the command-line interface for the googleapis tool.
package cli

import (
	"context"
	"fmt"

	"github.com/julieqiu/exp/googleapis/internal/repos"
	"github.com/julieqiu/exp/googleapis/internal/teams"
	"github.com/urfave/cli/v3"
)

// Run executes the CLI application.
func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:  "googleapis",
		Usage: "Tools for cataloging GitHub organizations",
		Commands: []*cli.Command{
			{
				Name:  "catalog",
				Usage: "Catalog GitHub organization resources",
				Commands: []*cli.Command{
					{
						Name:      "team",
						Usage:     "Catalog team(s) in the organization",
						ArgsUsage: "[team-name]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "org",
								Value:   "googleapis",
								Usage:   "GitHub organization name",
								Sources: cli.EnvVars("GITHUB_ORG"),
							},
							&cli.StringFlag{
								Name:  "output",
								Value: "data/team.yaml",
								Usage: "Output file path",
							},
							&cli.BoolFlag{
								Name:  "all",
								Usage: "Catalog all teams",
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							all := cmd.Bool("all")
							teamName := cmd.Args().First()

							if !all && teamName == "" {
								return fmt.Errorf("must specify either --all or provide a team name")
							}

							if all && teamName != "" {
								return fmt.Errorf("cannot specify both --all and a team name")
							}

							if all {
								return teams.RunAll(cmd.String("org"), cmd.String("output"))
							}
							return teams.RunSingle(cmd.String("org"), teamName, cmd.String("output"))
						},
					},
					{
						Name:      "repo",
						Usage:     "Catalog repository(ies) in the organization",
						ArgsUsage: "[repo-name]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "org",
								Value:   "googleapis",
								Usage:   "GitHub organization name",
								Sources: cli.EnvVars("GITHUB_ORG"),
							},
							&cli.StringFlag{
								Name:  "output",
								Value: "data/repository.yaml",
								Usage: "Output file path",
							},
							&cli.BoolFlag{
								Name:  "all",
								Usage: "Catalog all repositories",
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							all := cmd.Bool("all")
							repoName := cmd.Args().First()

							if !all && repoName == "" {
								return fmt.Errorf("must specify either --all or provide a repository name")
							}

							if all && repoName != "" {
								return fmt.Errorf("cannot specify both --all and a repository name")
							}

							if all {
								return repos.RunAll(cmd.String("org"), cmd.String("output"))
							}
							return repos.RunSingle(cmd.String("org"), repoName, cmd.String("output"))
						},
					},
				},
			},
		},
	}

	return cmd.Run(ctx, args)
}
