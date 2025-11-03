package surfer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

// Run runs the surfer CLI application.
func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:      "surfer",
		Usage:     "Generate gcloud command surface definitions",
		ArgsUsage: "<service>",
		Commands: []*cli.Command{
			{
				Name:      "generate",
				Usage:     "Generate gcloud surface definitions for a service",
				ArgsUsage: "<service>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "googleapis",
						Usage: "Path to googleapis repository (local directory or URL)",
						Value: "/Users/julieqiu/code/googleapis/googleapis",
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
	if cmd.Args().Len() == 0 {
		return fmt.Errorf("service name is required as the first argument")
	}

	service := cmd.Args().First()
	googleapis := cmd.String("googleapis")
	output := cmd.String("output")

	// Construct gcloud.yaml path from service name
	gcloudYAML := filepath.Join("testdata", service, "gcloud.yaml")

	// If output is not specified (default "."), use testdata/{service}/generated/
	if output == "." {
		output = filepath.Join("testdata", service, "generated")
	}

	// Print the full command being run
	cmdParts := []string{"surfer", "generate", service}
	cmdParts = append(cmdParts, fmt.Sprintf("--googleapis=%s", googleapis))
	cmdParts = append(cmdParts, fmt.Sprintf("--output=%s", output))
	fmt.Printf("%s\n\n", strings.Join(cmdParts, " "))

	return Generate(googleapis, gcloudYAML, output)
}
