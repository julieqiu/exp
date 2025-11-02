# Librarian

Librarian manages the lifecycle of client libraries and release artifacts, from
code generation to version tagging and publishing.

## Getting Started

### Initialize a repository

```
librarian init <language>
```

If `language` is provided, `librarian` will configure the repository for code
generation. Supported languages are:
  - go
  - python
If `language` is not provided, `librarian` will configure the repository for
releases only.

**Example: Releases only**

```
librarian init
```

Creates `.librarian/config.yaml`:

```yaml
librarian: <librarian version>

release:
  tag_format: '{package}-v{version}'
```

**Example: Go with code generation**

```
librarian init go
```

Creates `.librarian/config.yaml` with a `generator` section:

```yaml
librarian: <librarian version>

generator:
  language: go
  image: <full-image-url-with-tag>
  googleapis: <commit-sha>
  discovery: <commit-sha>

release:
  tag_format: '{package}-v{version}'
```

## Creating Client Libraries

```
librarian create [artifact] [api-path]
```

Creates a `.librarian.yaml` file in the artifact's directory and runs code
generation. The artifact state is automatically synced with the current config.

Each artifact has its own `.librarian.yaml` file:

```yaml
# <artifact-path>/.librarian.yaml
generated_at:
  commit: <sha>
  apis:
    - path: <api path>
released_at:
  commit: <sha|nil>
  tag: <tag|nil>
```

## Updating Client Libraries

### Update an existing client library

```
librarian update [artifact]
```

Regenerates the artifact and automatically syncs its `.librarian.yaml` file
with the current config (librarian version, image, googleapis SHA, etc.).

### Update all client libraries

```
librarian update --all
```

Scans for all `.librarian.yaml` files and regenerates all artifacts. Each
artifact's state is automatically synced with the current config.

## Releasing Artifacts

### Release an artifact

```
librarian release <artifact-path>
```

Adds release metadata to the artifact's `.librarian.yaml` file and creates
release files (such as CHANGELOG.md).

```yaml
# <artifact-path>/.librarian.yaml
released_at:
  tag: <version>      # next planned release version
  commit: <sha>           # commit to be tagged
```

### Release all artifacts

```
librarian release --all
```

Scans for all `.librarian.yaml` files and updates release metadata for
artifacts with a release section.

Scans for all `.librarian.yaml` files and creates git tags for all artifacts
where `released_at` has a tag. Updates `commit` in `released_at` after each
tag is created.


## Deleting Client Libraries

```
librarian delete [artifact]
```

Removes an artifact from librarian management. Deletes the `.librarian.yaml`
file from the artifact's directory.

## Managing Configuration

### Update versions in config.yaml

```
librarian config update
```

Fetches and updates `.librarian/config.yaml` with the latest versions of:
- Googleapis commit SHA (if generate config exists)
- Discovery artifact manager commit SHA (if generate config exists)
- Language Container Image (if generate config exists)

### Set a configuration value

```
librarian config set <key> <value>
```

Sets a specific configuration value in `.librarian/config.yaml`. Supported keys:
- `librarian`
- `generator.language`
- `generator.image`
- `generator.googleapis`
- `generator.discovery`
- `release.tag_format`
To update the generator image, run:

```
librarian config set generator.image <image>
```
