# Test Container

This directory contains the source code for the test container,
a lightweight mock implementation that simulates a language-specific code
generator for testing the Librarian pipeline.
It follows the same container contract as real generators (like `librariangen`)
but produces simple placeholder output instead of actually generating code.

## Purpose

The test container allows you to:

- Develop and test the Librarian CLI without needing real generator toolchains
- Verify container orchestration, mounting, and JSON communication
- Test workflows end-to-end with predictable, deterministic output
- Debug Librarian behavior in isolation from complex generation logic

## How it Works (The Container Contract)

The test container binary is designed to be run inside a Docker container
orchestrated by the Librarian tool.
It adheres to the same "container contract" as real generators by accepting
commands and expecting a set of mounted directories for its inputs and outputs.

The primary commands are `generate`, `release-stage`, `configure`, and `build`.

### `generate` Command

This command simulates code generation by creating placeholder files in the output directory.

**Example `generate` command:**
```bash
testcontainer generate \
    --source /source \
    --librarian /librarian \
    --input /input \
    --output /output
```

### `release-stage` Command

This command simulates updating version and changelog files for a release.

**Example `release-stage` command:**
```bash
testcontainer release-stage \
    --repo /repo \
    --librarian /librarian \
    --output /output
```

### `configure` Command

This command simulates validating and configuring library settings.

**Example `configure` command:**
```bash
testcontainer configure \
    --source /source \
    --librarian /librarian \
    --input /input \
    --repo /repo \
    --output /output
```

### `build` Command

This command simulates building and testing generated libraries.

**Example `build` command:**
```bash
testcontainer build \
    --librarian /librarian \
    --repo /repo
```

### `generate` Command Workflow

1. **Inputs:** The container is provided with several mounted directories:
   - `/source`: A complete checkout of the API source repository (e.g., googleapis)
   - `/librarian`: Contains a `generate-request.json` file specifying the library and APIs to generate
   - `/input`: Generator-specific templates or configuration from the language repository
   - `/output`: An empty directory where all generated files will be written

2. **Execution:**
   - The binary parses `generate-request.json`
   - For each API specified in the request, it creates simple placeholder files
   - It logs all operations for debugging
   - It validates that required directories exist

3. **Output:** Simple placeholder files are written to the `/output` directory:
   - `client.go` - A minimal client implementation
   - `doc.go` - Package documentation
   - `README.md` - Library documentation
   - `version.go` - Version constant

### `release-stage` Command Workflow

1. **Inputs:**
   - `/repo`: A complete checkout of the language repository containing the library
   - `/librarian`: Contains a `release-stage-request.json` file specifying version and changes
   - `/output`: An empty directory where modified files will be written

2. **Execution:**
   - The binary parses `release-stage-request.json`
   - It creates updated version and changelog files
   - It formats changelog entries from the provided changes

3. **Output:** Updated files are written to `/output`:
   - `version.go` - Updated version constant
   - `CHANGES.md` - Updated changelog with new entries

### `configure` Command Workflow

1. **Inputs:**
   - `/source`: API source repository checkout
   - `/librarian`: Contains a `configure-request.json` file
   - `/input`: Generator input directory
   - `/repo`: Read-only copy of language repository
   - `/output`: Empty directory for configuration files

2. **Execution:**
   - The binary parses `configure-request.json`
   - It validates library and API configurations
   - It creates configuration metadata files

3. **Output:** Configuration files written to `/output`:
   - `config.json` - Validated configuration
   - `.librarian.yaml` - Per-library metadata

### `build` Command Workflow

1. **Inputs:**
   - `/librarian`: Contains a `build-request.json` file
   - `/repo`: Complete language repository checkout

2. **Execution:**
   - The binary parses `build-request.json`
   - It simulates running tests and builds
   - It validates that the library can be built

3. **Output:** Exit code 0 for success, non-zero for failure. Logs build steps.

## Request JSON Formats

### generate-request.json

```json
{
  "id": "library-name",
  "version": "0.1.0",
  "apis": [
    {
      "path": "google/example/v1",
      "service_config": "example_v1.yaml"
    }
  ],
  "source_roots": [
    "library-name",
    "internal/snippets/library-name"
  ],
  "preserve_regex": [
    "library-name/CHANGES.md",
    "library-name/custom.go"
  ],
  "remove_regex": [
    "library-name/temp"
  ]
}
```

**Fields:**
- `id` - Library identifier
- `version` - Library version to generate
- `apis` - List of API paths to include in the library
- `source_roots` - Directories where generated code will be placed
- `preserve_regex` - Files/patterns to not overwrite during generation
- `remove_regex` - Files/patterns to remove after generation

### release-stage-request.json

```json
{
  "libraries": [
    {
      "id": "library-name",
      "version": "1.2.0",
      "changes": [
        {
          "type": "feat",
          "subject": "add new GetSecret API",
          "body": "This adds the ability to retrieve metadata.",
          "piper_cl_number": "123456789",
          "commit_hash": "a1b2c3d4"
        },
        {
          "type": "fix",
          "subject": "correct typo in documentation",
          "body": "Fixed a minor typo.",
          "piper_cl_number": "987654321",
          "commit_hash": "f6e5d4c3"
        }
      ],
      "apis": [
        {
          "path": "google/example/v1"
        }
      ],
      "source_roots": [
        "library-name"
      ],
      "release_triggered": true
    }
  ]
}
```

**Fields:**
- `libraries` - List of libraries to prepare for release
- `id` - Library identifier
- `version` - New version number
- `changes` - List of changelog entries
- `apis` - APIs included in this library
- `source_roots` - Library directories
- `release_triggered` - Whether to actually create the release

### configure-request.json

```json
{
  "libraries": [
    {
      "id": "library-name",
      "apis": [
        {
          "path": "google/example/v1",
          "service_config": "example_v1.yaml",
          "status": "new"
        }
      ],
      "source_roots": [
        "library-name"
      ]
    }
  ]
}
```

**Fields:**
- `libraries` - List of libraries to configure
- `id` - Library identifier
- `apis` - APIs to configure (with status: "new" or "existing")
- `source_roots` - Library directories

## Running

### Run as a CLI Binary

This is the fastest way to test during development.

1. **Prerequisites:**
   - Go toolchain installed

2. **Prepare Inputs:**
   ```bash
   mkdir -p /tmp/testcontainer/librarian /tmp/testcontainer/output
   ```

3. **Execute:**
   ```bash
   go run ./internal/container/test generate \
     --source=/tmp/source \
     --librarian=/tmp/testcontainer/librarian \
     --output=/tmp/testcontainer/output
   ```

### Run with Docker

1. **Build the container:**
   ```bash
   docker build -t testcontainer:latest ./internal/container/test
   ```

2. **Run the container:**
   ```bash
   docker run \
     -v /tmp/source:/source \
     -v /tmp/testcontainer/librarian:/librarian \
     -v /tmp/testcontainer/output:/output \
     testcontainer:latest \
     generate --source=/source --librarian=/librarian --output=/output
   ```

### Run via Librarian CLI

This is the intended production usage.

```bash
librarian generate \
  --image=testcontainer:latest \
  --repo=/path/to/repo \
  --library=example-lib \
  --api=google/example/v1 \
  --api-source=/path/to/googleapis
```

## Development & Testing

### Running Tests

```bash
go test ./internal/container/test/...
```

### Building the Binary

```bash
go build ./internal/container/test
```

### Iterative Development Workflow

When developing the test container, use this workflow for fast iteration:

1. **Make changes to the test container code:**
   ```bash
   vim internal/container/test/generate/generate.go
   ```

2. **Rebuild the Docker image:**
   ```bash
   docker build -t librarian-test:latest -f internal/container/test/Dockerfile internal/container/test
   ```

   The `latest` tag is overwritten each time, so your `.librarian.yaml` files don't need updating.

3. **Test with Librarian:**
   ```bash
   librarian generate secretmanager
   ```

4. **Repeat steps 1-3 as needed.**

### Tagging Stable Versions

When you reach a stable checkpoint and want to create a versioned tag:

```bash
# Build and tag a specific version
docker build -t librarian-test:v0.1.0 -f internal/container/test/Dockerfile internal/container/test

# Update the global config to use this version
librarian config set generate.container.tag v0.1.0

# Regenerate all artifacts to sync the new tag
librarian generate --all
```

The config structure separates image name and tag:

```yaml
generate:
  container:
    image: librarian-test    # Image name/registry path
    tag: latest              # Image tag
```

Update them independently or together:

```bash
# Syntactic sugar: set both image and tag at once
librarian config set generate.container librarian-test:v0.2.0

# Equivalent to:
librarian config set generate.container.image librarian-test
librarian config set generate.container.tag v0.2.0

# Or change just the tag
librarian config set generate.container.tag v0.3.0

# Or change just the image name (rarely needed)
librarian config set generate.container.image my-custom-test
```

After updating the config, regenerate to sync changes to all `.librarian.yaml` files:

```bash
librarian generate --all
```

### Example Test Workflow

1. Create test input:
   ```bash
   mkdir -p /tmp/test/{source,librarian,input,output}

   cat > /tmp/test/librarian/generate-request.json <<EOF
   {
     "id": "example",
     "version": "0.1.0",
     "apis": [{"path": "google/example/v1"}],
     "source_roots": ["example"]
   }
   EOF
   ```

2. Run the generator:
   ```bash
   go run ./internal/container/test generate \
     --source=/tmp/test/source \
     --librarian=/tmp/test/librarian \
     --input=/tmp/test/input \
     --output=/tmp/test/output
   ```

3. Verify output:
   ```bash
   ls -la /tmp/test/output
   # Should contain: client.go, doc.go, README.md, version.go
   ```

## Differences from Real Generators

The test container differs from real generators like `librariangen` in these ways:

1. **No actual code generation** - Creates simple placeholder files instead of running protoc
2. **No external dependencies** - Doesn't require protoc, plugins, or other toolchain components
3. **Deterministic output** - Always produces the same files for the same input
4. **Simplified validation** - Basic checks only, not full schema validation
5. **Fast execution** - Completes in milliseconds instead of seconds/minutes

## Use Cases

- **Librarian CLI development** - Test orchestration logic without real generators
- **Integration testing** - Verify end-to-end workflows with predictable output
- **Documentation examples** - Simple examples for how containers work
- **CI/CD testing** - Fast, reliable tests without heavy dependencies
- **Debugging** - Isolate Librarian behavior from generator complexity
