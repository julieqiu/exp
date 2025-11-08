package release

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestStage(t *testing.T) {
	// Setup temporary directories
	tmpDir := t.TempDir()
	librarianDir := filepath.Join(tmpDir, "librarian")
	repoDir := filepath.Join(tmpDir, "repo")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(librarianDir, 0755); err != nil {
		t.Fatalf("failed to create librarian dir: %v", err)
	}
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Copy test request file
	requestSrc := filepath.Join("..", "testdata", "secretmanager", "release-stage-request.json")
	requestDst := filepath.Join(librarianDir, "release-stage-request.json")
	data, err := os.ReadFile(requestSrc)
	if err != nil {
		t.Fatalf("failed to read request file: %v", err)
	}
	if err := os.WriteFile(requestDst, data, 0644); err != nil {
		t.Fatalf("failed to write request file: %v", err)
	}

	// Run release-stage
	cfg := &Config{
		LibrarianDir: librarianDir,
		RepoDir:      repoDir,
		OutputDir:    outputDir,
	}

	if err := Stage(context.Background(), cfg); err != nil {
		t.Fatalf("Stage() failed: %v", err)
	}

	// Verify output files
	expectedDir := filepath.Join("..", "testdata", "secretmanager", "release-stage-response")

	t.Run("version.go", func(t *testing.T) {
		gotPath := filepath.Join(outputDir, "version.go")
		wantPath := filepath.Join(expectedDir, "version.go")

		got, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}

		want, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatalf("failed to read expected file: %v", err)
		}

		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("CHANGES.md", func(t *testing.T) {
		gotPath := filepath.Join(outputDir, "CHANGES.md")
		wantPath := filepath.Join(expectedDir, "CHANGES.md")

		got, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}

		want, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatalf("failed to read expected file: %v", err)
		}

		// Replace {{DATE}} placeholder with today's date
		today := time.Now().Format("2006-01-02")
		wantStr := strings.ReplaceAll(string(want), "{{DATE}}", today)

		if diff := cmp.Diff(wantStr, string(got)); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
}
