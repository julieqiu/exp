Java GAPIC Generator Guide
==========================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/storage/v2/storage.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `storage_v2.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `storage_grpc_service_config.json`\)
-	**GAPIC Config (YAML):** Client library customizations (e.g., `storage_gapic.yaml`\)

Configuration Options
---------------------

| Option                | Description            | Example                             |
|-----------------------|------------------------|-------------------------------------|
| `java-gapic-package`  | Java package name      | `com.google.cloud.secretmanager.v1` |
| `transport`           | Protocol               | `grpc+rest`                         |
| `grpc-service-config` | Path to gRPC config    | `path/to/grpc_service_config.json`  |
| `gapic-config`        | Path to GAPIC config   | `path/to/api_gapic.yaml`            |
| `service-yaml`        | Path to service YAML   | `path/to/service.yaml`              |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
git clone https://github.com/googleapis/sdk-platform-java.git
cd sdk-platform-java && mvn install -DskipTests
```

Examples
--------

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis
```

### Secret Manager

```bash
mkdir -p java/secretmanager && protoc \
  --proto_path=$GOOGLEAPIS \
  --java_gapic_out=java/secretmanager/ \
  --java_gapic_opt=java-gapic-package=com.google.cloud.secretmanager.v1 \
  --java_gapic_opt=transport=grpc+rest \
  --java_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_grpc_service_config.json \
  --java_gapic_opt=service-yaml=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_v1.yaml \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/service.proto \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/resources.proto
```

### Cloud Storage

```bash
mkdir -p java/storage && protoc \
  --proto_path=$GOOGLEAPIS \
  --java_gapic_out=java/storage/ \
  --java_gapic_opt=java-gapic-package=com.google.cloud.storage.v2 \
  --java_gapic_opt=transport=grpc \
  --java_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/storage/v2/storage_grpc_service_config.json \
  --java_gapic_opt=service-yaml=$GOOGLEAPIS/google/storage/v2/storage_v2.yaml \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

Java uses **OwlBot** for post-processing generated code.

-	**Docker Image:** `gcr.io/cloud-devrel-public-resources/owlbot-java:latest`
-	**Config File:** `.OwlBot-hermetic.yaml` in each library directory

OwlBot copies generated code from staging directories to final locations and applies formatting.
