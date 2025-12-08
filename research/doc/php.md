PHP GAPIC Generator Guide
=========================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/storage/v2/storage.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `storage_v2.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `storage_grpc_service_config.json`\)
-	**GAPIC Config (YAML):** Client library customizations (e.g., `storage_gapic.yaml`\)

Configuration Options
---------------------

| Option                | Description          | Example                            |
|-----------------------|----------------------|------------------------------------|
| `php-gapic-package`   | PHP namespace        | `Google\Cloud\SecretManager\V1`    |
| `transport`           | Protocol             | `grpc+rest`                        |
| `grpc-service-config` | Path to gRPC config  | `path/to/grpc_service_config.json` |
| `gapic-config`        | Path to GAPIC config | `path/to/api_gapic.yaml`           |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
git clone https://github.com/googleapis/gapic-generator-php.git
cd gapic-generator-php && composer install
```

Examples
--------

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis
```

### Secret Manager

```bash
mkdir -p php/secretmanager && protoc \
  --proto_path=$GOOGLEAPIS \
  --php_gapic_out=php/secretmanager/ \
  --php_gapic_opt=transport=grpc+rest \
  --php_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_grpc_service_config.json \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/service.proto \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/resources.proto
```

### Cloud Storage

```bash
mkdir -p php/storage && protoc \
  --proto_path=$GOOGLEAPIS \
  --php_gapic_out=php/storage/ \
  --php_gapic_opt=transport=grpc \
  --php_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/storage/v2/storage_grpc_service_config.json \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

PHP uses **OwlBot** for post-processing generated code.

-	**Docker Image:** `gcr.io/cloud-devrel-public-resources/owlbot-php:latest`
-	**Config File:** `.github/.OwlBot.yaml`

OwlBot copies generated code from staging directories and applies PHP-specific formatting.
