# Librarian

Librarian manages the lifecycle of client libraries and other release
artifacts, from code generation to version tagging.

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

## Adding Client Libraries

```
librarian add <library-path> [api-path]
```

Registers a library for librarian management by creating a `.librarian.yaml`
file in the library's directory.

Each library has its own `.librarian.yaml` file:

```yaml
# <library-path>/.librarian.yaml
generate:               # only populated if generator section exists
  commit: <sha>
  apis:
    - path: <api path>
release:
  version: <tag|nil>
```

The `--commit` flag can be used to generate a preformatted commit message.

## Generating Client Libraries

### Generate an existing client library

```
librarian generate <library-path>
```

Alias: `librarian gen`

Regenerates the library and automatically syncs its `.librarian.yaml` file
with the current config (librarian version, image, googleapis SHA, etc.).

The `--commit` flag can be used to generate a preformatted commit message.

### Generate all client libraries

```
librarian generate --all
```

Scans for all `.librarian.yaml` files and regenerates all libraries. Each
library's state is automatically synced with the current config.

## Staging Libraries

### Stage a library for release

```
librarian stage <library-path>
```

Calculates the next version, updates metadata, and prepares release artifacts.

The `release` section of the `.librarian.yaml` is updated:

```yaml
# <library-path>/.librarian.yaml

release:
  version: <version>      # last tagged version
  staged:
    version: <version>    # next planned release version
    commit: <sha>         # commit to be tagged
```

The `--commit` flag can be used to generate a preformatted commit message.

### Stage all libraries for release

```
librarian stage --all
```

Scans for all `.librarian.yaml` files and updates release metadata for
libraries with a release section.

## Releasing Libraries

### Release a staged library

```
librarian release <library-path>
```

Creates a git tag for the staged library. On success, the
`release.staged` section is cleared and `release.version` is updated with the
new version and commit. Skips if the git tag already exists.

librarian release --all
```

Scans for all `.librarian.yaml` files and creates git tags for all libraries
where a `release.staged` section exists. On success, the `release.staged` section is cleared
and `release.version` is updated with the new version and commit. Skips if the git tag already
exists.

## Removing Client Libraries

```
librarian remove <library-path>
```

Removes a library from librarian management by deleting the `.librarian.yaml`
file from the library's directory. This does not delete the library's source code.

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
librarian config update --all --commit
librarian generate --all --commit
gh pr create --with-token=$(fetch token) --fill
```

### Stage

```
librarian stage --all --commit
gh pr create --with-token=$(fetch token) --fill
```

### Release

```
librarian release --all
gh release create --with-token=$(fetch token) --notes-from-tag
```
