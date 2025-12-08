C# GAPIC Generator Guide
========================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/storage/v2/storage.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `storage_v2.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `storage_grpc_service_config.json`\)

Configuration Options
---------------------

| Option                 | Description         | Example                            |
|------------------------|---------------------|------------------------------------|
| `csharp-gapic-package` | C# namespace        | `Google.Cloud.SecretManager.V1`    |
| `transport`            | Protocol            | `grpc+rest`                        |
| `grpc-service-config`  | Path to gRPC config | `path/to/grpc_service_config.json` |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
git clone https://github.com/googleapis/gapic-generator-csharp.git
cd gapic-generator-csharp && dotnet build
```

Examples
--------

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis
```

### Secret Manager

```bash
mkdir -p csharp/secretmanager && protoc \
  --proto_path=$GOOGLEAPIS \
  --csharp_gapic_out=csharp/secretmanager/ \
  --csharp_gapic_opt=transport=grpc+rest \
  --csharp_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_grpc_service_config.json \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/service.proto \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/resources.proto
```

### Cloud Storage

```bash
mkdir -p csharp/storage && protoc \
  --proto_path=$GOOGLEAPIS \
  --csharp_gapic_out=csharp/storage/ \
  --csharp_gapic_opt=transport=grpc \
  --csharp_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/storage/v2/storage_grpc_service_config.json \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

C# does **not** use OwlBot. The google-cloud-dotnet repository uses a custom `ReleaseManager` tool located in `/tools/Google.Cloud.Tools.ReleaseManager` for code generation and release automation.
