package configure

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfigure(t *testing.T) {
	// Setup temporary directories
	tmpDir := t.TempDir()
	librarianDir := filepath.Join(tmpDir, "librarian")
	inputDir := filepath.Join(tmpDir, "input")
	repoDir := filepath.Join(tmpDir, "repo")
	sourceDir := filepath.Join(tmpDir, "source")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(librarianDir, 0755); err != nil {
		t.Fatalf("failed to create librarian dir: %v", err)
	}
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Copy test request file
	requestSrc := filepath.Join("..", "testdata", "secretmanager", "configure-request.json")
	requestDst := filepath.Join(librarianDir, "configure-request.json")
	data, err := os.ReadFile(requestSrc)
	if err != nil {
		t.Fatalf("failed to read request file: %v", err)
	}
	if err := os.WriteFile(requestDst, data, 0644); err != nil {
		t.Fatalf("failed to write request file: %v", err)
	}

	// Run configure
	cfg := &Config{
		LibrarianDir: librarianDir,
		InputDir:     inputDir,
		RepoDir:      repoDir,
		OutputDir:    outputDir,
		SourceDir:    sourceDir,
	}

	if err := Configure(context.Background(), cfg); err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	// Verify output file
	gotPath := filepath.Join(outputDir, "configure-response.json")
	wantPath := filepath.Join("..", "testdata", "secretmanager", "configure-response.json")

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
}
