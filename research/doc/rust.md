Rust GAPIC Generator Guide (Sidekick)
=====================================

Rust uses Sidekick, a standalone generator tool (not a protoc plugin), located in github.com/googleapis/librarian.

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/storage/v2/storage.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `storage_v2.yaml`\)

Configuration Options
---------------------

### Command-Line Flags

| Flag                    | Description                                  | Example                          |
|-------------------------|----------------------------------------------|----------------------------------|
| `-project-root`         | Base project directory                       | `..`                             |
| `-specification-format` | Input format: `protobuf`, `openapi`, `disco` | `protobuf`                       |
| `-specification-source` | Path to API specification files              | `google/cloud/secretmanager/v1`  |
| `-service-config`       | YAML service configuration file              | `service.yaml`                   |
| `-language`             | Target language                              | `rust`                           |
| `-output`               | Output directory                             | `./generated`                    |
| `-codec-option`         | Code generation key=value options            | `package-name-override=my_crate` |

### Codec Options (Rust-specific)

| Option                  | Description                   | Example                |
|-------------------------|-------------------------------|------------------------|
| `package-name-override` | Override generated crate name | `google_cloud_storage` |
| `version`               | Generated crate version       | `0.1.0`                |
| `release-level`         | Maturity: `stable`/`preview`  | `preview`              |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
git clone https://github.com/googleapis/librarian.git
```

Examples
--------

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis
```

### Secret Manager

```bash
cd librarian && go run cmd/sidekick/main.go generate \
  -project-root=.. \
  -specification-format=protobuf \
  -specification-source=google/cloud/secretmanager/v1 \
  -service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_v1.yaml \
  -source-option=googleapis-root=$GOOGLEAPIS \
  -language=rust \
  -output=rust/secretmanager \
  -codec-option=package-name-override=google_cloud_secretmanager
```

### Cloud Storage

```bash
cd librarian && go run cmd/sidekick/main.go generate \
  -project-root=.. \
  -specification-format=protobuf \
  -specification-source=google/storage/v2 \
  -service-config=$GOOGLEAPIS/google/storage/v2/storage_v2.yaml \
  -source-option=googleapis-root=$GOOGLEAPIS \
  -language=rust \
  -output=rust/storage \
  -codec-option=package-name-override=google_cloud_storage
```

Configuration File (.sidekick.toml)
-----------------------------------

Sidekick also supports TOML configuration:

```toml
[general]
language = "rust"
specification-format = "protobuf"
specification-source = "google/storage/v2"
service-config = "path/to/storage_v2.yaml"

[codec]
package-name-override = "google_cloud_storage"
version = "0.1.0"
release-level = "preview"
```

Post-Processing
---------------

Rust does **not** use OwlBot. Sidekick (part of Librarian) handles the complete code generation workflow, so no separate post-processing step is required.
