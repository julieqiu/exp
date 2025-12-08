Node.js GAPIC Generator Guide
=============================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/cloud/secretmanager/v1/*.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `secretmanager_v1.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `secretmanager_grpc_service_config.json`\)

Outputs
-------

The generator outputs an unformatted npm package:

```
packages/google-cloud-secretmanager/
├── src/
│   ├── index.ts
│   └── v1/
│       ├── secret_manager_service_client.ts
│       ├── secret_manager_service_client_config.json
│       ├── secret_manager_service_proto_list.json
│       ├── gapic_metadata.json
│       └── index.ts
├── test/
│   └── gapic_secret_manager_service_v1.ts
├── protos/
│   └── google/cloud/secretmanager/v1/*.proto
├── package.json
├── tsconfig.json
├── README.md
├── .eslintrc.json
├── .prettierrc.js
└── ... other config files
```

Configuration Options
---------------------

| Option                | Description              | Example                            |
|-----------------------|--------------------------|------------------------------------|
| `package-name`        | NPM package name         | `@google-cloud/secret-manager`     |
| `main-service`        | Main service to generate | `SecretManagerService`             |
| `grpc-service-config` | Path to gRPC config JSON | `path/to/grpc_service_config.json` |
| `service-yaml`        | Path to service YAML     | `path/to/service.yaml`             |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
npm install -g @google-cloud/gapic-generator
```

Examples
--------

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis
```

### Secret Manager

```bash
protoc \
  --proto_path=$GOOGLEAPIS \
  --typescript_gapic_out=packages/google-cloud-secretmanager/ \
  --typescript_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_grpc_service_config.json \
  --typescript_gapic_opt=service-yaml=$GOOGLEAPIS/google/cloud/secretmanager/v1/secretmanager_v1.yaml \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/service.proto \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/resources.proto
```

### Cloud Storage

```bash
protoc \
  --proto_path=$GOOGLEAPIS \
  --typescript_gapic_out=packages/google-cloud-storage/ \
  --typescript_gapic_opt=grpc-service-config=$GOOGLEAPIS/google/storage/v2/storage_grpc_service_config.json \
  --typescript_gapic_opt=service-yaml=$GOOGLEAPIS/google/storage/v2/storage_v2.yaml \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

After generation, run these commands in the output directory:

### 1. Install dependencies

```bash
cd packages/google-cloud-secretmanager/
npm install
```

### 2. Format code

```bash
npm run fix
```

This runs prettier and eslint via gts to format all TypeScript files.

### 3. Apply patches (library-specific)

Some libraries need additional helper methods injected into the generated client:

```bash
node librarian.js
```

This adds methods like `secretPath()` and `matchProjectFromSecretName()` that the generator does not produce.
