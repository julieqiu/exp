// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRead(t *testing.T) {
	for _, test := range []struct {
		name    string
		yaml    string
		want    *Config
		wantErr bool
	}{
		{
			name: "minimal config",
			yaml: `version: v0.5.0
language: go
`,
			want: &Config{
				Version:  "v0.5.0",
				Language: "go",
			},
		},
		{
			name: "full config with editions",
			yaml: `version: v0.5.0
language: go

container:
  image: us-central1-docker.pkg.dev/project/go-generator
  tag: latest

sources:
  googleapis:
    url: https://github.com/googleapis/googleapis/archive/abc123.tar.gz
    sha256: abc123def456

generate:
  output_dir: ./
  defaults:
    transport: grpc+rest
    rest_numeric_enums: true
    release_level: stable

release:
  tag_format: '{id}/v{version}'

editions:
  - name: secretmanager
    version: 0.1.0
    generate:
      apis:
        - path: google/cloud/secretmanager/v1
        - path: google/cloud/secretmanager/v1beta2
  - name: custom-tool
    version: 1.0.0
`,
			want: &Config{
				Version:  "v0.5.0",
				Language: "go",
				Container: &Container{
					Image: "us-central1-docker.pkg.dev/project/go-generator",
					Tag:   "latest",
				},
				Sources: Sources{
					Googleapis: &Source{
						URL:    "https://github.com/googleapis/googleapis/archive/abc123.tar.gz",
						SHA256: "abc123def456",
					},
				},
				Generate: &Generate{
					OutputDir: "./",
					Defaults: &GenerateDefaults{
						Transport:        "grpc+rest",
						RestNumericEnums: boolPtr(true),
						ReleaseLevel:     "stable",
					},
				},
				Release: &Release{
					TagFormat: "{id}/v{version}",
				},
				Editions: []Edition{
					{
						Name:    "secretmanager",
						Version: stringPtr("0.1.0"),
						Generate: &EditionGenerate{
							APIs: []API{
								{Path: "google/cloud/secretmanager/v1"},
								{Path: "google/cloud/secretmanager/v1beta2"},
							},
						},
					},
					{
						Name:    "custom-tool",
						Version: stringPtr("1.0.0"),
					},
				},
			},
		},
		{
			name: "edition with detailed API config",
			yaml: `version: v0.5.0
language: python

editions:
  - name: google-cloud-secret-manager
    version: 0.1.0
    generate:
      apis:
        - path: google/cloud/secretmanager/v1
          grpc_service_config: secretmanager_grpc_service_config.json
          service_yaml: secretmanager_v1.yaml
          transport: grpc+rest
          rest_numeric_enums: true
          name_pretty: "Secret Manager"
          product_documentation: "https://cloud.google.com/secret-manager/docs"
          release_level: stable
          opt_args:
            - warehouse-package-name=google-cloud-secret-manager
      keep:
        - README.md
        - docs/
      remove:
        - temp.txt
`,
			want: &Config{
				Version:  "v0.5.0",
				Language: "python",
				Editions: []Edition{
					{
						Name:    "google-cloud-secret-manager",
						Version: stringPtr("0.1.0"),
						Generate: &EditionGenerate{
							APIs: []API{
								{
									Path:                 "google/cloud/secretmanager/v1",
									GRPCServiceConfig:    "secretmanager_grpc_service_config.json",
									ServiceYAML:          "secretmanager_v1.yaml",
									Transport:            "grpc+rest",
									RestNumericEnums:     boolPtr(true),
									NamePretty:           "Secret Manager",
									ProductDocumentation: "https://cloud.google.com/secret-manager/docs",
									ReleaseLevel:         "stable",
									OptArgs: []string{
										"warehouse-package-name=google-cloud-secret-manager",
									},
								},
							},
							Keep:   []string{"README.md", "docs/"},
							Remove: []string{"temp.txt"},
						},
					},
				},
			},
		},
		{
			name: "missing version",
			yaml: `language: go
`,
			want: &Config{
				Language: "go",
			},
			wantErr: false,
		},
		{
			name: "missing language",
			yaml: `version: v0.5.0
`,
			want: &Config{
				Version: "v0.5.0",
			},
			wantErr: false,
		},
		{
			name: "invalid language",
			yaml: `version: v0.5.0
language: javascript
`,
			want: &Config{
				Version:  "v0.5.0",
				Language: "javascript",
			},
			wantErr: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "librarian.yaml")
			if err := os.WriteFile(configPath, []byte(test.yaml), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			// Read config
			got, err := Read(configPath)
			if (err != nil) != test.wantErr {
				t.Fatalf("Read() error = %v, wantErr %v", err, test.wantErr)
			}

			if test.wantErr {
				return
			}

			// Compare result
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetEdition(t *testing.T) {
	config := &Config{
		Editions: []Edition{
			{Name: "secretmanager", Version: stringPtr("0.1.0")},
			{Name: "pubsub", Version: stringPtr("1.0.0")},
		},
	}

	for _, test := range []struct {
		name         string
		editionName  string
		wantEdition  *Edition
	}{
		{
			name:        "existing edition",
			editionName: "secretmanager",
			wantEdition: &Edition{Name: "secretmanager", Version: stringPtr("0.1.0")},
		},
		{
			name:        "non-existent edition",
			editionName: "nonexistent",
			wantEdition: nil,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := config.GetEdition(test.editionName)
			if diff := cmp.Diff(test.wantEdition, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	config := &Config{
		Version:  "v0.5.0",
		Language: "go",
		Editions: []Edition{
			{
				Name:    "secretmanager",
				Version: stringPtr("0.1.0"),
				Generate: &EditionGenerate{
					APIs: []API{
						{Path: "google/cloud/secretmanager/v1"},
					},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "librarian.yaml")

	// Write config
	if err := config.Write(configPath); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Read config back
	got, err := Read(configPath)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// Compare
	if diff := cmp.Diff(config, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestValidate(t *testing.T) {
	for _, test := range []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Version:  "v0.5.0",
				Language: "go",
				Editions: []Edition{
					{Name: "secretmanager", Version: stringPtr("0.1.0")},
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			config: &Config{
				Language: "go",
			},
			wantErr: true,
		},
		{
			name: "missing language",
			config: &Config{
				Version: "v0.5.0",
			},
			wantErr: true,
		},
		{
			name: "invalid language",
			config: &Config{
				Version:  "v0.5.0",
				Language: "javascript",
			},
			wantErr: true,
		},
		{
			name: "duplicate edition names",
			config: &Config{
				Version:  "v0.5.0",
				Language: "go",
				Editions: []Edition{
					{Name: "secretmanager", Version: stringPtr("0.1.0")},
					{Name: "secretmanager", Version: stringPtr("0.2.0")},
				},
			},
			wantErr: true,
		},
		{
			name: "empty edition name",
			config: &Config{
				Version:  "v0.5.0",
				Language: "go",
				Editions: []Edition{
					{Name: "", Version: stringPtr("0.1.0")},
				},
			},
			wantErr: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if (err != nil) != test.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestReadTestdata(t *testing.T) {
	for _, test := range []struct {
		name         string
		file         string
		wantLanguage string
		wantEditions int
	}{
		{
			name:         "Go config",
			file:         "testdata/go.yaml",
			wantLanguage: "go",
			wantEditions: 4, // secretmanager, pubsub, spanner, custom-tool
		},
		{
			name:         "Python config",
			file:         "testdata/python.yaml",
			wantLanguage: "python",
			wantEditions: 3, // google-cloud-secret-manager, google-cloud-pubsub, google-cloud-spanner
		},
		{
			name:         "Rust config",
			file:         "testdata/rust.yaml",
			wantLanguage: "rust",
			wantEditions: 3, // cloud-storage-v1, bigtable-admin-v2, secretmanager-v1
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Read config
			got, err := Read(test.file)
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}

			// Validate config
			if err := got.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			// Check language
			if got.Language != test.wantLanguage {
				t.Errorf("Language = %q, want %q", got.Language, test.wantLanguage)
			}

			// Check number of editions
			if len(got.Editions) != test.wantEditions {
				t.Errorf("len(Editions) = %d, want %d", len(got.Editions), test.wantEditions)
			}

			// Check that all editions have names
			for i, edition := range got.Editions {
				if edition.Name == "" {
					t.Errorf("Edition[%d] has empty name", i)
				}
			}

			// Check that sources are present for generated code
			if len(got.Editions) > 0 && got.Editions[0].Generate != nil {
				if got.Sources.Googleapis == nil {
					t.Error("Sources.Googleapis is nil but editions have generate config")
				}
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
