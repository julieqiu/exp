package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerate(t *testing.T) {
	// Setup temporary directories
	tmpDir := t.TempDir()
	librarianDir := filepath.Join(tmpDir, "librarian")
	inputDir := filepath.Join(tmpDir, "input")
	sourceDir := filepath.Join(tmpDir, "source")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(librarianDir, 0755); err != nil {
		t.Fatalf("failed to create librarian dir: %v", err)
	}
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Copy test request file
	requestSrc := filepath.Join("..", "testdata", "secretmanager", "generate-request.json")
	requestDst := filepath.Join(librarianDir, "generate-request.json")
	data, err := os.ReadFile(requestSrc)
	if err != nil {
		t.Fatalf("failed to read request file: %v", err)
	}
	if err := os.WriteFile(requestDst, data, 0644); err != nil {
		t.Fatalf("failed to write request file: %v", err)
	}

	// Run generate
	cfg := &Config{
		LibrarianDir: librarianDir,
		InputDir:     inputDir,
		OutputDir:    outputDir,
		SourceDir:    sourceDir,
	}

	if err := Generate(context.Background(), cfg); err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify output files
	expectedDir := filepath.Join("..", "testdata", "secretmanager", "generate-response")
	files := []string{"client.go", "doc.go", "version.go", "README.md"}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			gotPath := filepath.Join(outputDir, file)
			wantPath := filepath.Join(expectedDir, file)

			got, err := os.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("failed to read output file %s: %v", file, err)
			}

			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("failed to read expected file %s: %v", file, err)
			}

			if diff := cmp.Diff(string(want), string(got)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
