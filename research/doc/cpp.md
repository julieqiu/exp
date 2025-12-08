C++ GAPIC Generator Guide
=========================

Inputs
------

-	**Proto Files:** Protocol Buffer definitions (e.g., `google/storage/v2/storage.proto`\)
-	**Service Config (YAML):** Authentication and documentation (e.g., `storage_v2.yaml`\)
-	**gRPC Service Config (JSON):** Retry policies and timeouts (e.g., `storage_grpc_service_config.json`\)

Configuration Options
---------------------

| Option      | Description                  | Example                      |
|-------------|------------------------------|------------------------------|
| `namespace` | C++ namespace                | `google::cloud::storage::v2` |
| `product`   | Product name for the library | `storage`                    |

Prerequisites
-------------

```bash
git clone https://github.com/googleapis/googleapis.git $HOME/code/googleapis/googleapis
git clone https://github.com/googleapis/google-cloud-cpp.git
```

Generation
----------

C++ generation is done through CMake or Bazel within the google-cloud-cpp repository:

```bash
cd google-cloud-cpp

# Using CMake
cmake -S . -B build
cmake --build build --target google-cloud-cpp-generate

# Using Bazel
bazel build //generator:google-cloud-cpp-generator
```

### Manual protoc Invocation

```bash
export GOOGLEAPIS=$HOME/code/googleapis/googleapis

mkdir -p cpp/storage && protoc \
  --proto_path=$GOOGLEAPIS \
  --cpp_out=cpp/storage/ \
  --grpc_out=cpp/storage/ \
  --plugin=protoc-gen-grpc=$(which grpc_cpp_plugin) \
  $GOOGLEAPIS/google/storage/v2/storage.proto
```

Post-Processing
---------------

C++ does **not** use OwlBot. The google-cloud-cpp repository has an internal `/generator` directory with custom scripts for code generation automation.
