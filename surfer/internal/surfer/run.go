package surfer

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

// Run runs the surfer CLI application.
func Run(ctx context.Context, args []string) error {
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
						Value: "/Users/julieqiu/code/googleapis/googleapis",
					},
					&cli.StringFlag{
						Name:  "gcloud-yaml",
						Usage: "Path to gcloud.yaml configuration file",
						Value: "testdata/parallelstore/input/gcloud.yaml",
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

	return cmd.Run(ctx, args)
}

func generateAction(ctx context.Context, cmd *cli.Command) error {
	googleapis := cmd.String("googleapis")
	gcloudYAML := cmd.String("gcloud-yaml")
	output := cmd.String("output")

	// Print the full command being run
	cmdParts := []string{"surfer", "generate"}
	cmdParts = append(cmdParts, fmt.Sprintf("--googleapis=%s", googleapis))
	cmdParts = append(cmdParts, fmt.Sprintf("--gcloud-yaml=%s", gcloudYAML))
	cmdParts = append(cmdParts, fmt.Sprintf("--output=%s", output))
	fmt.Printf("%s\n\n", strings.Join(cmdParts, " "))

	return Generate(googleapis, gcloudYAML, output)
}
