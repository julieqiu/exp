# Librarian

Librarian automates the maintenance and release of versioned directories in a
repository.  A directory managed by Librarian may contain either generated code
(for a client library) or handwritten code (for a tool or service).

Librarian records generation input, release state, and version history, and
provides commands to regenerate and release the code in a repeatable way.

## Commands

**Setup**

- `librarian init [language]`

**Manage directories**

- `librarian add <path> [api-path]`
- `librarian remove <path>`
- `librarian edit <path>`
- `librarian list`

**Generate code**

- `librarian generate [<path> | --all] [--latest] [--commit]`

**Stage a release**

- `librarian stage [<path> | --all] [--notes <file>] [--commit]`

**Publish a release**

- `librarian release [<path> | --all] [--tag-format]`

**Inspection**

- `librarian status [path]`
- `librarian history [path]`

**Automation**

- `librarian automate generate`
- `librarian automate stage`
- `librarian automate release`

**Configuration**

- `librarian config get <key>`
- `librarian config set <key> <value>`
- `librarian config unset <key>`
- `librarian config update [key | --all]`

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

## Tracking a Directory

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

## Regeneration

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

## Staging a Release

```bash
librarian stage <path>
```

Determines the next version, updates metadata, and prepares release notes.
Does not tag or publish.

`<path>/.librarian.yaml` is updated:

```yaml
release:
  version: <current>
  staged:
    version: <next>
    commit: <sha>
```

Stage all tracked directories:

```bash
librarian stage --all
```

`--commit` writes a standard commit message for the change.

## Releasing

```bash
librarian release <path>
```

Tags the staged version and updates recorded release state. If no staged
release exists, the command does nothing.

Release all staged directories:

```bash
librarian release --all
```

After release, the `release.staged` section is removed:

```yaml
release:
  version: <new version>
```

## Removing a Directory

```bash
librarian remove <path>
```

Removes `<path>/.librarian.yaml`. Source code is not modified.

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
- `release.tag_format`

## Librarian Automation

Automation follows three phases:

### Generate

```bash
librarian config update --all --commit
librarian generate --all --commit
gh pr create --with-token=$(fetch token) --fill
```

### Stage

```bash
librarian stage --all --commit
gh pr create --with-token=$(fetch token) --fill
```

### Release

```bash
librarian release --all
gh release create --with-token=$(fetch token) --notes-from-tag
```

## Notes

- Librarian does not modify code outside the tracked directories.
- Librarian records only information required for reproducibility and release
  automation.
- The system is designed so that `git log` and `.librarian.yaml` describe the
  full history of generation inputs and release versions.
