Ruby GAPIC Generator Guide
==========================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/storage/v2/storage.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `storage_v2.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `storage_grpc_service_config.json`\)

Configuration Options
---------------------

Ruby uses YAML configuration files with symbol keys.

| Option                | Description              | Example                       |
|-----------------------|--------------------------|-------------------------------|
| `:gem:name`           | Ruby gem name            | `google-cloud-secret_manager` |
| `:gem:version`        | Gem version              | `0.1.0`                       |
| `:transports`         | List of transports       | `["grpc", "rest"]`            |
| `grpc-service-config` | Path to gRPC config JSON | `path/to/grpc_config.json`    |

### Configuration File (generator_config.yml)

```yaml
:gem:
  :name: google-cloud-secret_manager
  :version: "0.1.0"
:transports:
  - grpc
  - rest
```

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
gem install gapic-generator
```

Examples
--------

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis
```

### Secret Manager

```bash
mkdir -p ruby/secretmanager && protoc \
  --proto_path=$GOOGLEAPIS \
  --ruby_gapic_out=ruby/secretmanager/ \
  --ruby_gapic_opt=configuration=generator_config.yml \
  --ruby_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_grpc_service_config.json \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/service.proto \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/resources.proto
```

### Cloud Storage

```bash
mkdir -p ruby/storage && protoc \
  --proto_path=$GOOGLEAPIS \
  --ruby_gapic_out=ruby/storage/ \
  --ruby_gapic_opt=configuration=generator_config.yml \
  --ruby_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/storage/v2/storage_grpc_service_config.json \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

Ruby uses **OwlBot** for post-processing generated code.

-	**Docker Image:** `gcr.io/cloud-devrel-public-resources/owlbot-ruby:latest`
-	**Config File:** Per-library `.OwlBot.yaml` files

OwlBot copies generated code from staging directories and applies Ruby-specific formatting.
