package gogenerator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidateConfig(t *testing.T) {
	for _, test := range []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				LibrarianDir: "/librarian",
				SourceDir:    "/source",
				OutputDir:    "/output",
			},
			wantErr: false,
		},
		{
			name: "missing LibrarianDir",
			cfg: &Config{
				SourceDir: "/source",
				OutputDir: "/output",
			},
			wantErr: true,
		},
		{
			name: "missing SourceDir",
			cfg: &Config{
				LibrarianDir: "/librarian",
				OutputDir:    "/output",
			},
			wantErr: true,
		},
		{
			name: "missing OutputDir",
			cfg: &Config{
				LibrarianDir: "/librarian",
				SourceDir:    "/source",
			},
			wantErr: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateConfig(test.cfg)
			if (err != nil) != test.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestReadGenerateRequest(t *testing.T) {
	// Create temp directory with test request file
	tmpDir := t.TempDir()

	req := &GenerateRequest{
		ID:      "secretmanager",
		Version: "1.2.0",
		APIs: []API{
			{Path: "google/cloud/secretmanager/v1", Status: "existing"},
		},
		Status: "existing",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	requestPath := filepath.Join(tmpDir, "generate-request.json")
	if err := os.WriteFile(requestPath, data, 0644); err != nil {
		t.Fatalf("failed to write request file: %v", err)
	}

	// Test reading the request
	got, err := readGenerateRequest(tmpDir)
	if err != nil {
		t.Fatalf("readGenerateRequest() error = %v", err)
	}

	if diff := cmp.Diff(req, got); diff != "" {
		t.Errorf("readGenerateRequest() mismatch (-want +got):\n%s", diff)
	}
}

func TestFindString(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
		key     string
		want    string
	}{
		{
			name:    "simple string",
			content: `importpath = "cloud.google.com/go/functions/apiv2;functions"`,
			key:     "importpath",
			want:    "cloud.google.com/go/functions/apiv2;functions",
		},
		{
			name:    "with spaces",
			content: `grpc_service_config  =  "functions_v2_grpc_service_config.json"`,
			key:     "grpc_service_config",
			want:    "functions_v2_grpc_service_config.json",
		},
		{
			name:    "not found",
			content: `foo = "bar"`,
			key:     "missing",
			want:    "",
		},
		{
			name: "multiline",
			content: `go_gapic_library(
    name = "functions_go_gapic",
    importpath = "cloud.google.com/go/functions/apiv2;functions",
)`,
			key:  "importpath",
			want: "cloud.google.com/go/functions/apiv2;functions",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := findString(test.content, test.key)
			if got != test.want {
				t.Errorf("findString(%q, %q) = %q, want %q", test.content, test.key, got, test.want)
			}
		})
	}
}

func TestFindBool(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
		key     string
		want    bool
		wantErr bool
	}{
		{
			name:    "true value",
			content: `metadata = True`,
			key:     "metadata",
			want:    true,
			wantErr: false,
		},
		{
			name:    "false value",
			content: `rest_numeric_enums = False`,
			key:     "rest_numeric_enums",
			want:    false,
			wantErr: false,
		},
		{
			name:    "not found",
			content: `foo = True`,
			key:     "missing",
			want:    false,
			wantErr: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := findBool(test.content, test.key)
			if (err != nil) != test.wantErr {
				t.Errorf("findBool() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("findBool(%q, %q) = %v, want %v", test.content, test.key, got, test.want)
			}
		})
	}
}

func TestParseBazelConfig(t *testing.T) {
	// Create temp directory with BUILD.bazel
	tmpDir := t.TempDir()

	buildContent := `
go_gapic_library(
    name = "functions_go_gapic",
    importpath = "cloud.google.com/go/functions/apiv2;functions",
    grpc_service_config = "functions_v2_grpc_service_config.json",
    service_yaml = "cloudfunctions_v2.yaml",
    transport = "grpc+rest",
    metadata = True,
)

go_grpc_library(
    name = "functions_go_grpc",
)
`

	buildPath := filepath.Join(tmpDir, "BUILD.bazel")
	if err := os.WriteFile(buildPath, []byte(buildContent), 0644); err != nil {
		t.Fatalf("failed to write BUILD.bazel: %v", err)
	}

	got, err := parseBazelConfig(tmpDir)
	if err != nil {
		t.Fatalf("parseBazelConfig() error = %v", err)
	}

	want := &BazelConfig{
		hasGAPIC:          true,
		hasGoGRPC:         true,
		gapicImportPath:   "cloud.google.com/go/functions/apiv2;functions",
		grpcServiceConfig: "functions_v2_grpc_service_config.json",
		serviceYAML:       "cloudfunctions_v2.yaml",
		transport:         "grpc+rest",
		metadata:          true,
	}

	if diff := cmp.Diff(want, got, cmp.AllowUnexported(BazelConfig{})); diff != "" {
		t.Errorf("parseBazelConfig() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetModulePath(t *testing.T) {
	for _, test := range []struct {
		name       string
		config     *ModuleConfig
		libraryID  string
		want       string
	}{
		{
			name: "unversioned module",
			config: &ModuleConfig{
				Name: "secretmanager",
			},
			libraryID: "secretmanager",
			want:      "cloud.google.com/go/secretmanager",
		},
		{
			name: "versioned module",
			config: &ModuleConfig{
				Name:              "dataproc",
				ModulePathVersion: "v2",
			},
			libraryID: "dataproc",
			want:      "cloud.google.com/go/dataproc/v2",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := getModulePath(test.config, test.libraryID)
			if got != test.want {
				t.Errorf("getModulePath() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestGatherProtoFiles(t *testing.T) {
	// Create temp directory with proto files
	tmpDir := t.TempDir()

	// Create some test proto files
	protoFiles := []string{"api.proto", "service.proto", "types.proto"}
	for _, f := range protoFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("syntax = \"proto3\";"), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", f, err)
		}
	}

	// Create a non-proto file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}

	for _, test := range []struct {
		name         string
		nestedProtos []string
		wantCount    int
	}{
		{
			name:      "no nested protos",
			wantCount: 3, // api.proto, service.proto, types.proto
		},
		{
			name:         "with nested protos",
			nestedProtos: []string{"nested/extra.proto"},
			wantCount:    4, // 3 top-level + 1 nested
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := gatherProtoFiles(tmpDir, test.nestedProtos)
			if err != nil {
				t.Fatalf("gatherProtoFiles() error = %v", err)
			}

			if len(got) != test.wantCount {
				t.Errorf("gatherProtoFiles() returned %d files, want %d", len(got), test.wantCount)
			}

			// Verify all returned files end with .proto
			for _, f := range got {
				if filepath.Ext(f) != ".proto" {
					t.Errorf("gatherProtoFiles() returned non-proto file: %s", f)
				}
			}
		})
	}
}

func TestUpdateSnippetsMetadata(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	snippetsDir := filepath.Join(tmpDir, "internal", "generated", "snippets", "secretmanager")
	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		t.Fatalf("failed to create snippets dir: %v", err)
	}

	// Create test snippet metadata file with $VERSION placeholder
	metadataContent := `{
  "clientLibrary": {
    "version": "$VERSION",
    "name": "secretmanager"
  }
}`

	metadataPath := filepath.Join(snippetsDir, "snippet_metadata.secretmanager.json")
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("failed to write metadata file: %v", err)
	}

	// Run update
	req := &GenerateRequest{
		ID:      "secretmanager",
		Version: "1.5.0",
	}

	if err := updateSnippetsMetadata(tmpDir, req); err != nil {
		t.Fatalf("updateSnippetsMetadata() error = %v", err)
	}

	// Verify the file was updated
	updated, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	want := `{
  "clientLibrary": {
    "version": "1.5.0",
    "name": "secretmanager"
  }
}`

	if diff := cmp.Diff(want, string(updated)); diff != "" {
		t.Errorf("updateSnippetsMetadata() file content mismatch (-want +got):\n%s", diff)
	}
}

func TestFlattenOutput(t *testing.T) {
	// Create temp directory with cloud.google.com/go structure
	tmpDir := t.TempDir()
	cloudPath := filepath.Join(tmpDir, "cloud.google.com", "go")
	if err := os.MkdirAll(cloudPath, 0755); err != nil {
		t.Fatalf("failed to create cloud path: %v", err)
	}

	// Create some test content
	testDirs := []string{"secretmanager", "functions"}
	for _, dir := range testDirs {
		dirPath := filepath.Join(cloudPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("failed to create %s: %v", dir, err)
		}
		// Create a test file
		testFile := filepath.Join(dirPath, "client.go")
		if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Run flatten
	if err := flattenOutput(tmpDir); err != nil {
		t.Fatalf("flattenOutput() error = %v", err)
	}

	// Verify directories were moved to top level
	for _, dir := range testDirs {
		topLevelPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(topLevelPath); os.IsNotExist(err) {
			t.Errorf("expected directory %s at top level, but it doesn't exist", dir)
		}

		// Verify test file exists
		testFile := filepath.Join(topLevelPath, "client.go")
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Errorf("expected file %s, but it doesn't exist", testFile)
		}
	}

	// Verify cloud.google.com was removed
	cloudDir := filepath.Join(tmpDir, "cloud.google.com")
	if _, err := os.Stat(cloudDir); !os.IsNotExist(err) {
		t.Errorf("expected cloud.google.com to be removed, but it still exists")
	}
}

func TestRunCommand(t *testing.T) {
	ctx := context.Background()

	for _, test := range []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command",
			args:    []string{"echo", "hello"},
			wantErr: false,
		},
		{
			name:    "failed command",
			args:    []string{"false"},
			wantErr: true,
		},
		{
			name:    "no command",
			args:    []string{},
			wantErr: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := runCommand(ctx, test.args, "")
			if (err != nil) != test.wantErr {
				t.Errorf("runCommand() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
