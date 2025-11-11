package librarian

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRun_Version(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "--version"})
	if err != nil {
		t.Errorf("Run() with --version failed: %v", err)
	}
}

func TestRun_Help(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "--help"})
	if err != nil {
		t.Errorf("Run() with --help failed: %v", err)
	}
}

func TestRun_CommandsExist(t *testing.T) {
	for _, test := range []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "init command exists",
			args:    []string{"librarianx", "init", "--help"},
			wantErr: "",
		},
		{
			name:    "install command exists",
			args:    []string{"librarianx", "install", "--help"},
			wantErr: "",
		},
		{
			name:    "new command exists",
			args:    []string{"librarianx", "new", "--help"},
			wantErr: "",
		},
		{
			name:    "generate command exists",
			args:    []string{"librarianx", "generate", "--help"},
			wantErr: "",
		},
		{
			name:    "test command exists",
			args:    []string{"librarianx", "test", "--help"},
			wantErr: "",
		},
		{
			name:    "update command exists",
			args:    []string{"librarianx", "update", "--help"},
			wantErr: "",
		},
		{
			name:    "release command exists",
			args:    []string{"librarianx", "release", "--help"},
			wantErr: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			err := Run(ctx, test.args)
			if err != nil {
				t.Errorf("Run() failed for %v: %v", test.args, err)
			}
		})
	}
}

func TestRun_ConfigCommandRemoved(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "config", "--help"})
	if err == nil {
		t.Error("Run() with 'config' command should fail, but it succeeded")
	}
	// Just verify an error occurred - the exact error depends on the CLI framework
}

func TestRun_AddCommandRemoved(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "add", "--help"})
	if err == nil {
		t.Error("Run() with 'add' command should fail, but it succeeded")
	}
	// Just verify an error occurred - the exact error depends on the CLI framework
}

func TestInitCommand_RequiresLanguage(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "init"})
	if err == nil {
		t.Error("init command without language should fail")
	}
	if !errors.Is(err, errLanguageRequired) {
		t.Errorf("want %v; got %v", errLanguageRequired, err)
	}
}

func TestNewCommand_RequiresArtifactPath(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "new"})
	if err == nil {
		t.Error("new command without artifact path should fail")
	}
	if !errors.Is(err, errArtifactPathRequired) {
		t.Errorf("want %v; got %v", errArtifactPathRequired, err)
	}
}

func TestGenerateCommand_RequiresArtifactOrAll(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "generate"})
	if err == nil {
		t.Error("generate command without artifact or --all should fail")
	}
	if !errors.Is(err, errArtifactOrAllRequired) {
		t.Errorf("want %v; got %v", errArtifactOrAllRequired, err)
	}
}

func TestGenerateCommand_AllFlag(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "generate", "--all"})
	if err == nil {
		t.Error("expected not yet implemented error")
		return
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented' error, got: %v", err)
	}
}

func TestTestCommand_RequiresArtifactOrAll(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "test"})
	if err == nil {
		t.Error("test command without artifact or --all should fail")
	}
	if !errors.Is(err, errArtifactOrAllRequired) {
		t.Errorf("want %v; got %v", errArtifactOrAllRequired, err)
	}
}

func TestUpdateCommand_RequiresFlag(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "update"})
	if err == nil {
		t.Error("update command without flags should fail")
	}
	if !errors.Is(err, errUpdateFlagRequired) {
		t.Errorf("want %v; got %v", errUpdateFlagRequired, err)
	}
}

func TestUpdateCommand_ShaWithAll(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "update", "--all", "--sha", "abc123"})
	if err == nil {
		t.Error("update --all --sha should fail")
	}
	if !errors.Is(err, errShaWithAll) {
		t.Errorf("want %v; got %v", errShaWithAll, err)
	}
}

func TestReleaseCommand_RequiresArtifactOrAll(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "release"})
	if err == nil {
		t.Error("release command without artifact or --all should fail")
	}
	if !errors.Is(err, errArtifactOrAllRequired) {
		t.Errorf("want %v; got %v", errArtifactOrAllRequired, err)
	}
}

func TestReleaseCommand_DryRunByDefault(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "release", "secretmanager"})
	if err == nil {
		t.Error("expected not yet implemented error")
		return
	}
	if !strings.Contains(err.Error(), "DRY-RUN mode") {
		t.Errorf("expected DRY-RUN mode error, got: %v", err)
	}
}

func TestReleaseCommand_ExecuteMode(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "release", "secretmanager", "--execute"})
	if err == nil {
		t.Error("expected not yet implemented error")
		return
	}
	if !strings.Contains(err.Error(), "EXECUTE mode") {
		t.Errorf("expected EXECUTE mode error, got: %v", err)
	}
}

func TestReleaseCommand_AllFlag(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, []string{"librarianx", "release", "--all"})
	if err == nil {
		t.Error("expected not yet implemented error")
		return
	}
	if !strings.Contains(err.Error(), "all: true") {
		t.Errorf("expected all: true in error, got: %v", err)
	}
}

func TestRunInit_NotImplemented(t *testing.T) {
	ctx := context.Background()
	err := runInit(ctx, "go")
	if err == nil {
		t.Error("runInit should return not implemented error")
	}
	want := "init command not yet implemented for language: go"
	if diff := cmp.Diff(want, err.Error()); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestRunInstall_NotImplemented(t *testing.T) {
	ctx := context.Background()
	err := runInstall(ctx, "python", true)
	if err == nil {
		t.Error("runInstall should return not implemented error")
	}
	if !strings.Contains(err.Error(), "install command not yet implemented") {
		t.Errorf("expected not implemented error, got: %v", err)
	}
}

func TestRunNew_NotImplemented(t *testing.T) {
	ctx := context.Background()
	err := runNew(ctx, "secretmanager", []string{"google/cloud/secretmanager/v1"})
	if err == nil {
		t.Error("runNew should return not implemented error")
	}
	if !strings.Contains(err.Error(), "new command not yet implemented") {
		t.Errorf("expected not implemented error, got: %v", err)
	}
}

func TestRunGenerate_NotImplemented(t *testing.T) {
	ctx := context.Background()
	err := runGenerate(ctx, "secretmanager")
	if err == nil {
		t.Error("runGenerate should return not implemented error")
	}
	if !strings.Contains(err.Error(), "generate command not yet implemented") {
		t.Errorf("expected not implemented error, got: %v", err)
	}
}

func TestRunTest_NotImplemented(t *testing.T) {
	ctx := context.Background()
	err := runTest(ctx, "secretmanager")
	if err == nil {
		t.Error("runTest should return not implemented error")
	}
	if !strings.Contains(err.Error(), "test command not yet implemented") {
		t.Errorf("expected not implemented error, got: %v", err)
	}
}

func TestRunUpdate_NotImplemented(t *testing.T) {
	ctx := context.Background()
	err := runUpdate(ctx, false, true, false, "")
	if err == nil {
		t.Error("runUpdate should return not implemented error")
	}
	if !strings.Contains(err.Error(), "update command not yet implemented") {
		t.Errorf("expected not implemented error, got: %v", err)
	}
}

func TestRunRelease_NotImplemented(t *testing.T) {
	for _, test := range []struct {
		name        string
		artifactPath string
		all         bool
		execute     bool
		skipTests   bool
		skipPublish bool
		wantMode    string
	}{
		{
			name:         "dry-run mode",
			artifactPath: "secretmanager",
			all:          false,
			execute:      false,
			skipTests:    false,
			skipPublish:  false,
			wantMode:     "DRY-RUN mode",
		},
		{
			name:         "execute mode",
			artifactPath: "secretmanager",
			all:          false,
			execute:      true,
			skipTests:    false,
			skipPublish:  false,
			wantMode:     "EXECUTE mode",
		},
		{
			name:         "all dry-run",
			artifactPath: "",
			all:          true,
			execute:      false,
			skipTests:    false,
			skipPublish:  false,
			wantMode:     "DRY-RUN mode",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			err := runRelease(ctx, test.artifactPath, test.all, test.execute, test.skipTests, test.skipPublish)
			if err == nil {
				t.Error("runRelease should return not implemented error")
				return
			}
			if !strings.Contains(err.Error(), test.wantMode) {
				t.Errorf("expected %s in error, got: %v", test.wantMode, err)
			}
		})
	}
}
