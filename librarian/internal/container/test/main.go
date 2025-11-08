package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/julieqiu/exp/librarian/internal/container/test/build"
	"github.com/julieqiu/exp/librarian/internal/container/test/configure"
	"github.com/julieqiu/exp/librarian/internal/container/test/generate"
	"github.com/julieqiu/exp/librarian/internal/container/test/release"
)

const version = "0.1.0"

func main() {
	logLevel := slog.LevelInfo
	if os.Getenv("GOOGLE_SDK_GO_LOGGING_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))
	slog.Debug("testcontainer: invoked", "args", os.Args)
	if err := run(context.Background(), os.Args[1:]); err != nil {
		slog.Error("testcontainer: failed", "error", err)
		os.Exit(1)
	}
	slog.Debug("testcontainer: finished successfully")
}

var (
	generateFunc  = generate.Generate
	releaseFunc   = release.Stage
	buildFunc     = build.Build
	configureFunc = configure.Configure
)

// run executes the appropriate command based on the CLI's invocation arguments.
// The idiomatic structure is `testcontainer [command] [flags]`.
func run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("testcontainer: expected a command")
	}

	// The --version flag is a special case and not a command.
	if args[0] == "--version" {
		fmt.Println(version)
		return nil
	}

	cmd := args[0]
	flags := args[1:]

	if strings.HasPrefix(cmd, "-") {
		return fmt.Errorf("testcontainer: command cannot be a flag: %s", cmd)
	}

	switch cmd {
	case "generate":
		return handleGenerate(ctx, flags)
	case "release-stage":
		return handleReleaseStage(ctx, flags)
	case "configure":
		return handleConfigure(ctx, flags)
	case "build":
		return handleBuild(ctx, flags)
	default:
		return fmt.Errorf("testcontainer: unknown command: %s", cmd)
	}
}

// handleGenerate parses flags for the generate command and calls the generator.
func handleGenerate(ctx context.Context, args []string) error {
	cfg := &generate.Config{}
	generateFlags := flag.NewFlagSet("generate", flag.ContinueOnError)
	generateFlags.StringVar(&cfg.LibrarianDir, "librarian", "/librarian", "Path to the librarian-tool input directory. Contains generate-request.json.")
	generateFlags.StringVar(&cfg.InputDir, "input", "/input", "Path to the .librarian/generator-input directory from the language repository.")
	generateFlags.StringVar(&cfg.OutputDir, "output", "/output", "Path to the empty directory where the test container writes its output.")
	generateFlags.StringVar(&cfg.SourceDir, "source", "/source", "Path to a complete checkout of the googleapis repository.")
	if err := generateFlags.Parse(args); err != nil {
		return fmt.Errorf("testcontainer: failed to parse flags: %w", err)
	}
	return generateFunc(ctx, cfg)
}

// handleReleaseStage parses flags for the release-stage command and calls the release tool.
func handleReleaseStage(ctx context.Context, args []string) error {
	cfg := &release.Config{}
	releaseFlags := flag.NewFlagSet("release-stage", flag.ContinueOnError)
	releaseFlags.StringVar(&cfg.LibrarianDir, "librarian", "/librarian", "Path to the librarian-tool input directory. Contains release-stage-request.json.")
	releaseFlags.StringVar(&cfg.RepoDir, "repo", "/repo", "Path to the language repository checkout.")
	releaseFlags.StringVar(&cfg.OutputDir, "output", "/output", "Path to the empty directory where the test container writes its output.")
	if err := releaseFlags.Parse(args); err != nil {
		return fmt.Errorf("testcontainer: failed to parse flags: %w", err)
	}
	return releaseFunc(ctx, cfg)
}

// handleBuild parses flags for the build command and calls the builder.
func handleBuild(ctx context.Context, args []string) error {
	cfg := &build.Config{}
	buildFlags := flag.NewFlagSet("build", flag.ContinueOnError)
	buildFlags.StringVar(&cfg.LibrarianDir, "librarian", "/librarian", "Path to the librarian-tool input directory. Contains build-request.json.")
	buildFlags.StringVar(&cfg.RepoDir, "repo", "/repo", "Path to the root of the complete language repository.")
	if err := buildFlags.Parse(args); err != nil {
		return fmt.Errorf("testcontainer: failed to parse flags: %w", err)
	}
	return buildFunc(ctx, cfg)
}

// handleConfigure parses flags for the configure command and calls the configure code.
func handleConfigure(ctx context.Context, args []string) error {
	cfg := &configure.Config{}
	configureFlags := flag.NewFlagSet("configure", flag.ContinueOnError)
	configureFlags.StringVar(&cfg.LibrarianDir, "librarian", "/librarian", "Path to the librarian-tool input directory. Contains configure-request.json.")
	configureFlags.StringVar(&cfg.InputDir, "input", "/input", "Path to the .librarian/generator-input directory from the language repository.")
	configureFlags.StringVar(&cfg.RepoDir, "repo", "/repo", "Path to a read-only copy of relevant language repo files.")
	configureFlags.StringVar(&cfg.OutputDir, "output", "/output", "Path to the empty directory where the test container writes its output.")
	configureFlags.StringVar(&cfg.SourceDir, "source", "/source", "Path to a complete checkout of the googleapis repository.")
	if err := configureFlags.Parse(args); err != nil {
		return fmt.Errorf("testcontainer: failed to parse flags: %w", err)
	}
	return configureFunc(ctx, cfg)
}
