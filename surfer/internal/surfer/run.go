package surfer

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

// Run runs the surfer CLI application.
func Run(ctx context.Context) error {
	cmd := &cli.Command{
		Name:  "surfer",
		Usage: "Generate gcloud command surface definitions",
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "Generate gcloud surface definitions from a gcloud.yaml configuration file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "googleapis",
						Usage: "Path to googleapis repository (local directory or URL)",
						Value: "testdata/input/googleapis",
					},
					&cli.StringFlag{
						Name:  "gcloud-yaml",
						Usage: "Path to gcloud.yaml configuration file",
						Value: "testdata/input/parallelstore/gcloud.yaml",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Output directory for generated surfaces",
						Value: ".",
					},
				},
				Action: generateAction,
			},
		},
	}

	return cmd.Run(ctx, os.Args)
}

func generateAction(ctx context.Context, cmd *cli.Command) error {
	googleapis := cmd.String("googleapis")
	gcloudYAML := cmd.String("gcloud-yaml")
	output := cmd.String("output")

	return Generate(googleapis, gcloudYAML, output)
}
