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
librarian:
  version: <librarian version>
  mode: <mode>

release:
  tag_format: '{package}-v{version}'
```

For generation modes (go, python), additional fields will be added when needed:

```yaml
librarian:
  version: <librarian version>
  mode: <mode>

generate:
  image: <full-image-url-with-tag>
  googleapis: <commit-sha>
  discovery: <commit-sha>
  custom:
    - key: value

release:
  tag_format: '{package}-v{version}'
```

## Managing Configuration

### Update versions in config.yaml

```
librarian config update
```

Fetches and updates `.librarian/config.yaml` with the latest versions of:
- Librarian version
- Googleapis commit SHA (if generate config exists)
- Discovery artifact manager commit SHA (if generate config exists)

Add `--sync` to also regenerate all artifacts and sync their `.librarian.yaml` files:

```
librarian config update --sync
```

### Set a configuration value

```
librarian config set <key> <value>
```

Sets a specific configuration value in `.librarian/config.yaml`. Supported keys:
- `version`
- `mode`
- `release.tag_format`
- `generate.image`
- `generate.googleapis`
- `generate.discovery`

Add `--sync` to also regenerate all artifacts after setting the config:

```
librarian config set <key> <value> --sync
```

## Generating Client Libraries

### Generate a new artifact

```
librarian generate <artifact-path> <api>
```

Creates a `.librarian.yaml` file in the artifact's directory and runs code
generation. The artifact state is automatically synced with the current config.

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

Regenerates the artifact and automatically syncs its `.librarian.yaml` file with the current config (librarian version, image, googleapis SHA, etc.).

### Regenerate all artifacts

```
librarian generate --all
```

Scans for all `.librarian.yaml` files and regenerates all artifacts. Each artifact's state is automatically synced with the current config.

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
