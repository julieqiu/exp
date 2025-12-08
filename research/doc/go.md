Go GAPIC Generator Guide
========================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/storage/v2/storage.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `storage_v2.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `storage_grpc_service_config.json`\)

Configuration Options
---------------------

| Option                | Description                              | Example                                     |
|-----------------------|------------------------------------------|---------------------------------------------|
| `go-gapic-package`    | Import path and package name (required)  | `cloud.google.com/go/storage/apiv2;storage` |
| `module`              | Prefix to strip from output paths        | `cloud.google.com/go`                       |
| `transport`           | Protocol: `grpc`, `rest`, or `grpc+rest` | `grpc+rest`                                 |
| `grpc-service-config` | Path to gRPC config JSON                 | `path/to/grpc_service_config.json`          |
| `api-service-config`  | Path to service YAML                     | `path/to/service.yaml`                      |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
go install github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_gapic@latest
```

Examples
--------

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis
```

### Secret Manager

```bash
mkdir -p go/secretmanager && protoc \
  --proto_path=$GOOGLEAPIS \
  --go_gapic_out=go/secretmanager/ \
  '--go_gapic_opt=go-gapic-package=cloud.google.com/go/secretmanager/apiv1;secretmanager,module=cloud.google.com/go,transport=grpc+rest' \
  --go_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_grpc_service_config.json \
  --go_gapic_opt=api-service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_v1.yaml \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/service.proto \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/resources.proto
```

### Cloud Storage

```bash
mkdir -p go/storage && protoc \
  --proto_path=$GOOGLEAPIS \
  --go_gapic_out=go/storage/ \
  '--go_gapic_opt=go-gapic-package=cloud.google.com/go/storage/apiv2;storage,module=cloud.google.com/go,transport=grpc' \
  --go_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/storage/v2/storage_grpc_service_config.json \
  --go_gapic_opt=api-service-config=$GOOGLEAPIS/google/storage/v2/storage_v2.yaml \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

Go does **not** use OwlBot. The google-cloud-go repository migrated to Librarian with `librariangen` for code generation automation. Generated code is managed directly through the Librarian workflow in `github.com/googleapis/librarian`.
