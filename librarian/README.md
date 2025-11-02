# Librarian

Librarian automates the maintenance and release of versioned directories in a
repository.  A directory managed by Librarian may contain either generated code
(for a client library) or handwritten code (for a tool or service).

Librarian records generation input, release state, and version history, and
provides commands to regenerate and release the code in a repeatable way.

## Overview

### Librarian

**Core commands**

- [librarian init](#repository-setup): Initialize repository for library management
- [librarian add](#managing-directories): Track a directory for management
- [librarian edit](#editing-artifact-configuration): Edit artifact configuration (keep, remove, exclude, dir)
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

Creates a new `.librarian.yaml`.

If a language is specified, Librarian also sets up the code generation
environment for that language:

| Language   | Behavior                                           |
| ---------- | -------------------------------------------------- |
| **go**     | Builds the Go generator container using Docker     |
| **python** | Builds the Python generator container using Docker |
| **rust**   | Installs generator dependencies locally            |
| **dart**   | Installs generator dependencies locally            |

**Example: setting up a release-only repository**

```bash
librarian init
```

Produces:

```yaml
librarian: <librarian version>

release:
  tag_format: '{package}-v{version}'
```

**Example: Python repository with code generation**

```bash
librarian init [language]
```

Produces:

```yaml
librarian: <librarian version>

generator:
  language: python
  googleapis: <commit at latest>
  discovery: <commit at latest>

release:
  tag_format: '{package}-v{version}'
```

## Managing Directories

### Adding a Directory

```bash
librarian add <path> [api]
```

`add` tells Librarian to track the directory at `path`. If `api` is provided,
Librarian also records the API for code generation.

This creates `<path>/.librarian.yaml`:

```yaml
generate:        # present only if api is provided
  commit: <sha>
  apis:
    - path: <api>

release:
  version: null
```

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

Configure how files are handled during generation and release:

**Keep files during generation:**

```bash
librarian edit <path> --keep README.md --keep docs/
```

Files and directories in the keep list are not overwritten during code generation.

**Remove files after generation:**

```bash
librarian edit <path> --remove temp.txt --remove build/
```

Files in the remove list are deleted after code generation completes.

**Exclude files from release:**

```bash
librarian edit <path> --exclude tests/ --exclude .gitignore
```

Files in the exclude list are not included when creating releases.

**Set generation directory:**

```bash
# Override where generated code is written (default: "generated")
librarian edit <path> --dir src/
```

This overrides the global default set in `.librarian/config.yaml`.

**View current configuration:**

```bash
librarian edit <path>
```

Running `edit` without flags displays the current configuration for the artifact.

## Generating a Client Library

For directories with code generation configured:

```bash
librarian generate <path>
```

Updates generated code using the tool versions at `.librarian/config.yaml`.
Librarian updates `.librarian.yaml` automatically.

`--commit` writes a standard commit message for the change.

Regenerate all tracked directories:

```bash
librarian generate --all
```

## Releasing

### Preparing a Release

```bash
librarian prepare <path>
```

Determines the next version, updates metadata, and prepares release notes.
Does not tag or publish.

`<path>/.librarian.yaml` is updated:

```yaml
release:
  version: <current>
  prepared:
    version: <next>
    commit: <sha>
```

Prepare all tracked directories:

```bash
librarian prepare --all
```

`--commit` writes a standard commit message for the change.

### Publishing a Release

```bash
librarian release <path>
```

Tags the prepared version and updates recorded release state. If no prepared
release exists, the command does nothing.

Release all prepared directories:

```bash
librarian release --all
```

After release, the `release.prepared` section is removed:

```yaml
release:
  version: <new version>
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
- `generate.dir` - Default generation directory (default: "generated")
- `release.tag_format`

**Example: Set global generation directory**

```bash
# Set default generation directory for all artifacts
librarian config set generate.dir generated/
```

Artifacts can override this default using `librarian edit --dir`.

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
