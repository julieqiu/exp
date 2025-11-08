# Using the Test Container with Librarian

This guide shows how to use the test container for developing and testing Librarian workflows without needing a real language generator toolchain.

## Initial Setup

### 1. Initialize a test repository

```bash
librarian init test
```

This command:
- Creates `.librarian/config.yaml` with test container configuration
- Builds the test container Docker image and tags it as `librarian-test:latest`
- Sets up the repository for both generation and release capabilities

**Example** `.librarian/config.yaml`:

```yaml
librarian:
  version: v0.5.0
  language: test

generate:
  container:
    image: librarian-test
    tag: latest
  googleapis:
    repo: github.com/googleapis/googleapis
    ref: main
  discovery:
    repo: github.com/googleapis/discovery-artifact-manager
    ref: main
  dir: generated/

release:
  tag_format: '{name}-v{version}'
```

### 2. Add an artifact for management

```bash
librarian add secretmanager google/cloud/secretmanager/v1
```

This command:
- Creates `secretmanager/.librarian.yaml` with generation and release sections
- Records the API path (`google/cloud/secretmanager/v1`)
- Syncs the current config values (container image, googleapis ref, etc.)

**Example** `secretmanager/.librarian.yaml`:

```yaml
generate:
  apis:
    - path: google/cloud/secretmanager/v1
  commit: null
  librarian: v0.5.0
  container:
    image: librarian-test
    tag: latest
  googleapis:
    repo: github.com/googleapis/googleapis
    ref: main
  discovery:
    repo: github.com/googleapis/discovery-artifact-manager
    ref: main

release:
  version: null
```

### 3. Generate the artifact

```bash
librarian generate secretmanager
```

This command:
- Mounts the test container with the appropriate directories
- Passes `generate-request.json` to the container
- Runs the test container's `generate` command
- Copies output from the container to `secretmanager/`
- Updates `secretmanager/.librarian.yaml` with the current commit

The test container creates these placeholder files:
- `secretmanager/client.go` - Mock client implementation
- `secretmanager/doc.go` - Package documentation
- `secretmanager/README.md` - Library documentation
- `secretmanager/version.go` - Version constant

## Development Workflow

### Rebuilding the test container

During development, you'll modify the test container code and need to rebuild it. Here are the recommended approaches:

#### Option 1: Manual rebuild (recommended for quick iteration)

```bash
# From the repository root
docker build -t librarian-test:latest -f internal/container/test/Dockerfile internal/container/test
```

Then regenerate your artifacts:

```bash
librarian generate secretmanager
```

This workflow is fast because:
- You control when the container rebuilds
- Docker uses build cache effectively
- You can test multiple times without rebuilding

#### Option 2: Rebuild via librarian init

```bash
librarian init test
```

This rebuilds the container image, but also resets `.librarian/config.yaml` to defaults. Use this when you want to verify the full initialization workflow.

#### Option 3: Add a rebuild command (future enhancement)

```bash
# Not yet implemented
librarian config rebuild-container
```

This would rebuild the container without affecting config.yaml.

### Testing container changes

1. **Modify the test container code:**
   ```bash
   # Edit files in internal/container/test/
   vim internal/container/test/generate/generate.go
   ```

2. **Rebuild the container:**
   ```bash
   docker build -t librarian-test:latest -f internal/container/test/Dockerfile internal/container/test
   ```

3. **Test with an artifact:**
   ```bash
   librarian generate secretmanager
   ```

4. **Verify the output:**
   ```bash
   ls -la secretmanager/
   cat secretmanager/client.go
   ```

### Common tasks

#### Add multiple APIs to an artifact

```bash
librarian add storage google/storage/v1,google/storage/v2
```

The test container will process both API paths and include them in the generated output.

#### Regenerate all artifacts

```bash
librarian generate --all
```

This regenerates every artifact that has a `generate` section in its `.librarian.yaml`.

#### Configure artifact settings

```bash
# Keep custom files during generation
librarian edit secretmanager --keep README.custom.md --keep scripts/

# Remove temporary files after generation
librarian edit secretmanager --remove .build-temp

# Exclude files from releases
librarian edit secretmanager --exclude tests/
```

#### Test the release workflow

```bash
# Prepare a release
librarian prepare secretmanager

# Publish the release (creates git tag)
librarian release secretmanager
```

## Testing Different Scenarios

### Test with local googleapis

If you want to test with a local checkout of googleapis instead of fetching from GitHub:

1. **Update config to use a local path:**
   ```bash
   librarian config set generate.googleapis.repo ../googleapis
   ```

2. **Regenerate:**
   ```bash
   librarian generate secretmanager
   ```

The test container will be mounted with your local googleapis directory.

### Test with specific commits

```bash
# Set a specific googleapis commit
librarian config set generate.googleapis.ref a1b2c3d4e5f6

# Regenerate to use that commit
librarian generate secretmanager
```

### Test container commands directly

You can run the test container directly to debug or test specific behavior:

```bash
# Prepare input
mkdir -p /tmp/test/{source,librarian,input,output}

cat > /tmp/test/librarian/generate-request.json <<'EOF'
{
  "id": "secretmanager",
  "version": "0.1.0",
  "apis": [{"path": "google/cloud/secretmanager/v1"}],
  "source_roots": ["secretmanager"]
}
EOF

# Run container directly
docker run --rm \
  -v /tmp/test/source:/source:ro \
  -v /tmp/test/librarian:/librarian:ro \
  -v /tmp/test/input:/input:ro \
  -v /tmp/test/output:/output \
  librarian-test:latest \
  generate \
  --source=/source \
  --librarian=/librarian \
  --input=/input \
  --output=/output

# Check output
ls -la /tmp/test/output
```

## Switching Between Test and Real Generators

### Switch from test to a real generator

```bash
# Initialize with a real language
librarian init python

# Update existing artifacts to use the new generator
librarian generate --all
```

This regenerates all artifacts with the Python generator instead of the test container.

### Switch back to test

```bash
# Reinitialize with test
librarian init test

# Rebuild container
docker build -t librarian-test:latest -f internal/container/test/Dockerfile internal/container/test

# Regenerate with test container
librarian generate --all
```

## Troubleshooting

### Container image not found

If you see `Error: docker image not found: librarian-test:latest`:

```bash
# Rebuild the container
docker build -t librarian-test:latest -f internal/container/test/Dockerfile internal/container/test
```

### Container fails to run

Check container logs:

```bash
# Run container with verbose logging
docker run --rm \
  -e GOOGLE_SDK_GO_LOGGING_LEVEL=debug \
  librarian-test:latest \
  --version
```

### Output directory is empty

The test container writes to `/output` inside the container. Verify your volume mounts are correct:

```bash
# Check that librarian is mounting directories correctly
librarian generate secretmanager --verbose
```

### Changes not reflected after rebuild

Docker may be using cached layers. Force a rebuild:

```bash
docker build --no-cache -t librarian-test:latest -f internal/container/test/Dockerfile internal/container/test
```

## Advanced Usage

### Testing the configure command

The test container also implements a `configure` command for validating library configurations:

```bash
# Not yet exposed via librarian CLI
# Run directly with docker for testing
docker run --rm \
  -v $(pwd):/repo:ro \
  -v /tmp/librarian:/librarian:ro \
  -v /tmp/output:/output \
  librarian-test:latest \
  configure \
  --repo=/repo \
  --librarian=/librarian \
  --output=/output
```

### Testing the build command

The test container's `build` command simulates running tests and builds:

```bash
# Not yet exposed via librarian CLI
docker run --rm \
  -v $(pwd):/repo:ro \
  -v /tmp/librarian:/librarian:ro \
  librarian-test:latest \
  build \
  --repo=/repo \
  --librarian=/librarian
```

### Testing release-stage command

The `release-stage` command updates version and changelog files:

```bash
# Prepare a release (this eventually calls release-stage)
librarian prepare secretmanager
```

## Best Practices

1. **Rebuild container after significant changes** - Always rebuild after modifying the test container's core logic

2. **Use version tags for stability** - Tag specific versions of the test container:
   ```bash
   docker build -t librarian-test:v0.1.0 -f internal/container/test/Dockerfile internal/container/test
   librarian config set generate.container.tag v0.1.0
   ```

3. **Keep test output in gitignore** - Add generated test files to `.gitignore` to avoid committing placeholder code

4. **Test with minimal APIs first** - Start with a single API path before testing multi-API artifacts

5. **Verify container contract** - Ensure the test container follows the same contract as real generators (same flags, same JSON formats, same directory structure)
