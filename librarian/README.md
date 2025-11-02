# Librarian

Librarian automates the maintenance and release of versioned directories in a
repository. A directory managed by Librarian may contain either generated code
(for a client library) or handwritten code (for a tool or service).

**Repository model**: Each repository supports a single language (Go, Python, Rust, or Dart) and can contain multiple artifacts for that language. Repository capabilities are determined by which sections exist in `.librarian/config.yaml`:

- `generate` section present → repository supports code generation
- `release` section present → repository supports release management
- Both sections present → repository supports both

Each artifact can independently have generation and/or release enabled based on which sections are present in its `.librarian.yaml` file.

Librarian records generation input, release state, and version history, and
provides commands to regenerate and release the code in a repeatable way.

## Overview

### Librarian

**Core commands**

- [librarian init](#repository-setup): Initialize repository for library management
- [librarian add](#managing-directories): Track a directory for management
- [librarian edit](#editing-artifact-configuration): Edit artifact configuration (name, keep, remove, exclude)
- [librarian remove](#removing-a-directory): Stop tracking a directory
- [librarian generate](#generating-a-client-library): Generate or regenerate code for tracked directories
- [librarian prepare](#preparing-a-release): Prepare a release with version updates and notes
- [librarian release](#publishing-a-release): Tag and publish a prepared release

**Configuration commands**

- [librarian config get](#configuration): Read a configuration value
- [librarian config set](#configuration): Set a configuration value
- [librarian config update](#configuration): Update toolchain versions to latest

**Inspection commands**

- [librarian list](#inspection): List all tracked directories
- [librarian status](#inspection): Show generation and release status
- [librarian history](#inspection): View release history

### Librarianops

**Automation commands**

- [librarianops generate](#automate-code-generation): Automate code generation workflow
- [librarianops prepare](#automate-release-preparation): Automate release preparation workflow
- [librarianops release](#automate-release-publishing): Automate release publishing workflow

## Repository Setup

```bash
librarian init [language]
```

Initializes a repository for library management. Repository capabilities are determined by which sections are created.

**Languages supported:**
- `go` - Builds the Go generator container using Docker
- `python` - Builds the Python generator container using Docker
- `rust` - Installs generator dependencies locally
- `dart` - Installs generator dependencies locally

**Example: Release-only repository**

```bash
librarian init
```

Creates `.librarian/config.yaml`:

```yaml
librarian:
  version: v0.5.0

release:
  tag_format: '{name}-v{version}'
```

**What this enables:**
- `librarian add <path>` - Track handwritten code for release
- `librarian prepare <path>` - Prepare releases
- `librarian release <path>` - Publish releases

**Example: Repository with code generation and releases**

```bash
librarian init python
```

Creates `.librarian/config.yaml`:

```yaml
librarian:
  version: v0.5.0
  language: python

generate:
  image: us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator:latest
  googleapis: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
  discovery: f9e8d7c6b5a4f3e2d1c0b9a8f7e6d5c4b3a2f1e0
  dir: generated

release:
  tag_format: '{name}-v{version}'
```

**What this enables:**
- `librarian add <path> <api>` - Generate code from API definitions
- `librarian generate <path>` - Regenerate code
- `librarian prepare <path>` - Prepare releases
- `librarian release <path>` - Publish releases

**Note**: The presence of the `generate` section enables generation commands. The presence of the `release` section enables release commands.

## Managing Directories

### Adding a Directory

```bash
librarian add <path> [api]
```

Tracks a directory for management. The sections created in `<path>/.librarian.yaml` depend on:
1. Which sections exist in `.librarian/config.yaml`
2. Whether an API is provided

**In a release-only repository** (no `generate` section in config):
```bash
librarian add packages/my-tool
```

Creates `<path>/.librarian.yaml`:
```yaml
release:
  version: null
```

**In a repository with generation** (has `generate` section in config):
```bash
# Add handwritten code (no API)
librarian add packages/my-tool
```

Creates `<path>/.librarian.yaml`:
```yaml
release:
  version: null
```

```bash
# Add generated code (with API)
librarian add packages/storage google/storage/v1
```

Creates `packages/storage/.librarian.yaml`:
```yaml
generate:
  apis:
    - path: google/storage/v1
  commit: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
  librarian: v0.5.0
  image: us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator:latest
  googleapis-sha: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
  discovery-sha: f9e8d7c6b5a4f3e2d1c0b9a8f7e6d5c4b3a2f1e0

release:
  version: null
```

**Note**: The `generate` section is only created when an API is provided AND the repository has a `generate` section in its config. The `release` section is created if the repository has a `release` section in its config.

`--commit` writes a standard commit message for the change.

### Removing a Directory

```bash
librarian remove <path>
```

Removes `<path>/.librarian.yaml`. Source code is not modified.

## Editing Artifact Configuration

```bash
librarian edit <path> [flags]
```

Configure artifact-specific settings:

**Set language-specific metadata:**

The language metadata should match the repository's language (set via `librarian init`).

```bash
# For Go repositories
librarian edit <path> --language go:module=github.com/user/repo

# For Python repositories
librarian edit <path> --language python:package=my-package

# For Rust repositories
librarian edit <path> --language rust:crate=my_crate

# For Dart repositories
librarian edit <path> --language dart:package=my_package
```

Language-specific metadata is used by generators and tooling for proper package/module configuration. The format is `--language LANG:KEY=VALUE` where LANG matches your repository's language and KEY is the property name (module, package, or crate).

**Keep files during generation:**

```bash
librarian edit <path> --keep README.md --keep docs
```

Files and directories in the keep list are not overwritten during code generation.

**Remove files after generation:**

```bash
librarian edit <path> --remove temp.txt --remove build
```

Files in the remove list are deleted after code generation completes.

**Exclude files from release:**

```bash
librarian edit <path> --exclude tests --exclude .gitignore
```

Files in the exclude list are not included when creating releases.

**View current configuration:**

```bash
librarian edit <path>
```

Running `edit` without flags displays the current configuration for the artifact.

## Generating a Client Library

For artifacts with a `generate` section in their `.librarian.yaml`:

```bash
librarian generate <path>
```

Generates or regenerates code using the tool versions from `.librarian/config.yaml`.
Librarian updates the artifact's `.librarian.yaml` automatically.

`--commit` writes a standard commit message for the change.

Regenerate all artifacts that have a `generate` section:

```bash
librarian generate --all
```

**Note**: This command only works in repositories that have a `generate` section in `.librarian/config.yaml`, and only affects artifacts that have a `generate` section in their `.librarian.yaml`.

## Releasing

### Preparing a Release

For artifacts with a `release` section in their `.librarian.yaml`:

```bash
librarian prepare <path>
```

Determines the next version, updates metadata, and prepares release notes.
Does not tag or publish.

`packages/storage/.librarian.yaml` is updated:

```yaml
release:
  version: v1.2.0
  prepared:
    version: v1.3.0
    commit: e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3
```

Prepare all artifacts that have a `release` section:

```bash
librarian prepare --all
```

`--commit` writes a standard commit message for the change.

**Note**: This command only works in repositories that have a `release` section in `.librarian/config.yaml`, and only affects artifacts that have a `release` section in their `.librarian.yaml`.

### Publishing a Release

For artifacts with a `release` section and a prepared release:

```bash
librarian release <path>
```

Tags the prepared version and updates recorded release state. If no prepared
release exists, the command does nothing.

Release all prepared artifacts:

```bash
librarian release --all
```

After release, the `release.prepared` section is removed:

```yaml
release:
  version: v1.3.0
```

## Configuration

### Update versions in config.yaml

Update toolchain information to latest:

```bash
librarian config update [key]
librarian config update --all
```

Supported keys:

- `generator.image`
- `generator.googleapis`
- `generator.discovery`

Set a configuration key explicitly:

```bash
librarian config set <key> <value>
```

Supported keys:

- `generator.language`
- `generator.image`
- `generator.googleapis`
- `generator.discovery`
- `generate.dir` - Default generation directory (default: \"generated\")
- `release.tag_format`

**Example: Set global generation directory**

```bash
librarian config set generate.dir generated
```

## Inspection

View information about tracked directories and their release history.

List all tracked directories:

```bash
librarian list
```

Show the current generation and release status for a directory:

```bash
librarian status <path>
```

View the release history for a directory:

```bash
librarian history <path>
```

## Automation with librarianops

The `librarianops` command automates common librarian workflows for CI/CD pipelines.

### Configuration

**Flags:**

- `--project` - GCP project ID (default: `cloud-sdk-librarian-prod`)
- `--dry-run` - Print commands without executing them

```bash
# Use custom project
librarianops --project my-project generate

# Dry run to see what would be executed
librarianops --dry-run generate
```

### Automate code generation

```bash
librarianops generate
```

This runs:
1. `librarian config update --all --commit` - Update to latest versions
2. `librarian generate --all --commit` - Regenerate all artifacts
3. `gh pr create --with-token=$(fetch token) --fill` - Create pull request

### Automate release preparation

```bash
librarianops prepare
```

This runs:
1. `librarian prepare --all --commit` - Prepare all artifacts
2. `gh pr create --with-token=$(fetch token) --fill` - Create pull request

### Automate release publishing

```bash
librarianops release
```

This runs:
1. `librarian release --all` - Release all prepared artifacts
2. `gh release create --with-token=$(fetch token) --notes-from-tag` - Create GitHub releases

## Notes

- Librarian does not modify code outside the tracked directories.
- Librarian records only information required for reproducibility and release
  automation.
- The system is designed so that `git log` and `.librarian.yaml` describe the
  full history of generation inputs and release versions.
