# A Tour of librarianx

Librarianx is a tool for managing Google Cloud client libraries across multiple
languages. It handles code generation from API definitions, version management,
and publishing to package registries.

This tour walks through realistic workflows for Go, Python, and Rust libraries.
You'll see how to set up a repository, generate your first library, handle
updates, and publish releases.

## Installation

Start by installing librarianx:

```
$ go install github.com/julieqiu/librarianx@latest
```

## Your First Library: Go Secret Manager

Let's build a Go client library for Google Cloud Secret Manager. First,
create a workspace:

```
$ mkdir libraries
$ cd libraries
$ mkdir google-cloud-go
$ cd google-cloud-go
```

### Initialize the Repository

Initialize a Go repository with `librarianx init`:

```
$ librarianx init go
Created .librarian/config.yaml
```

This creates a repository configuration file. Let's see what's inside:

```
$ cat .librarian/config.yaml
librarian:
  version: v0.5.0
  language: go

generate:
  container:
    image: us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/go-librarian-generator
    tag: latest
  googleapis:
    path: https://github.com/googleapis/googleapis/archive/9fcfbea0aa5b50fa22e190faceb073d74504172b.tar.gz
    sha256: 81e6057ffd85154af5268c2c3c8f2408745ca0f7fa03d43c68f4847f31eb5f98
  dir: ./

release:
  tag_format: '{id}/v{version}'
```

The config defines how to generate code (container image, googleapis location)
and how to format release tags.

### Install Dependencies

Before generating code, install the Go generator dependencies. You can either
install them locally or use a Docker container:

```
$ librarianx install go --use-container
Using Docker container for code generation
Container image: us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/go-librarian-generator:latest
```

The `--use-container` flag ensures consistent generation across different
environments. You can omit it to install dependencies locally instead.

### Add Your First API

Add the Secret Manager API to your repository:

```
$ librarianx add secretmanager google/cloud/secretmanager/v1
Parsing googleapis BUILD.bazel files...
Created secretmanager/.librarian.yaml
```

This command:
1. Downloads googleapis (if needed)
2. Reads `google/cloud/secretmanager/v1/BUILD.bazel` to extract configuration
3. Creates an artifact config at `secretmanager/.librarian.yaml`

Notice that Go uses directory names without prefixes (secretmanager, not
google-cloud-secretmanager). This matches Go module conventions.

Let's look at what was created:

```
$ cat secretmanager/.librarian.yaml
generate:
  apis:
    - path: google/cloud/secretmanager/v1
      grpc_service_config: secretmanager_grpc_service_config.json
      service_yaml: secretmanager_v1.yaml
      transport: grpc+rest
  metadata:
    name_pretty: "Secret Manager"
    product_documentation: "https://cloud.google.com/secret-manager/docs"
    release_level: "stable"
  language:
    go:
      module: cloud.google.com/go/secretmanager

release:
  version: null
```

All the protoc configuration was extracted from BUILD.bazel and saved here.
This makes generation fast and reproducible.

### Generate the Library

Generate the Go client library:

```
$ librarianx generate secretmanager
Downloading googleapis...
Running generator container...
Applying file filters...
Generated secretmanager/
```

Let's see what was created:

```
$ ls secretmanager/
apiv1/
  secret_manager_client.go
  secret_manager_client_test.go
go.mod
go.sum
README.md
```

Your first client library is ready!

### Add Another API Version

Secret Manager has a beta API. Let's add it to the same library:

```
$ librarianx add secretmanager google/cloud/secretmanager/v1beta2
Updated secretmanager/.librarian.yaml
```

Notice this updated the existing artifact config instead of creating a new one.
The artifact now has two APIs:

```
$ cat secretmanager/.librarian.yaml | grep "path:"
    - path: google/cloud/secretmanager/v1
    - path: google/cloud/secretmanager/v1beta2
```

Regenerate to include the beta API:

```
$ librarianx generate secretmanager
Running generator container...
Generated secretmanager/
```

Now you have both stable and beta APIs in one module:

```
$ ls secretmanager/
apiv1/
  secret_manager_client.go
apiv1beta2/
  secret_manager_client.go
go.mod
go.sum
README.md
```

### Release Your Library

To release this library, you'll run three commands: `prepare`, `tag`, and `publish`.

First, make some changes and commit:

```
$ echo "# Custom documentation" >> secretmanager/README.md
$ git add .
$ git commit -m "feat(secretmanager): add Secret Manager client library"
```

Prepare the release:

```
$ librarianx release prepare
Detected 1 library with pending releases:
  - secretmanager: null → 0.1.0 (initial release)

Updated files:
  secretmanager/.librarian.yaml
  secretmanager/CHANGELOG.md

Created commit: chore(release): prepare secretmanager v0.1.0
```

This detected that the library has never been released (version: null) and
bumped it to v0.1.0. Let's see what changed:

```
$ git show HEAD
commit abc123...
Author: You <you@example.com>
Date:   Mon Jan 15 10:00:00 2025

    chore(release): prepare secretmanager v0.1.0

diff --git a/secretmanager/.librarian.yaml b/secretmanager/.librarian.yaml
index def456..abc123 100644
--- a/secretmanager/.librarian.yaml
+++ b/secretmanager/.librarian.yaml
@@ -15,4 +15,5 @@ generate:
       module: cloud.google.com/go/secretmanager

 release:
-  version: null
+  version: 0.1.0
```

Now create git tags:

```
$ librarianx release tag
Creating tags for 1 library:

secretmanager/v0.1.0
  Tag created: secretmanager/v0.1.0
  Pushed to origin
```

Finally, publish:

```
$ librarianx release publish
secretmanager/v0.1.0
  Tag verified: secretmanager/v0.1.0
  No action needed - pkg.go.dev will automatically index this release
  Track indexing: https://pkg.go.dev/cloud.google.com/go/secretmanager/apiv1
```

For Go, the publish step is a no-op. Go modules are published automatically
when you push git tags. The command just verifies the tags exist.

## Adding More Libraries

Let's add Access Approval to our Go repository:

```
$ librarianx add accessapproval google/cloud/accessapproval/v1
Created accessapproval/.librarian.yaml
```

Generate it:

```
$ librarianx generate accessapproval
Generated accessapproval/
```

### Updating Everything

Time passes. You want to update to the latest googleapis and regenerate all
libraries. This is common when googleapis adds new methods or fixes bugs.

Update the googleapis reference:

```
$ librarianx config update generate.googleapis
Updated .librarian/config.yaml:
  googleapis.path: https://github.com/googleapis/googleapis/archive/a1b2c3d4...tar.gz
  googleapis.sha256: 867048ec8f0850a4d77ad836319e4c0a0c624928611af8a900cd77e676164e8e
```

Regenerate all libraries:

```
$ librarianx generate --all
Generated secretmanager/
Generated accessapproval/
```

Commit the changes:

```
$ git add .
$ git commit -m "feat: update to googleapis a1b2c3d4"
```

Prepare releases for everything that changed:

```
$ librarianx release prepare
Detected 2 libraries with pending releases:
  - secretmanager: 0.1.0 → 0.2.0 (minor - new features)
  - accessapproval: null → 0.1.0 (initial release)

Created commit: chore(release): prepare releases
```

Tag and publish:

```
$ librarianx release tag
Creating tags for 2 libraries...

$ librarianx release publish
Publishing 2 libraries...
```

Done! Both libraries are updated.

## Python Libraries

Let's try Python. Python requires installing dependencies before generation.

```
$ cd ../
$ mkdir google-cloud-python
$ cd google-cloud-python
```

Initialize a Python repository:

```
$ librarianx init python
Created .librarian/config.yaml
```

Python projects typically use a `packages/` directory for generated libraries.
Let's configure that:

```
$ librarianx config set generate.dir packages/
Updated .librarian/config.yaml
```

Install Python generator dependencies:

```
$ librarianx install python --use-container
Using Docker container for code generation
Container image: us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/python-librarian-generator:latest
```

Add Secret Manager:

```
$ librarianx add google-cloud-secret-manager google/cloud/secretmanager/v1 google/cloud/secretmanager/v1beta2
Created packages/google-cloud-secret-manager/.librarian.yaml
```

Notice Python uses package names with prefixes (google-cloud-secret-manager).
This matches PyPI naming conventions.

Generate the library:

```
$ librarianx generate google-cloud-secret-manager
Generated packages/google-cloud-secret-manager/
```

Check the output:

```
$ ls packages/google-cloud-secret-manager/
google/
  cloud/
    secretmanager_v1/
      __init__.py
      services/
      types/
    secretmanager_v1beta2/
tests/
setup.py
README.rst
```

Release workflow is similar to Go, but publishes to PyPI:

```
$ git add .
$ git commit -m "feat(secretmanager): add Secret Manager Python client"
$ librarianx release prepare
$ librarianx release tag
$ librarianx release publish
```

For Python, the publish step uploads to PyPI:

```
$ librarianx release publish
google-cloud-secret-manager v0.1.0
  Checked out tag: google-cloud-secret-manager/v0.1.0
  Building distribution...
  Built: dist/google-cloud-secret-manager-0.1.0.tar.gz
  Built: dist/google_cloud_secret_manager-0.1.0-py3-none-any.whl
  Uploading to PyPI...
  Published: https://pypi.org/project/google-cloud-secret-manager/0.1.0/
```

## Rust Libraries

Rust works similarly:

```
$ cd ../
$ mkdir google-cloud-rust
$ cd google-cloud-rust
```

Initialize a Rust repository:

```
$ librarianx init rust
Created .librarian/config.yaml
```

Rust typically uses a `generated/` directory:

```
$ librarianx config set generate.dir generated/
Updated .librarian/config.yaml
```

Install Rust generator dependencies:

```
$ librarianx install rust --use-container
Using Docker container for code generation
Container image: us-central1-docker.pkg.dev/cloud-sdk-librarian-prod/images-prod/rust-librarian-generator:latest
```

Add libraries:

```
$ librarianx add secretmanager google/cloud/secretmanager/v1
Created generated/google-cloud-secretmanager-v1/.librarian.yaml

$ librarianx add accessapproval google/cloud/accessapproval/v1
Created generated/google-cloud-accessapproval-v1/.librarian.yaml
```

Generate both:

```
$ librarianx generate --all
Generated generated/google-cloud-secretmanager-v1/
Generated generated/google-cloud-accessapproval-v1/
```

Check the output:

```
$ ls generated/google-cloud-secretmanager-v1/
src/
  lib.rs
  client.rs
  types.rs
Cargo.toml
README.md
```

Release workflow is similar, but publishes to crates.io:

```
$ git add .
$ git commit -m "feat: add Rust client libraries"
$ librarianx release prepare
$ librarianx release tag
$ librarianx release publish
```

Rust uses `cargo publish` to upload to crates.io:

```
$ librarianx release publish
google-cloud-secretmanager-v1 v0.1.0
  Checked out tag: google-cloud-secretmanager-v1/v0.1.0
  Running cargo semver-checks...
  Validation passed
  Publishing to crates.io...
  Published: https://crates.io/crates/google-cloud-secretmanager-v1/0.1.0
```

## Working with Handwritten Code

Not all code needs to be generated. You can use librarianx just for release
management.

Go back to the Go repository and create a handwritten library:

```
$ cd ../google-cloud-go
$ mkdir custom-tool
$ echo "package customtool\n\nfunc Hello() { println(\"hello\") }" > custom-tool/tool.go
```

Add it to librarian:

```
$ librarianx add custom-tool
Created custom-tool/.librarian.yaml
```

This created a config with only a `release` section (no `generate` section):

```
$ cat custom-tool/.librarian.yaml
release:
  version: null
```

Now you can release it like any other library:

```
$ git add .
$ git commit -m "feat(custom-tool): add custom tool"
$ librarianx release prepare
$ librarianx release tag
$ librarianx release publish
```

## Summary

Librarianx provides a consistent workflow across languages:

1. **Initialize** - `librarianx init <language>`
2. **Install** - `librarianx install <language> --use-container`
3. **Add APIs** - `librarianx add <name> <api-paths>`
4. **Generate** - `librarianx generate <name>`
5. **Release** - `librarianx release prepare/tag/publish`

The same commands work for Go, Python, and Rust. Configuration lives in
`.librarian.yaml` files, making everything transparent and version-controlled.

Key differences by language:
- **Go**: Modules auto-publish to pkg.go.dev when tags are pushed
- **Python**: Uses `packages/` directory, publishes to PyPI
- **Rust**: Uses `generated/` directory, publishes to crates.io with semver checks

Try it out! Feedback and bug reports are welcome at
https://github.com/julieqiu/librarianx/issues.
