# CCPM (Custom Connect Plugin Management) CLI Commands

This directory contains the CLI commands for managing Custom Connect Plugin Management (CCPM) resources.

## Overview

The CCPM CLI provides commands to manage:
- **Custom Connect Plugins**: The main plugin artifacts containing connector and SMT jars
- **Plugin Versions**: Different versions of plugins with specific configurations

## Command Structure

```
confluent ccpm [command]
```

### Available Commands

#### Plugin Management
- `confluent ccpm plugin list` - List Custom Connect Plugins
- `confluent ccpm plugin create` - Create a Custom Connect Plugin
- `confluent ccpm plugin describe <id>` - Describe a Custom Connect Plugin
- `confluent ccpm plugin update <id>` - Update a Custom Connect Plugin
- `confluent ccpm plugin delete <id>` - Delete a Custom Connect Plugin

#### Version Management
- `confluent ccpm plugin version list` - List Custom Connect Plugin Versions
- `confluent ccpm plugin version create` - Create a Custom Connect Plugin Version (handles upload internally)
- `confluent ccpm plugin version describe <version-id>` - Describe a Custom Connect Plugin Version
- `confluent ccpm plugin version delete <version-id>` - Delete a Custom Connect Plugin Version

## API Resources

Based on the CCPM API specification, the following resources are supported:

### CustomConnectPlugin
- **Path**: `/ccpm/v1/plugins`
- **Operations**: create, read, update, delete, list
- **Attributes**:
  - `display_name` (required): Display name of the plugin
  - `description` (optional): Description of the plugin
  - `cloud` (required, immutable): Cloud provider (AWS, GCP, AZURE)
  - `runtime_language` (read-only): Runtime language of the plugin
- **Relationships**:
  - `environment` (required): Belongs to environment

### CustomConnectPluginVersion
- **Path**: `/ccpm/v1/plugins/{plugin_id}/versions`
- **Operations**: create, read, delete, list
- **Attributes**:
  - `version` (required, immutable): Version string (must comply with SemVer)
  - `sensitive_config_properties` (optional, immutable): Array of sensitive property names
  - `documentation_link` (optional, immutable): Documentation link URL
  - `content_format` (read-only): Archive format (ZIP, JAR)
  - `connector_classes` (required, immutable): Array of connector class definitions
  - `upload_source` (required for create): Upload source configuration
- **Status**:
  - `phase`: Processing state (PROCESSING, READY, FAILED)
  - `error_message`: Error message if failed
- **Relationships**:
  - `environment` (required): Belongs to environment

## Detailed Command Reference

### Plugin Commands

#### `confluent ccpm plugin create`
Creates a new Custom Connect Plugin.

**Required Flags:**
- `--name`: Display name of the custom Connect plugin
- `--cloud`: Cloud provider (AWS, GCP, AZURE)
- `--environment`: Environment ID

**Optional Flags:**
- `--description`: Description of the custom Connect plugin
- `--output`: Output format (json, yaml, table)

**Examples:**
```bash
# Create a custom Connect plugin for AWS
confluent ccpm plugin create --name "My Custom Plugin" --description "A custom connector for data processing" --cloud AWS --environment env-12345

# Create a custom Connect plugin for GCP with minimal description
confluent ccpm plugin create --name "GCP Data Connector" --cloud GCP --environment env-abcdef
```

#### `confluent ccpm plugin list`
Lists all Custom Connect Plugins in an environment.

**Required Flags:**
- `--environment`: Environment ID

**Optional Flags:**
- `--cloud`: Filter by cloud provider (AWS, GCP, AZURE)
- `--output`: Output format (json, yaml, table)

**Examples:**
```bash
# List all custom Connect plugins in an environment
confluent ccpm plugin list --environment env-12345

# List custom Connect plugins filtered by cloud provider
confluent ccpm plugin list --environment env-12345 --cloud AWS
```

#### `confluent ccpm plugin describe <id>`
Describes a specific Custom Connect Plugin.

**Required Flags:**
- `--environment`: Environment ID

**Optional Flags:**
- `--output`: Output format (json, yaml, table)

**Examples:**
```bash
# Describe a custom Connect plugin by ID
confluent ccpm plugin describe plugin-123456 --environment env-12345
```

#### `confluent ccpm plugin update <id>`
Updates a Custom Connect Plugin.

**Required Flags:**
- `--environment`: Environment ID

**Optional Flags:**
- `--name`: New display name of the custom Connect plugin
- `--description`: New description of the custom Connect plugin

**Examples:**
```bash
# Update the name and description of a custom Connect plugin
confluent ccpm plugin update plugin-123456 --name "Updated Plugin Name" --description "Updated description" --environment env-12345

# Update only the name of a custom Connect plugin
confluent ccpm plugin update plugin-123456 --name "New Plugin Name" --environment env-12345
```

#### `confluent ccpm plugin delete <id>`
Deletes a Custom Connect Plugin.

**Required Flags:**
- `--environment`: Environment ID

**Optional Flags:**
- `--force`: Force delete without confirmation

**Examples:**
```bash
# Delete a custom Connect plugin by ID
confluent ccpm plugin delete plugin-123456 --environment env-12345

# Force delete a custom Connect plugin without confirmation
confluent ccpm plugin delete plugin-123456 --environment env-12345 --force
```

### Version Commands

#### `confluent ccpm plugin version create`
Creates a new version of a Custom Connect Plugin.

**Required Flags:**
- `--plugin`: Plugin ID
- `--version`: Version of the custom Connect plugin (must comply with SemVer)
- `--environment`: Environment ID
- `--plugin-file`: Custom plugin ZIP or JAR file
- `--connector-classes`: A comma-separated list of connector classes in format 'class_name:type'

**Optional Flags:**
- `--sensitive-properties`: A comma-separated list of sensitive configuration property names
- `--documentation-link`: URL to the plugin documentation
- `--output`: Output format (json, yaml, table)

**Connector Class Format:**
- Format: `class_name:type`
- Types: `SOURCE` or `SINK`
- Example: `io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE`

**Examples:**
```bash
# Create a new version 1.0.0 of a custom connect plugin
confluent ccpm plugin version create --plugin plugin-123456 --version 1.0.0 --environment env-abcdef --plugin-file datagen.zip --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE'

# Create a new version 2.1.0 of a custom connect plugin with multiple connector classes and optional fields
confluent ccpm plugin version create --plugin plugin-123456 --version 2.1.0 --environment env-abcdef --plugin-file datagen.zip --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE,io.confluent.kafka.connect.sink.SinkConnector:SINK' --sensitive-properties 'passwords,keys,tokens' --documentation-link 'https://github.com/confluentinc/kafka-connect-datagen'
```

#### `confluent ccpm plugin version list`
Lists all versions of a Custom Connect Plugin.

**Required Flags:**
- `--plugin`: Plugin ID
- `--environment`: Environment ID

**Optional Flags:**
- `--output`: Output format (json, yaml, table)

**Examples:**
```bash
# List all versions of a custom connect plugin
confluent ccpm plugin version list --plugin plugin-123456 --environment env-abcdef
```

#### `confluent ccpm plugin version describe <version-id>`
Describes a specific version of a Custom Connect Plugin.

**Required Flags:**
- `--plugin`: Plugin ID
- `--environment`: Environment ID

**Optional Flags:**
- `--output`: Output format (json, yaml, table)

**Examples:**
```bash
# Describe a specific version of a custom connect plugin
confluent ccpm plugin version describe version-789012 --plugin plugin-123456 --environment env-abcdef

# Get detailed information about version 1.0.0 of a plugin
confluent ccpm plugin version describe version-1.0.0 --plugin plugin-123456 --environment env-abcdef
```

#### `confluent ccpm plugin version delete <version-id>`
Deletes a specific version of a Custom Connect Plugin.

**Required Flags:**
- `--plugin`: Plugin ID
- `--environment`: Environment ID

**Optional Flags:**
- `--force`: Force delete without confirmation

**Examples:**
```bash
# Delete a specific version of a custom connect plugin
confluent ccpm plugin version delete ver-789012 --plugin plugin-123456 --environment env-abcdef

# Force delete a plugin version without confirmation
confluent ccpm plugin version delete ver-789012 --plugin plugin-123456 --environment env-abcdef --force
```

## Output Fields

### Plugin Output Fields
- `ID`: Unique identifier for the plugin
- `Name`: Display name of the plugin
- `Description`: Description of the plugin
- `Cloud`: Cloud provider (AWS, GCP, AZURE)
- `Runtime Language`: Runtime language of the plugin (read-only)
- `Environment`: Environment ID

### Version Output Fields
- `Plugin ID`: ID of the parent plugin
- `Plugin Name`: Name of the parent plugin
- `ID`: Unique identifier for the version
- `Version`: Version string (SemVer format)
- `Content Format`: Archive format (ZIP, JAR)
- `Documentation Link`: URL to plugin documentation
- `Sensitive Config Properties`: Array of sensitive property names
- `Connector Classes`: Comma-separated list of connector classes with types
- `Phase`: Processing state (PROCESSING, READY, FAILED)
- `Error Message`: Error message if processing failed
- `Environment`: Environment ID

## File Structure

```
internal/ccpm/
├── command.go                    # Main command entry point
├── command_plugin.go             # Plugin command structure and output formatting
├── command_plugin_create.go      # Plugin create command
├── command_plugin_delete.go      # Plugin delete command
├── command_plugin_describe.go    # Plugin describe command
├── command_plugin_list.go        # Plugin list command
├── command_plugin_update.go      # Plugin update command
├── command_version.go            # Version command structure and output formatting
├── command_version_create.go     # Version create command
├── command_version_delete.go     # Version delete command
├── command_version_describe.go   # Version describe command
├── command_version_list.go       # Version list command
└── README.md                     # This documentation
```

## Integration Points

- **Main Command**: Added to `internal/command.go`
- **Client**: Uses `pkg/ccloudv2/ccpm.go` for API calls
- **Authentication**: Uses existing authentication flow via `pcmd.AuthenticatedCLICommand`
- **Output**: Uses existing output formatting via `pkg/output`
- **File Upload**: Supports ZIP and JAR files with automatic presigned URL handling
- **Cloud Support**: Supports AWS, GCP, and Azure with cloud-specific upload handling

## Implementation Notes

- **File Upload**: The version create command automatically handles file upload using presigned URLs
- **Cloud Provider**: File upload method varies by cloud provider (Azure uses blob storage, others use form data)
- **Validation**: Connector classes must be in format `class_name:type` where type is either `SOURCE` or `SINK`
- **SemVer**: Version strings must comply with Semantic Versioning (e.g., 1.0.0, 2.1.0)
- **Immutable Fields**: Cloud provider and version are immutable after creation
- **Status Tracking**: Plugin versions have processing states (PROCESSING, READY, FAILED) with error messages 