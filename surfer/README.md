# Surfer

Surfer generates gcloud command surface definitions from Google API protocol
buffers and configuration files.

## Overview

Surfer automates the creation of gcloud CLI command surfaces by processing:
- Protocol buffer definitions from googleapis
- Service configurations from googleapis
- Custom gcloud configuration files (gcloud.yaml)

The tool produces YAML files that define gcloud commands, including their
arguments, help text, request mappings, and async operation configurations.

## How It Works

1. **Parse configuration**: Surfer reads the gcloud.yaml file to understand the surface definition
2. **Load proto definitions**: Surfer loads protocol buffer definitions from the googleapis repository
3. **Build API model**: Using `internal/api`, Surfer creates an internal representation of the API surface
4. **Generate commands**: Surfer transforms the API model into gcloud command YAML files and writes them to the output directory

## Command

```bash
surfer generate [flags]
```

Generates gcloud surface definitions from a gcloud.yaml configuration file.

**Flags:**

- `--googleapis` - Path to googleapis repository (local directory or URL). Default: `/Users/julieqiu/code/googleapis/googleapis`
- `--gcloud-yaml` - Path to gcloud.yaml configuration file. Default: `testdata/parallelstore/gcloud.yaml`
- `--output` - Output directory for generated surfaces. Default: current working directory

**Examples:**

```bash
# Generate with defaults (testdata)
surfer generate

# Generate with custom googleapis
surfer generate \
  --googleapis=/path/to/googleapis \
  --gcloud-yaml=parallelstore/gcloud.yaml

# Generate with specific output directory
surfer generate \
  --googleapis=/path/to/googleapis \
  --gcloud-yaml=parallelstore/gcloud.yaml \
  --output=parallelstore/surface

# Generate using googleapis URL
surfer generate \
  --googleapis=https://github.com/googleapis/googleapis \
  --gcloud-yaml=myservice/gcloud.yaml \
  --output=generated
```

## Configuration

### gcloud.yaml

The gcloud.yaml file configures how the surface is generated. It follows the schema defined in `internal/gcloudconfig`.

**Top-level fields:**
- `service_name` (required) - The fully qualified service name (e.g., "parallelstore.googleapis.com")
- `apis` (required) - List of API configurations
- `resource_patterns` (optional) - Additional resource patterns not in proto descriptors

**API configuration:**
- `name` (required) - API name (e.g., "Parallelstore")
- `api_version` (required) - API version (e.g., "v1")
- `root_is_hidden` - Whether to hide the root command group
- `release_tracks` - List of release tracks (ALPHA, BETA, GA)
- `help_text` - Help text rules for services, messages, methods, and fields
- `output_formatting` - Output format specifications for commands
- `command_operations_config` - Long-running operation configurations

**Example:**
```yaml
service_name: parallelstore.googleapis.com
apis:
  - name: Parallelstore
    api_version: v1
    root_is_hidden: true
    release_tracks:
      - GA
    help_text:
      method_rules:
        - selector: google.cloud.parallelstore.v1.Parallelstore.CreateInstance
          help_text:
            brief: Creates a Parallelstore instance
            description: |
              Creates a Parallelstore instance.
            examples:
              - |-
                To create an instance run:
                $ {command} my-instance --capacity-gib=12000
    output_formatting:
      - selector: google.cloud.parallelstore.v1.Parallelstore.ListInstances
        format: |-
          table(name, capacityGib:label=Capacity, state)
```

## Output Structure

Generated command files are organized by resource and operation:

```
<output-dir>/
└── <service>/
    └── surface/
        ├── <resource>/
        │   ├── <command>.yaml
        │   └── _partials/
        │       └── _<command>_<track>.yaml
        └── ...
```

**Example output:**
```
parallelstore/
└── surface/
    ├── instances/
    │   ├── create.yaml
    │   ├── list.yaml
    │   ├── describe.yaml
    │   └── _partials/
    │       ├── _create_ga.yaml
    │       ├── _list_ga.yaml
    │       └── _describe_ga.yaml
    └── operations/
        ├── list.yaml
        └── _partials/
            └── _list_ga.yaml
```

## Testdata

The repository includes test data demonstrating the expected configuration and
output:

**Input:**
- `testdata/parallelstore/gcloud.yaml` - Sample gcloud.yaml configuration

**Output:**
- `testdata/parallelstore/generated/` - Generated command surfaces

**Note:** Proto files are loaded from the googleapis repository specified via
the `--googleapis` flag and are not included in testdata.

## Implementation

Surfer uses several internal packages:

- `internal/api` - Core API model representation and transformation logic
- `internal/gcloudconfig` - Go types for parsing gcloud.yaml configuration
- `internal/commandyaml` - Types for generated gcloud command structures

The generation process:
1. Parse the gcloud.yaml file using `gcloudconfig.Config`
2. Load proto descriptors from the googleapis repository
3. Build API model using `api.API`
4. Apply custom configurations from gcloud.yaml (help text, output formatting, etc.)
5. Generate command YAML files using the API model
6. Write generated files to the output directory

## gcloud Command YAML

Surfer generates gcloud command definitions as YAML files. These files define the
command-line interface for interacting with Google Cloud APIs through the gcloud
CLI.

### File Organization

Commands are organized into two types of files:

**Top-level command files** (e.g., `create.yaml`, `list.yaml`):
- Located directly in the resource directory
- Contain only `_PARTIALS_: true` to indicate they reference partial files

**Partial files** (e.g., `_create_ga.yaml`, `_list_alpha.yaml`):
- Located in `_partials/` subdirectory under each resource
- Contain the actual command definitions
- Named with pattern `_<command>_<track>.yaml` where track is `ga`, `beta`, or
  `alpha`

### Command Structure

Each partial file contains a YAML array with command definitions. The structure
follows the types defined in `internal/gcloud/command.go`.

**Top-level fields:**

- `release_tracks` ([]string) - Release tracks for this command (e.g., `["GA"]`,
  `["ALPHA", "BETA"]`)
- `auto_generated` (bool) - Whether this command is autogenerated
- `hidden` (bool) - Whether to hide the command from help output
- `help_text` (object) - Help text for the command
- `arguments` (object) - Command arguments and flags
- `request` (object) - API request configuration
- `response` (object, optional) - Response field configuration
- `async` (object, optional) - Asynchronous operation configuration

### help_text

Defines user-facing documentation for the command.

**Fields:**
- `brief` (string) - One-line description of the command
- `description` (string) - Detailed multi-line description
- `examples` (string) - Example usage with the `{command}` placeholder

**Example:**
```yaml
help_text:
  brief: Creates a Parallelstore instance
  description: |
    Creates a Parallelstore instance.
  examples: |-
    To create an instance `my-instance` in location `us-central1-a` run:

    $ {command} my-instance --capacity-gib=12000 --location=us-central1-a
```

### arguments

Defines command-line arguments and flags.

**Fields:**
- `params` ([]object) - List of parameters (arguments and flags)

**Parameter fields:**

- `arg_name` (string) - Command-line argument name (e.g., `capacity-gib`)
- `api_field` (string) - Corresponding API field path (e.g., `instance.capacityGib`)
- `help_text` (string) - Description of the parameter
- `is_positional` (bool) - Whether this is a positional argument
- `is_primary_resource` (bool) - Whether this is the primary resource identifier
- `required` (bool) - Whether the parameter is required
- `request_id_field` (string) - Field name for request ID (used with primary
  resources)
- `resource_spec` (!REF string) - Reference to resource specification
- `resource_method_params` (map[string]string) - Mapping of resource parameters
- `type` (string) - Parameter type (e.g., `long`, `string`)
- `repeated` (bool) - Whether the parameter can be specified multiple times
- `spec` ([]object) - Field specifications for complex types
- `choices` ([]object) - Allowed values for enum parameters

**Parameter with choices:**
```yaml
- arg_name: file-stripe-level
  api_field: instance.fileStripeLevel
  help_text: Stripe level for files
  choices:
    - arg_value: file-stripe-level-min
      enum_value: FILE_STRIPE_LEVEL_MIN
      help_text: Minimum file striping
    - arg_value: file-stripe-level-balanced
      enum_value: FILE_STRIPE_LEVEL_BALANCED
      help_text: Medium file striping
```

**Resource parameter:**
```yaml
- is_positional: true
  is_primary_resource: true
  request_id_field: instanceId
  resource_spec: !REF googlecloudsdk.command_lib.parallelstore.v1_resources:project_location_instance
  required: true
  help_text: Resource identifier
```

**Complex parameter with spec:**
```yaml
- arg_name: labels
  api_field: instance.labels
  repeated: true
  help_text: Cloud Labels for organizing resources
  spec:
    - api_field: key
    - api_field: value
```

### request

Defines how the command maps to an API request.

**Fields:**
- `api_version` (string) - API version to use (e.g., `v1`)
- `collection` ([]string) - API collection path (e.g.,
  `["parallelstore.projects.locations.instances"]`)
- `method` (string, optional) - HTTP method or operation name
- `ALPHA`, `BETA`, `GA` (object, optional) - Track-specific request configuration
  with `api_version` field

**Example:**
```yaml
request:
  api_version: v1
  collection:
    - parallelstore.projects.locations.instances
```

**Track-specific versions:**
```yaml
request:
  ALPHA:
    api_version: v1alpha1
  BETA:
    api_version: v1beta1
  GA:
    api_version: v1
  collection:
    - parallelstore.projects.locations.instances
```

### response

Configures response field handling.

**Fields:**
- `id_field` (string) - Field name to use as the resource ID

**Example:**
```yaml
response:
  id_field: name
```

### async

Configures long-running operation support.

**Fields:**
- `collection` ([]string) - Operations collection path for polling

**Example:**
```yaml
async:
  collection:
    - parallelstore.projects.locations.operations
```

### Complete Example

```yaml
- release_tracks:
    - GA
  auto_generated: true
  hidden: true
  help_text:
    brief: Creates a Parallelstore instance
    description: |
      Creates a Parallelstore instance.
    examples: |-
      To create an instance run:

      $ {command} my-instance --capacity-gib=12000 --location=us-central1-a
  arguments:
    params:
      - is_positional: true
        is_primary_resource: true
        request_id_field: instanceId
        resource_spec: !REF googlecloudsdk.command_lib.parallelstore.v1_resources:project_location_instance
        required: true
        help_text: Resource name of the instance
      - arg_name: capacity-gib
        api_field: instance.capacityGib
        type: long
        required: true
        help_text: Storage capacity in GiB
      - arg_name: description
        api_field: instance.description
        help_text: Description of the instance
  request:
    api_version: v1
    collection:
      - parallelstore.projects.locations.instances
  async:
    collection:
      - parallelstore.projects.locations.operations
```
