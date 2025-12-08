Python GAPIC Generator Guide
============================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/cloud/secretmanager/v1/*.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `secretmanager_v1.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `secretmanager_grpc_service_config.json`\)

Outputs
-------

The generator outputs a partially formatted Python package (basic whitespace fixes applied):

```
packages/google-cloud-secret-manager/
├── google/
│   └── cloud/
│       ├── secretmanager/
│       │   └── __init__.py
│       └── secretmanager_v1/
│           ├── __init__.py
│           ├── gapic_metadata.json
│           ├── gapic_version.py
│           ├── py.typed
│           ├── services/
│           │   └── secret_manager_service/
│           │       ├── __init__.py
│           │       ├── client.py         # sync client
│           │       ├── async_client.py   # async client
│           │       ├── pagers.py
│           │       └── transports/
│           │           ├── __init__.py
│           │           ├── base.py
│           │           ├── grpc.py
│           │           └── rest.py
│           └── types/
│               ├── __init__.py
│               └── *.py                  # message types
├── tests/
├── docs/
├── samples/generated_samples/
├── setup.py
├── noxfile.py
├── README.rst
├── .flake8
├── mypy.ini
└── ... other config files
```

Configuration Options
---------------------

| Option                   | Description                     | Example         |
|--------------------------|---------------------------------|-----------------|
| `python-gapic-namespace` | Python namespace for the client | `google.cloud`  |
| `python-gapic-name`      | Package name (used in imports)  | `secretmanager` |
| `transport`              | Protocol                        | `grpc+rest`     |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
pip install gapic-generator
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
  --python_gapic_out=packages/google-cloud-secret-manager/ \
  --python_gapic_opt=python-gapic-namespace=google.cloud \
  --python_gapic_opt=python-gapic-name=secretmanager \
  --python_gapic_opt=transport=grpc+rest \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/service.proto \
  $GOOGLEAPIS/google/cloud/secretmanager/v1/resources.proto
```

### Cloud Storage

```bash
protoc \
  --proto_path=$GOOGLEAPIS \
  --python_gapic_out=packages/google-cloud-storage/ \
  --python_gapic_opt=python-gapic-namespace=google.cloud \
  --python_gapic_opt=python-gapic-name=storage \
  --python_gapic_opt=transport=grpc \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

After generation, run these commands in the output directory:

### 1. Format code

```bash
cd packages/google-cloud-secret-manager/
nox -s format
```

This runs Black (code formatter) and isort (import sorter).

### 2. Lint check (optional)

```bash
nox -s lint
```

This runs Black in check mode and Flake8 to verify code style.

### Notes

-	Python does **not** use OwlBot. The google-cloud-python repository uses a librarian-based system.
-	Unlike Node.js, Python does not need library-specific patches - the generator produces complete code.
-	Some packages in google-cloud-python have `scripts/` folders with `fixup_*.py` files. These are legacy artifacts from a 2020 migration and are not generated. They are being removed: https://github.com/googleapis/librarian/issues/2943
