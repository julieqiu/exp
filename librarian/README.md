# Librarian

Librarian automates the maintenance and release of versioned directories in a
repository.  A directory managed by Librarian may contain either generated code
(for a client library) or handwritten code (for a tool or service).

Librarian records generation input, release state, and version history, and
provides commands to regenerate and release the code in a repeatable way.

## Commands

**Setup**
- librarian init [language]

**Manage directories**
- librarian add <path> [api-path]
- librarian remove <path>
- librarian edit <path>
- librarian list

**Generate code**

- librarian generate [<path> | --all] [--latest] [--commit]

**Stage a release**

- librarian stage [<path> | --all] [--notes <file>] [--commit]

**Publish a release**

- librarian release [<path> | --all] [--tag-format]

**Inspection**

- librarian status [path]
- librarian history [path]

**Automation**

- librarian automate generate
- librarian automate stage
- librarian automate release

**Configuration**

- librarian config get <key>
- librarian config set <key> <value>
- librarian config unset <key>
- librarian config update [key | --all]

## Repository Setup

```
librarian init <language>
```

`librarian init` creates `.librarian/config.yaml`.

If language is provided, Librarian also configures code generation for that
language. Supported languages:
  - go
  - python

**Example: setting up a release-only repository**

```
librarian init
```

Produces:

```
librarian: <librarian version>

release:
  tag_format: '{package}-v{version}'
```

**Example: Python repository with code generation

```
librarian init [language]
```

Produces:

```
librarian: <librarian version>

generator:
  language: python
  googleapis: <commit at latest>
  discovery: <commit at latest>

release:
  tag_format: '{package}-v{version}'
```

## Tracking a Directory

```
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

## Regeneration

For directories with code generation configured:

```
librarian generate <path>
```

Updates generated code using the tool versions at `.librarian/config.yaml`.
Librarian updates .librarian.yaml automatically.

`--commit` writes a standard commit message for the change.

Regenerate all tracked directories:

```
librarian generate --all
```

## Staging a Release

```
librarian stage <path>
```

Determines the next version, updates metadata, and prepares release notes.
Does not tag or publish.

`<path>/.librarian.yaml` is updated:

```
release:
  version: <current>
  staged:
    version: <next>
    commit: <sha>
```

Stage all tracked directories:

```
librarian stage --all
```

`--commit` writes a standard commit message for the change.

## Releasing

```
librarian release <path>
```

Tags the staged version and updates recorded release state. If no staged
release exists, the command does nothing.

Release all staged directories:

```
librarian release --all
```

After release, the `release.staged` section is removed:

```
release:
  version: <new version>
```

## Removing a Directory

```
librarian remove <path>
```

Removes `<path>/.librarian.yaml`. Source code is not modified.

## Configuration

### Update versions in config.yaml

Update toolchain information to latest:

```
librarian config update [key]
librarian config update --all
```

Supported keys:
- `generator.image`
- `generator.googleapis`
- `generator.discovery`


Set a configuration key explicitly:

```
librarian config set <key> <value>
```

Supported keys:
- `generator.language`
- `generator.image`
- `generator.googleapis`
- `generator.discovery`
- `release.tag_format`

## Librarian Automation

Automation follows three phases:

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

## Notes

- Librarian does not modify code outside the tracked directories.
- Librarian records only information required for reproducibility and release
  automation.
- The system is designed so that `git log` and `.librarian.yaml` describe the
  full history of generation inputs and release versions.
