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

func TestLoad(t *testing.T) {
	for _, test := range []struct {
		name    string
		yaml    string
		want    *Config
		wantErr bool
	}{
		{
			name: "minimal config",
			yaml: `librarian:
  version: v0.5.0
  language: go
`,
			want: &Config{
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "go",
				},
			},
		},
		{
			name: "full config with editions",
			yaml: `librarian:
  version: v0.5.0
  language: go

sources:
  googleapis:
    url: https://github.com/googleapis/googleapis/archive/abc123.tar.gz
    sha256: abc123def456

generate:
  container:
    image: us-central1-docker.pkg.dev/project/go-generator
    tag: latest
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
    apis:
      - google/cloud/secretmanager/v1
      - google/cloud/secretmanager/v1beta2
  - name: custom-tool
    version: 1.0.0
`,
			want: &Config{
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "go",
				},
				Sources: Sources{
					Googleapis: &Source{
						URL:    "https://github.com/googleapis/googleapis/archive/abc123.tar.gz",
						SHA256: "abc123def456",
					},
				},
				Generate: Generate{
					Container: &Container{
						Image: "us-central1-docker.pkg.dev/project/go-generator",
						Tag:   "latest",
					},
					OutputDir: "./",
					Defaults: &GenerateDefaults{
						Transport:        "grpc+rest",
						RestNumericEnums: boolPtr(true),
						ReleaseLevel:     "stable",
					},
				},
				Release: Release{
					TagFormat: "{id}/v{version}",
				},
				Editions: []Edition{
					{
						Name:    "secretmanager",
						Version: stringPtr("0.1.0"),
						APIs: []string{
							"google/cloud/secretmanager/v1",
							"google/cloud/secretmanager/v1beta2",
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
			yaml: `librarian:
  version: v0.5.0
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
          opt_args:
            - warehouse-package-name=google-cloud-secret-manager
      metadata:
        name_pretty: "Secret Manager"
        product_documentation: "https://cloud.google.com/secret-manager/docs"
        release_level: "stable"
      language:
        python:
          package: google-cloud-secret-manager
      keep:
        - README.md
        - docs/
      remove:
        - temp.txt
`,
			want: &Config{
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "python",
				},
				Editions: []Edition{
					{
						Name:    "google-cloud-secret-manager",
						Version: stringPtr("0.1.0"),
						Generate: &EditionGenerate{
							APIs: []API{
								{
									Path:              "google/cloud/secretmanager/v1",
									GRPCServiceConfig: "secretmanager_grpc_service_config.json",
									ServiceYAML:       "secretmanager_v1.yaml",
									Transport:         "grpc+rest",
									RestNumericEnums:  boolPtr(true),
									OptArgs: []string{
										"warehouse-package-name=google-cloud-secret-manager",
									},
								},
							},
							Metadata: &Metadata{
								NamePretty:           "Secret Manager",
								ProductDocumentation: "https://cloud.google.com/secret-manager/docs",
								ReleaseLevel:         "stable",
							},
							Language: &LanguageConfig{
								Python: &PythonConfig{
									Package: "google-cloud-secret-manager",
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
			yaml: `librarian:
  language: go
`,
			want: &Config{
				Librarian: Librarian{
					Language: "go",
				},
			},
			wantErr: false,
		},
		{
			name: "missing language",
			yaml: `librarian:
  version: v0.5.0
`,
			want: &Config{
				Librarian: Librarian{
					Version: "v0.5.0",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid language",
			yaml: `librarian:
  version: v0.5.0
  language: javascript
`,
			want: &Config{
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "javascript",
				},
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

			// Load config
			got, err := Load(configPath)
			if (err != nil) != test.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, test.wantErr)
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

func TestSave(t *testing.T) {
	config := &Config{
		Librarian: Librarian{
			Version:  "v0.5.0",
			Language: "go",
		},
		Editions: []Edition{
			{
				Name:    "secretmanager",
				Version: stringPtr("0.1.0"),
				APIs:    []string{"google/cloud/secretmanager/v1"},
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "librarian.yaml")

	// Save config
	if err := config.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load config back
	got, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
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
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "go",
				},
				Editions: []Edition{
					{Name: "secretmanager", Version: stringPtr("0.1.0")},
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			config: &Config{
				Librarian: Librarian{
					Language: "go",
				},
			},
			wantErr: true,
		},
		{
			name: "missing language",
			config: &Config{
				Librarian: Librarian{
					Version: "v0.5.0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid language",
			config: &Config{
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "javascript",
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate edition names",
			config: &Config{
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "go",
				},
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
				Librarian: Librarian{
					Version:  "v0.5.0",
					Language: "go",
				},
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

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
