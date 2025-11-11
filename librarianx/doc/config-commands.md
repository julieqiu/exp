# Librarian Config Commands Design

## Objective

Define a consistent CLI interface for managing librarian.yaml configuration using `librarian config` subcommands.

## Background

Currently, users must manually edit librarian.yaml to update configuration. This requires understanding YAML syntax, the config schema, and proper nesting. Common operations like adding an API to an edition or updating metadata are error-prone and time-consuming.

This document proposes a `librarian config` command with subcommands to manage configuration programmatically, inspired by successful patterns from go, npm, git, and kubectl.

## Research: Config Command Patterns

### Go (`go env`)

**Commands:**
- `go env` - List all environment variables
- `go env GOPATH` - Get specific variable
- `go env -w GOPATH=/path` - Set (write) variable
- `go env -u GOPATH` - Unset (delete) variable

**Characteristics:**
- Simple flag-based interface (`-w` for write, `-u` for unset)
- No separate subcommands
- Stores in `os.UserConfigDir()/go/env` file
- Persistent across invocations

### npm (`npm config`)

**Commands:**
- `npm config list` - Show all config settings
- `npm config get <key>` - Get value(s)
- `npm config set <key>=<value>` - Set value
- `npm config delete <key>` - Delete key
- `npm config edit` - Open in editor

**Characteristics:**
- Explicit subcommands (list, get, set, delete, edit)
- Supports multiple keys in one invocation
- Clear separation between operations
- Most widely used pattern

### Git (`git config`)

**Commands:**
- `git config --list` - List all variables
- `git config <key>` - Get value
- `git config <key> <value>` - Set value
- `git config --unset <key>` - Remove key
- `git config --edit` - Open in editor

**Characteristics:**
- Hybrid: flags + positional arguments
- Single command with different behaviors based on arguments
- Supports `--global`, `--local`, `--system` scopes
- Advanced: `--rename-section`, `--remove-section`

### kubectl (`kubectl config`)

**Commands:**
- `kubectl config view` - Display merged config
- `kubectl config get-contexts` - List contexts
- `kubectl config set-context <name> --cluster=...` - Set context
- `kubectl config use-context <name>` - Switch context
- `kubectl config delete-context <name>` - Delete context
- `kubectl config current-context` - Show current context

**Characteristics:**
- Explicit subcommands
- Domain-specific operations (contexts, clusters, users)
- Composite operations (set multiple fields at once)
- Hierarchical config structure

### Cargo (`cargo config`) [Unstable]

**Commands (proposed):**
- `cargo config get <key>` - Get value
- `cargo config set <key> <value>` - Set value  [planned]
- `cargo config delete <key>` - Delete value  [planned]

**Characteristics:**
- Following npm/git patterns
- Comment-preserving TOML editor needed
- Still in development

## Design Decision: go mod edit Style

**Recommended approach:** Single command with editing flags (like `go mod edit`)

**Rationale:**
1. **Composable** - Multiple edits in single invocation
2. **Scriptable** - Perfect for tools and automation
3. **Familiar to Go developers** - Matches `go mod edit` pattern
4. **Atomic operations** - All changes applied together or not at all
5. **Clean interface** - No opening editors, pure command-line editing
6. **Order-preserving** - Flags processed in order, enabling complex transformations

## Proposed Command Structure

### Primary Command: `librarian config`

Following the `go mod edit` pattern, `librarian config` provides a command-line interface for editing `librarian.yaml`, primarily for use by tools or scripts.

```bash
# Usage
librarian config [editing flags] [-fmt|-print|-json] [librarian.yaml]
```

**By default:**
- Reads and writes `librarian.yaml` in the current directory
- Can specify a different target file after editing flags

### Output Flags

#### `-print`
Print the final config in YAML format instead of writing back to file.

```bash
librarian config -set language=python -print
```

#### `-json`
Print the final config in JSON format instead of writing back to file.

```bash
librarian config -set language=python -json
```

#### `-fmt`
Reformat the config file without making other changes. Implied by any other modifications.

```bash
# Only needed when no other flags specified
librarian config -fmt
```

### Repository-Level Editing Flags

#### `-set <key>=<value>`
Set a configuration value. Can be repeated for multiple changes.

```bash
# Set single value
librarian config -set language=python

# Set multiple values (processed in order)
librarian config -set language=go -set generate.container.tag=v1.0.0

# Set nested values using dot notation
librarian config -set sources.googleapis.url=https://github.com/googleapis/googleapis/archive/xyz789.tar.gz
```

#### `-unset <key>`
Remove a configuration key. Can be repeated.

```bash
# Remove single key
librarian config -unset sources.discovery

# Remove multiple keys
librarian config -unset sources.discovery -unset generate.defaults.rest_numeric_enums
```

### Edition Editing Flags

All edition flags require `-edition <name>` to specify which edition to modify.

#### `-edition <name>`
Specifies which edition to operate on. Must come before edition-specific flags.

```bash
librarian config -edition secretmanager -set-version 0.2.0
```

#### `-add-edition <name>`
Add a new edition to the config.

```bash
# Add minimal edition
librarian config -add-edition secretmanager

# Add edition and set initial values
librarian config -add-edition secretmanager -edition secretmanager -add-api google/cloud/secretmanager/v1
```

#### `-drop-edition <name>`
Remove an edition from config.

```bash
librarian config -drop-edition secretmanager
```

#### `-set-version <version>`
Set edition version (requires `-edition`).

```bash
librarian config -edition secretmanager -set-version 1.0.0
```

#### `-add-api <path>`
Add an API path to edition's apis array (requires `-edition`). Can be repeated.

```bash
# Add single API
librarian config -edition secretmanager -add-api google/cloud/secretmanager/v1

# Add multiple APIs
librarian config -edition secretmanager \
  -add-api google/cloud/secretmanager/v1 \
  -add-api google/cloud/secretmanager/v1beta2
```

#### `-drop-api <path>`
Remove an API path from edition's apis array (requires `-edition`).

```bash
librarian config -edition secretmanager -drop-api google/cloud/secretmanager/v1beta2
```

#### `-set-metadata <key>=<value>`
Set edition metadata field (requires `-edition`).

```bash
# Set single metadata field
librarian config -edition secretmanager -set-metadata name_pretty="Secret Manager"

# Set multiple fields
librarian config -edition secretmanager \
  -set-metadata name_pretty="Secret Manager" \
  -set-metadata release_level=stable \
  -set-metadata product_documentation=https://cloud.google.com/secret-manager/docs
```

#### `-unset-metadata <key>`
Remove edition metadata field (requires `-edition`).

```bash
librarian config -edition secretmanager -unset-metadata api_description
```

#### `-add-keep <pattern>`
Add pattern to edition's keep array (requires `-edition`).

```bash
librarian config -edition secretmanager -add-keep README.md -add-keep docs/
```

#### `-drop-keep <pattern>`
Remove pattern from edition's keep array (requires `-edition`).

```bash
librarian config -edition secretmanager -drop-keep docs/
```

#### `-add-remove <pattern>`
Add pattern to edition's remove array (requires `-edition`).

```bash
librarian config -edition secretmanager -add-remove temp.txt
```

#### `-drop-remove <pattern>`
Remove pattern from edition's remove array (requires `-edition`).

```bash
librarian config -edition secretmanager -drop-remove temp.txt
```

### Query Commands (No Editing)

These are separate subcommands for querying config, not editing flags.

#### `librarian config list`

Display all configuration settings.

```bash
# List entire config
librarian config list

# List in JSON format
librarian config list -json
```

#### `librarian config get <key>`

Get value for a configuration key.

```bash
# Get repository-level value
librarian config get language
# Output: go

# Get nested value
librarian config get sources.googleapis.url

# Get edition-level value
librarian config get -edition secretmanager version

# Output as JSON
librarian config get -json sources.googleapis
```

#### `librarian config editions`

List all editions.

```bash
# List edition names
librarian config editions

# Output:
# secretmanager
# pubsub
```

## Key Notation

Use **dot notation** for nested keys:

```bash
# Repository level
sources.googleapis.url
generate.container.image
generate.container.tag

# Edition level (requires --edition flag)
--edition secretmanager generate.metadata.name_pretty
--edition secretmanager generate.apis[0].path
```

**Array access:**
- Use `--add` flag to append to arrays
- Use `--from-array` with `delete` to remove from arrays
- Use `[index]` notation for direct access (advanced)

## Scoping

Configuration has two scopes:

1. **Repository-level** (default)
   - `version`, `language`, `sources`, `generate`, `release`

2. **Edition-level** (requires `--edition <name>`)
   - Anything under `editions[].`
   - Examples: `version`, `apis`, `generate.metadata`, etc.

**Examples:**
```bash
# Repository-level: no flag needed
librarian config get language
librarian config set generate.container.tag latest

# Edition-level: requires --edition flag
librarian config get --edition secretmanager version
librarian config set --edition secretmanager generate.metadata.release_level stable
```

## Output Formats

Support multiple output formats via `--format` or `--json` flag:

```bash
# Default (human-readable)
librarian config list

# JSON
librarian config list --json

# YAML (for piping to tools)
librarian config get sources --format yaml
```

## Error Handling

**Clear, actionable error messages:**

```bash
$ librarian config get invalid.key
Error: Configuration key 'invalid.key' not found

Valid repository-level keys:
  - version
  - language
  - sources.googleapis.url
  - sources.googleapis.sha256
  ...

To list all keys: librarian config list
```

```bash
$ librarian config set language rust
Error: Unsupported language 'rust'

Supported languages: go, python, dart
```

```bash
$ librarian config set --edition nonexistent version 1.0.0
Error: Edition 'nonexistent' not found

Available editions:
  - secretmanager
  - pubsub

To add new edition: librarian config edition add <name>
```

## Implementation Considerations

### Config API

Build on existing `internal/config/config.go`:

```go
// Get retrieves a config value by dot-notation key
func (c *Config) Get(key string) (interface{}, error)

// Set updates a config value by dot-notation key
func (c *Config) Set(key string, value interface{}) error

// Delete removes a config key
func (c *Config) Delete(key string) error

// GetEdition retrieves edition-specific config
func (c *Config) GetEdition(name, key string) (interface{}, error)

// SetEdition updates edition-specific config
func (c *Config) SetEdition(name, key string, value interface{}) error

// AddEdition creates a new edition
func (c *Config) AddEdition(name string, apis []string) error

// DeleteEdition removes an edition
func (c *Config) DeleteEdition(name string) error

// Validate checks config for errors
func (c *Config) Validate() []error
```

### Dot Notation Parser

Implement key parser for dot notation:

```go
// ParseKey converts "sources.googleapis.url" to nested access
func ParseKey(key string) []string {
    return strings.Split(key, ".")
}

// GetNestedValue traverses config struct using parsed key
func GetNestedValue(cfg interface{}, parts []string) (interface{}, error)

// SetNestedValue updates config struct using parsed key
func SetNestedValue(cfg interface{}, parts []string, value interface{}) error
```

### YAML Preservation

**Challenge:** Preserving comments and formatting when updating YAML

**Solutions:**
1. Use comment-preserving YAML library (e.g., `gopkg.in/yaml.v3`)
2. For `edit` command, just open editor (no parsing needed)
3. For `set/delete`, use targeted updates instead of full rewrites

### Validation

Implement schema-based validation:

```go
type Schema struct {
    Fields map[string]FieldSpec
}

type FieldSpec struct {
    Type     string   // "string", "int", "bool", "array", "object"
    Required bool
    Enum     []string // For restricted values
    Pattern  string   // Regex for validation
}

func (s *Schema) Validate(config *Config) []error
```

## Alternatives Considered

### Alternative 1: npm-style Subcommands

```bash
librarian config set language go
librarian config get language
librarian config delete language
```

**Rejected because:**
- Verbose for scripting (need to type `set` for every change)
- Cannot compose multiple edits atomically
- Less efficient for tools/automation
- Separate commands mean separate file writes

### Alternative 2: Git-style Hybrid

```bash
librarian config language        # Get
librarian config language go     # Set
librarian config --unset language # Delete
```

**Rejected because:**
- Ambiguous (is 2nd arg a value or subcommand?)
- Harder to parse and validate
- Less clear intent

### Alternative 3: Opening Editor

```bash
librarian config edit  # Opens $EDITOR
```

**Rejected because:**
- Not scriptable
- Requires interactive session
- Hard to automate
- Manual YAML editing error-prone
- (But could be added as convenience command later)

### Alternative 4: Environment Variables

```bash
LIBRARIAN_LANGUAGE=go librarian generate
```

**Rejected because:**
- Not persistent
- Hard to discover available options
- Doesn't work well for complex nested config
- Use case: per-command overrides (could add later)

## Examples

### Complete Workflow: Adding New API Version

```bash
# 1. Check current APIs
librarian config get -edition secretmanager apis
# google/cloud/secretmanager/v1

# 2. Add new API version
librarian config -edition secretmanager -add-api google/cloud/secretmanager/v1beta2

# 3. Verify
librarian config get -edition secretmanager apis
# google/cloud/secretmanager/v1
# google/cloud/secretmanager/v1beta2

# 4. Regenerate code
librarian generate secretmanager
```

### Complete Workflow: Updating googleapis

```bash
# 1. Check current version
librarian config get sources.googleapis.url
# https://github.com/googleapis/googleapis/archive/abc123.tar.gz

# 2. Update using dedicated command (preferred)
librarian update --googleapis

# OR update manually via config command
librarian config \
  -set sources.googleapis.url=https://github.com/googleapis/googleapis/archive/xyz789.tar.gz \
  -set sources.googleapis.sha256=867048ec8f0850a4d77ad836319e4c0a0c624928611af8a900cd77e676164e8e

# 3. Regenerate all libraries
librarian generate --all
```

### Complete Workflow: Changing Metadata

```bash
# Update multiple metadata fields at once
librarian config -edition secretmanager \
  -set-metadata release_level=stable \
  -set-metadata product_documentation=https://cloud.google.com/secret-manager/docs \
  -set-metadata name_pretty="Secret Manager"

# View all metadata
librarian config get -edition secretmanager generate.metadata -json
```

### Complete Workflow: Setting Up New Edition

```bash
# Create edition and configure it in one command
librarian config \
  -add-edition secretmanager \
  -edition secretmanager \
  -add-api google/cloud/secretmanager/v1 \
  -add-api google/cloud/secretmanager/v1beta2 \
  -set-version 0.1.0 \
  -set-metadata name_pretty="Secret Manager" \
  -set-metadata release_level=stable

# Verify with print flag (doesn't write to file)
librarian config -edition secretmanager -print

# Now generate code
librarian generate secretmanager
```

### Scripting Example: Batch Updates

```bash
# Update multiple editions in a script
for edition in secretmanager pubsub storage; do
  librarian config -edition $edition -set-metadata release_level=stable
done

# Conditional updates based on edition state
if librarian config get -edition secretmanager version | grep -q "null"; then
  librarian config -edition secretmanager -set-version 0.1.0
fi
```

## Summary

The `librarian config` command provides a scriptable, composable interface for managing configuration, following the `go mod edit` pattern:

**Editing flags (modify config):**
- `-set <key>=<value>` / `-unset <key>` - Repository-level config
- `-add-edition <name>` / `-drop-edition <name>` - Edition management
- `-edition <name>` - Scope for edition-specific operations
- `-add-api <path>` / `-drop-api <path>` - API management
- `-set-metadata <key>=<value>` / `-unset-metadata <key>` - Metadata
- `-add-keep/-drop-keep`, `-add-remove/-drop-remove` - File patterns
- `-set-version <version>` - Edition version

**Output flags:**
- `-print` - Output YAML to stdout instead of writing
- `-json` - Output JSON to stdout instead of writing
- `-fmt` - Reformat without other changes

**Query subcommands:**
- `librarian config list` - View all config
- `librarian config get <key>` - Retrieve specific values
- `librarian config editions` - List all editions

**Key benefits:**
1. **Composable** - Multiple edits in one atomic operation
2. **Scriptable** - Perfect for automation and tools
3. **Familiar** - Matches `go mod edit` pattern Go developers know
4. **Safe** - All changes applied together or not at all
5. **Inspectable** - `-print` and `-json` for testing before applying
