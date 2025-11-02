# Librarian

Librarian manages the lifecycle of client libraries and release libraries, from
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

#### Example: Releases only

```
librarian init
```

Creates `.librarian/config.yaml`:

```yaml
librarian: <librarian version>

release:
  tag_format: '{package}-v{version}'
```

#### Example: Go with code generation

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
librarian create [library-path] [api-path]
```

Creates a `.librarian.yaml` file in the library's directory and runs code
generation. The library state is automatically synced with the current config.

Each library has its own `.librarian.yaml` file:

```yaml
# <library-path>/.librarian.yaml
generated_at:
  commit: <sha>
  apis:
    - path: <api path>
released_at:
  tag: <tag|nil>
```

## Updating Client Libraries

### Update an existing client library

```
librarian update [library-path]
```

Regenerates the library and automatically syncs its `.librarian.yaml` file
with the current config (librarian version, image, googleapis SHA, etc.).

### Update all client libraries

```
librarian update --all
```

Scans for all `.librarian.yaml` files and regenerates all libraries. Each
library's state is automatically synced with the current config.

## Releasing Artifacts

### Release a library

```
librarian release <library-path>
```

Adds release metadata to the library's `.librarian.yaml` file and creates
release files (such as CHANGELOG.md).

```yaml
# <library-path>/.librarian.yaml
released_at:
  tag: <version>
```

### Release all libraries

```
librarian release --all
```

Scans for all `.librarian.yaml` files and updates release metadata for
libraries with a release section.

Scans for all `.librarian.yaml` files and creates git tags for all libraries
where `released_at` has a tag. Updates `commit` in `released_at` after each
tag is created.


## Deleting Client Libraries

```
librarian delete [library-path]
```

Removes a library from librarian management. Deletes the `.librarian.yaml`
file from the library's directory.

## Managing Configuration

### Update versions in config.yaml

```
librarian config update [key]
```

Fetches and updates `.librarian/config.yaml` with the latest versions of:

- `generator.image`
- `generator.googleapis`
- `generator.discovery`

```
librarian config update --all
```

Updates all values above.

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

## Automation

The automation infrastructure will run these commands:

### Generate

```
librarian config update --all
librarian update --all
```

### Release

```
librarian release --all
```
