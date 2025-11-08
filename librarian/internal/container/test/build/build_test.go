package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBuild(t *testing.T) {
	// Setup temporary directories
	tmpDir := t.TempDir()
	librarianDir := filepath.Join(tmpDir, "librarian")
	repoDir := filepath.Join(tmpDir, "repo")

	if err := os.MkdirAll(librarianDir, 0755); err != nil {
		t.Fatalf("failed to create librarian dir: %v", err)
	}
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	// Copy test request file
	requestSrc := filepath.Join("..", "testdata", "secretmanager", "build-request.json")
	requestDst := filepath.Join(librarianDir, "build-request.json")
	data, err := os.ReadFile(requestSrc)
	if err != nil {
		t.Fatalf("failed to read request file: %v", err)
	}
	if err := os.WriteFile(requestDst, data, 0644); err != nil {
		t.Fatalf("failed to write request file: %v", err)
	}

	// Run build
	cfg := &Config{
		LibrarianDir: librarianDir,
		RepoDir:      repoDir,
	}

	if err := Build(context.Background(), cfg); err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Build command doesn't produce output files, just verify it runs successfully
	// The test passes if Build() doesn't return an error
}
