# Surfer

Surfer is a code generator that creates gcloud CLI command definitions from Google Cloud API specifications.

## What Does It Do?

Surfer generates YAML files that the gcloud CLI uses to create commands for Google Cloud services. Instead of manually writing command definitions, Surfer automatically generates them from:
- API protocol buffer definitions (`.proto` files)
- A service configuration file (`gcloud.yaml`)

The generated YAML files define everything gcloud needs to know about a command: its arguments, help text, how it maps to API calls, and how to handle responses.

## Quick Start

```bash
# Generate gcloud command definitions for the parallelstore service
surfer generate parallelstore

# This reads: testdata/parallelstore/gcloud.yaml
# And writes to: testdata/parallelstore/generated/
```

## How It Works

1. **Read configuration**: Reads `testdata/{service}/gcloud.yaml` to understand which API to generate commands for and how to customize them
2. **Load API definitions**: Loads protocol buffer (`.proto`) files from the googleapis repository that define the service's API
3. **Build model**: Creates an internal representation of the API (services, methods, messages, fields)
4. **Generate YAML**: Transforms the API model into gcloud command YAML files according to the configuration

## Usage

```bash
surfer generate <service> [flags]
```

Generate gcloud command YAML files for a service.

**Arguments:**

- `<service>` - Name of the service to generate (e.g., `parallelstore`, `memorystore`)
  - Configuration file must exist at `testdata/{service}/gcloud.yaml`
  - See [examples](#examples) for available services

**Flags:**

- `--googleapis <path>` - Location of the googleapis repository containing `.proto` files
  - Can be a local directory path or a Git URL
  - Default: `/Users/julieqiu/code/googleapis/googleapis`

- `--output <path>` - Where to write the generated YAML files
  - Default: `testdata/{service}/generated/`

### Examples

```bash
# Basic usage - generate commands for parallelstore
surfer generate parallelstore
# Reads from:  testdata/parallelstore/gcloud.yaml
# Writes to:   testdata/parallelstore/generated/

# Generate for a different service
surfer generate memorystore

# Use a local clone of googleapis
surfer generate parallelstore --googleapis=/path/to/googleapis

# Use googleapis from GitHub directly
surfer generate parallelstore --googleapis=https://github.com/googleapis/googleapis

# Write output to a custom location
surfer generate parallelstore --output=/custom/output/directory
```

Available services in testdata: `parallelstore`, `memorystore`, `parametermanager`

## Configuration

Each service needs a `gcloud.yaml` configuration file at `testdata/{service}/gcloud.yaml`. This file tells Surfer:
- Which API to generate commands for
- What version of the API to use
- How to customize the generated commands (help text, formatting, etc.)

The configuration schema is defined in `internal/gcloudconfig`.

### Basic Structure

```yaml
service_name: <fully-qualified-service-name>  # e.g., parallelstore.googleapis.com
apis:
  - name: <API-name>           # e.g., Parallelstore
    api_version: <version>      # e.g., v1
    release_tracks:             # Which gcloud tracks to generate for
      - GA                      # Can be: ALPHA, BETA, GA
    root_is_hidden: true        # Whether to hide the root command
```

### Customization Options

You can customize the generated commands by adding these fields to the API configuration:

- **`help_text`** - Override help text for methods and fields
- **`output_formatting`** - Customize how command output is displayed
- **`command_operations_config`** - Configure long-running operation behavior
- **`method_generation_filters`** - Control which API methods get commands generated

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

## Project Structure

```
testdata/
├── parallelstore/
│   ├── gcloud.yaml       # Input: Configuration for parallelstore service
│   └── generated/        # Output: Generated command YAML files
│       ├── create.yaml
│       ├── delete.yaml
│       ├── list.yaml
│       └── _partials/    # Track-specific command definitions
│           ├── _create_ga.yaml
│           ├── _delete_ga.yaml
│           └── _list_ga.yaml
├── memorystore/
│   ├── gcloud.yaml
│   └── generated/
└── parametermanager/
    ├── gcloud.yaml
    └── generated/

internal/
├── gcloudconfig/         # Input configuration schema (gcloud.yaml)
├── commandyaml/          # Output command YAML schema
└── surfer/               # CLI and generation logic
```

### Generated Files

Each generated YAML file defines a gcloud command:
- **Top-level files** (e.g., `create.yaml`) - Contain `_PARTIALS_: true` to reference partial files
- **Partial files** (e.g., `_create_ga.yaml`) - Contain the actual command definition for a specific release track

The partial file structure allows different command definitions for ALPHA, BETA, and GA tracks.

## How Surfer Works Internally

### Key Packages

- **`internal/gcloudconfig`** - Defines the schema for `gcloud.yaml` input files
- **`internal/commandyaml`** - Defines the schema for generated command YAML output files
- **`internal/sidekick`** - Parses protocol buffer definitions and builds API models
- **`internal/surfer`** - CLI implementation and orchestrates the generation process

### Generation Flow

1. **Parse configuration** - Read and validate `testdata/{service}/gcloud.yaml`
2. **Load API model** - Parse `.proto` files from googleapis and build an API model (services, methods, messages, fields)
3. **Apply customizations** - Merge configuration from `gcloud.yaml` (help text, output formatting, etc.)
4. **Generate YAML** - Transform the API model into gcloud command YAML files
5. **Write output** - Save files to `testdata/{service}/generated/`

---

## Reference: Command YAML Format

This section describes the structure of the generated command YAML files. These files define how gcloud commands work: their arguments, help text, API mappings, and behavior.

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
