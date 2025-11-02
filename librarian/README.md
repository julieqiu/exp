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

## Adding Client Libraries

librarian add <library-path> [api-path]
```

Adds a library to librarian management by creating a `.librarian.yaml` file in the library's directory and running code generation. The library state is automatically synced with the current config.

Each library has its own `.librarian.yaml` file:

```yaml
# <library-path>/.librarian.yaml
generated_at:
  commit: <sha>
  apis:
    - path: <api path>
release:
  published:
    version: <tag|nil>
    commit: <sha|nil>
```

The `--commit` flag can be used to run `git commit` with a preformatted commit
message.

## Generating Client Libraries


### Generate an existing client library

```
librarian generate <library-path>
```

Alias: `librarian gen`

Regenerates the library and automatically syncs its `.librarian.yaml` file
with the current config (librarian version, image, googleapis SHA, etc.).

The `--commit` flag can be used to run `git commit` with a preformatted commit
message.

### Generate all client libraries

```
librarian generate --all
```

Scans for all `.librarian.yaml` files and regenerates all libraries. Each
library's state is automatically synced with the current config.

Use the `--latest` flag to first update `.librarian/config.yaml` to the latest
version of the config.




## Staging Libraries

### Stage a library for release

```
librarian stage <library-path>
```

Adds release metadata to the library's `.librarian.yaml` file and creates
release files (such as CHANGELOG.md).

The `release` section of the `.librarian.yaml` is updated:
```yaml
# <library-path>/.librarian.yaml
release:
  staged:
    version: <version>    # next planned release version
    commit: <sha>         # commit to be tagged
  published:
    version: <version>    # last released version
    commit: <sha>         # last released commit
```

The `--commit` flag can be used to run `git commit` with a preformatted commit
message.

### Stage all libraries for release

```
librarian stage --all
```

Scans for all `.librarian.yaml` files and updates release metadata for
libraries with a release section.

## Publishing Libraries

### Tag a staged library

```
librarian publish <library-path>
```

Creates a git tag for the staged library and publishes it. On success, the
`release.staged` section is cleared and `release.published` is updated with the
new version and commit. Skips if the git tag already exists.

### Tag all staged libraries

```
librarian publish --all
```

Scans for all `.librarian.yaml` files and creates git tags for all libraries
where a `staged` section exists. On success, the `staged` section is removed
and `released` is updated with the new tag. Skips if the git tag already
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
librarian generate --all --latest --commit
gh pr create --with-token=$(fetch token) --fill
```

### Release Stage

```
librarian stage --all --commit
gh pr create --with-token=$(fetch token) --fill
```

### Release Publish

```
librarian publish --all
gh release create --with-token=$(fetch token) --notes-from-tag
```
