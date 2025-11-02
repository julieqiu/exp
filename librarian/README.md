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

```
librarian:
  mode: <mode>
  version: <librarian version>
  release_tag_format: '{id}-v{version}'
```

Creates `.librarian/state.yaml`:

```
packages: {}
```

## Generating Client Libraries

### Generate a new package

```
librarian generate <package> <api>
```

This registers the package and API in .librarian/state.yaml, and runs code
generation.

```
packages:
  <package>:
    path: <path>              # path to package
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

### Regenerate an existing package

```
librarian generate <package>
```

The package must already exist in .librarian/state.yaml.

### Regenerate all packages

```
librarian generate --all
```

Runs generation for all packages with a `generate:` section.

## Releasing Packages

### Stage a new package for release

```
librarian release stage <package> <path>
```

Adds a handwritten package for release management. Creates release metadata
(such as CHANGELOG.md) and adds the release section to `.librarian/state.yaml`.

```
packages:
  <package>:
    path: <path>               # path to package
    release:
    last_released_at:
        tag: <tag|null>
        commit: <sha|null>
    next_released_at:
        version: <version>      # next planned release version
        commit: <sha>           # commit to be tagged
```

### Stage an existing package for release

```
librarian release stage <package>
```

Updates release metadata for a package already managed by librarian.

### Stage all packages for release

```
librarian release stage --all
```

Updates release metadata for all packages with a release section.

## Tagging Packages

### Tag a package

```
librarian release tag <package>
```

Creates a git tag if `next_released_at` is later than `last_released_at`.
Updates `last_released_at` after the tag is created. Skips if the git tag
already exists.

### Tag all packages

```
librarian release tag --all
```

Creates git tags for all packages where `next_released_at` is later than
`last_released_at`.
