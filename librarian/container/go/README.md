# Go Generator Container

This package implements the Go client library generator for the Librarian system. It generates Go GAPIC client libraries from googleapis proto definitions.

## Overview

The `gogenerator` package provides a single entry point: the `Generate` function. This function orchestrates the complete workflow of generating, building, and validating Go client libraries.

## Architecture

The generator follows the container-based architecture described in the Librarian README:

```
librarian CLI (orchestrator)
  ↓ prepares request JSON
  ↓ mounts directories
  ↓ calls Generate()
gogenerator (executor)
  ↓ reads BUILD.bazel
  ↓ runs protoc
  ↓ post-processes
  ↓ builds and tests
  ↓ writes to /output
```

## Workflow

The `Generate` function performs these steps in order:

1. **Validate Configuration** - Ensures all required paths are set
2. **Read Generate Request** - Parses `generate-request.json` from LibrarianDir
3. **Load Repo Config** - Loads `repo-config.yaml` with module-specific overrides
4. **Invoke Protoc** - For each API:
   - Parse `BUILD.bazel` from googleapis API directory
   - Construct protoc command arguments
   - Execute protoc to generate Go code
   - Generate `.repo-metadata.json` for each API
5. **Fix Permissions** - Set all generated `.go` files to 0644
6. **Flatten Output** - Move `/output/cloud.google.com/go/*` to `/output/*`
7. **Apply Module Version** - Reorganize output for versioned modules (if needed)
8. **Post-Process** (if enabled):
   - Update snippet metadata with version info
   - Run `goimports` formatter
   - Run `go mod init` and `go mod tidy` for new modules only
9. **Delete Output Paths** - Remove paths specified in config
10. **Build** - Run `go build ./...` to validate compilation
11. **Test** - Run `go test ./... -short` to validate tests pass

## Configuration

### Config Struct

```go
type Config struct {
    LibrarianDir         string // Path to librarian input directory
    SourceDir            string // Path to googleapis repository checkout
    OutputDir            string // Path where generated code will be written
    InputDir             string // Path to generator input templates/config
    DisablePostProcessor bool   // Controls whether post-processing runs
}
```

### GenerateRequest JSON

Expected at `{LibrarianDir}/generate-request.json`:

```json
{
  "id": "secretmanager",
  "version": "1.2.0",
  "apis": [
    {
      "path": "google/cloud/secretmanager/v1",
      "service_config": "secretmanager_v1.yaml",
      "status": "existing"
    }
  ],
  "source_roots": ["secretmanager"],
  "preserve_regex": [],
  "remove_regex": [],
  "status": "existing"
}
```

### RepoConfig YAML

Optional file at `{LibrarianDir}/.librarian/generator-input/repo-config.yaml`:

```yaml
modules:
  - name: secretmanager
    module_path_version: v2
    apis:
      - path: google/cloud/secretmanager/v1
        disable_gapic: false
        nested_protos: []
    delete_generation_output_paths:
      - temp
```

## Bazel Integration

The generator parses `BUILD.bazel` files from googleapis to extract configuration:

### Supported BUILD.bazel Rules

- `go_gapic_library` - GAPIC client library configuration
- `go_grpc_library` - Modern gRPC library (preferred)
- `go_proto_library` - Legacy proto library (with optional gRPC plugin)

### Extracted Configuration

From `go_gapic_library`:
- `importpath` - Go import path with package suffix
- `service_yaml` - Service config filename
- `grpc_service_config` - gRPC service config filename
- `transport` - "grpc", "rest", or "grpc+rest"
- `release_level` - "beta" or "ga"
- `metadata` - Whether to generate gapic_metadata.json
- `diregapic` - Whether to use Discovery REST GAPIC
- `rest_numeric_enums` - Whether REST client supports numeric enums

## Protoc Command Construction

The generator constructs protoc commands like this:

```bash
protoc --experimental_allow_proto3_optional \
  --go_out=/output \
  --go-grpc_out=/output \
  --go-grpc_opt=require_unimplemented_servers=false \
  --go_gapic_out=/output \
  --go_gapic_opt=go-gapic-package=cloud.google.com/go/functions/apiv2;functions \
  --go_gapic_opt=api-service-config=/source/google/cloud/functions/v2/cloudfunctions_v2.yaml \
  --go_gapic_opt=transport=grpc+rest \
  -I=/source \
  /source/google/cloud/functions/v2/functions.proto \
  /source/google/cloud/functions/v2/operations.proto
```

### Plugins Used

- **Modern** (when `go_grpc_library` present):
  - `protoc-gen-go` - Proto message generation
  - `protoc-gen-go-grpc` - gRPC service stubs
  - `protoc-gen-go_gapic` - GAPIC client library

- **Legacy** (when only `go_proto_library` present):
  - `protoc-gen-go_v1` - Proto and gRPC in one plugin

## Post-Processing

When `DisablePostProcessor` is false:

1. **Update Snippet Metadata**
   - Replaces `$VERSION` placeholder with actual version
   - Updates files in `internal/generated/snippets/{module}/*/snippet_metadata.*.json`

2. **Run goimports**
   - Formats code and manages imports
   - Runs in output directory

3. **Initialize Module** (only for new modules)
   - Runs `go mod init {modulePath}`
   - Runs `go mod tidy`
   - Only executes when `status == "new"`

## Module Paths

The generator uses these module path conventions:

- **Unversioned**: `cloud.google.com/go/{module_name}`
- **Versioned**: `cloud.google.com/go/{module_name}/{version}`

Example:
- `secretmanager` → `cloud.google.com/go/secretmanager`
- `dataproc` with `module_path_version: v2` → `cloud.google.com/go/dataproc/v2`

## Output Structure

Before flattening:
```
/output/
  cloud.google.com/go/
    secretmanager/
      client.go
      ...
```

After flattening:
```
/output/
  secretmanager/
    client.go
    ...
```

For versioned modules, the generator also moves:
```
/output/secretmanager/v2/* → /output/secretmanager/*
/output/internal/generated/snippets/secretmanager/v2/* → /output/internal/generated/snippets/secretmanager/*
```

## Testing

Run tests:

```bash
go test ./container/go/... -v
```

Tests cover:
- Configuration validation
- JSON request parsing
- Bazel BUILD.bazel parsing
- Module path generation
- Proto file gathering
- Output flattening
- Snippet metadata updates
- Command execution

## Dependencies

External tools required:
- `protoc` - Protocol buffer compiler
- `protoc-gen-go` - Go proto generator
- `protoc-gen-go-grpc` - Go gRPC generator
- `protoc-gen-go_gapic` - Go GAPIC generator
- `goimports` - Go import formatter
- `go` - Go toolchain (for build/test)

## Error Handling

All errors are wrapped with context using `fmt.Errorf` with `%w` verb. This allows the caller to inspect the error chain and understand exactly where failures occurred.

Example error messages:
- `"invalid configuration: LibrarianDir is required"`
- `"failed to read generate request: failed to read /librarian/generate-request.json: no such file or directory"`
- `"protoc generation failed: protoc failed for google/cloud/functions/v2: command failed: exit status 1"`

## Future Enhancements

Possible improvements:
- Support for YAML parsing (currently uses placeholder for RepoConfig)
- More sophisticated .repo-metadata.json generation
- Parallel protoc execution for multiple APIs
- Incremental generation (skip unchanged APIs)
- Detailed progress logging
