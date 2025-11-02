# Librarian

Librarian manages the lifecycle of client libraries and release artifacts, from
code generation to version tagging and publishing.

## Getting Started

### Initialize a repository

```
librarian init [mode]
```

Supported modes are:
  - go: generate & release Go client libraries
  - python: generate & release Python client libraries
  - release-only: release only; for handwritten packages

Creates `.librarian/config.yaml`:

```yaml
version: <librarian version>
mode: <mode>
release_tag_format: '{package}-v{version}'
```

For generation modes (go, python), additional fields will be added when needed:

```yaml
generate:
  image: <full-image-url-with-tag>
  googleapis: <commit-sha>
  discovery: <commit-sha>
  custom:
    - key: value
```

## Managing Configuration

### Update all versions to latest

```
librarian config update
```

Fetches and updates the configuration with the latest versions of:
- Librarian version
- Googleapis commit SHA (if generate config exists)
- Discovery artifact manager commit SHA (if generate config exists)

After updating, regenerates all libraries to use the new versions. Use `--no-sync` to skip regeneration.

### Set a configuration value

```
librarian config set <key> <value>
```

Sets a specific configuration value in `.librarian/config.yaml`. Supported keys:
- `version`
- `mode`
- `release_tag_format`
- `generate.image`
- `generate.googleapis`
- `generate.discovery`

## Generating Client Libraries

### Generate a new artifact

```
librarian generate <artifact-path> <api>
```

Creates a `.librarian.yaml` file in the artifact's directory and runs code
generation.

Each artifact has its own `.librarian.yaml` file:

```yaml
# <artifact-path>/.librarian.yaml
generate:
  apis:
    - path: <api path>
  commit: <sha>
  librarian: <version>     # from config.yaml
  image: <container image> # from config.yaml
  googleapis-sha: <sha>    # listed if the api source is googleapis
  discovery-sha: <sha>     # listed if the api source is discovery docs
release:
  last_released_at:
    tag: <tag|nil>
    commit: <sha|nil>
```

### Regenerate an existing artifact

```
librarian generate <artifact-path>
```

The artifact must have a `.librarian.yaml` file in its directory.

### Regenerate all artifacts

```
librarian generate --all
```

Scans for all `.librarian.yaml` files and runs generation for artifacts with a `generate:` section.

### Remove an artifact

```
librarian remove <artifact-path>
```

Removes an artifact from librarian management. Deletes the `.librarian.yaml` file from the artifact's directory.

## Releasing Artifacts

### Stage a new artifact for release

```
librarian release stage <artifact-path>
```

Adds release metadata to the artifact's `.librarian.yaml` file and creates
release files (such as CHANGELOG.md).

```yaml
# <artifact-path>/.librarian.yaml
release:
  last_released_at:
    tag: <tag|null>
    commit: <sha|null>
  next_release_at:
    version: <version>      # next planned release version
    commit: <sha>           # commit to be tagged
```

### Stage an existing artifact for release

```
librarian release stage <artifact-path>
```

Updates release metadata for an artifact already managed by librarian.

### Stage all artifacts for release

```
librarian release stage --all
```

Scans for all `.librarian.yaml` files and updates release metadata for artifacts with a release section.

## Tagging Artifacts

### Tag an artifact

```
librarian release tag <artifact-path>
```

Creates a git tag if `next_released_at` is later than `last_released_at`.
Updates `last_released_at` and removes `next_released_at` after the tag is
created. Skips if the git tag already exists.

### Tag all artifacts

```
librarian release tag --all
```

Scans for all `.librarian.yaml` files and creates git tags for all artifacts
where `next_released_at` is later than `last_released_at`. Updates
`last_released_at` and removes `next_released_at` after each tag is created.
