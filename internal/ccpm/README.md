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
  - `description`: Description of the plugin
  - `cloud` (required, immutable): Cloud provider (AWS, GCP, AZURE)
  - `runtime_language` (read-only): Runtime language of the plugin
- **Relationships**:
  - `environment` (required): Belongs to environment

### CustomConnectPluginVersion
- **Path**: `/ccpm/v1/plugins/{plugin_id}/versions`
- **Operations**: create, read, delete, list
- **Attributes**:
  - `version` (required, immutable): Version string (must comply with SemVer)
  - `sensitive_config_properties` (immutable): Array of sensitive property names
  - `documentation_link` (immutable): Documentation link URL
  - `content_format` (read-only): Archive format (ZIP, JAR)
  - `connector_classes` (required, immutable): Array of connector class definitions
  - `upload_source` (required for create): Upload source configuration
- **Status**:
  - `phase`: Processing state (PROCESSING, READY, FAILED)
  - `error_message`: Error message if failed
- **Relationships**:
  - `environment` (required): Belongs to environment

## Implementation Status

**Note**: The current implementation contains placeholder commands that display informative messages. The actual API integration will be implemented when the CCPM SDK becomes available.

### TODO Items
1. Add CCPM SDK dependency to `go.mod`
2. Implement actual API calls in `pkg/ccloudv2/ccpm.go`
3. Replace placeholder implementations in command files with real API calls
4. Add proper error handling and validation
5. Add comprehensive tests
6. Add documentation and examples

## Usage Examples

```bash
# List all plugins
confluent ccpm plugin list --environment env-12345

# Create a plugin
confluent ccpm plugin create --name "My Custom Plugin" --description "A custom connector" --cloud AWS --environment env-12345

# Create a plugin version (upload handled automatically)
confluent ccpm plugin version create --plugin plugin-123456 --version 1.0.0 --environment env-abcdef --plugin-file datagen.zip --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE'

# List plugin versions
confluent ccpm plugin version list --plugin plugin-123456 --environment env-abcdef
```

## File Structure

```
internal/ccpm/
├── command.go                    # Main command entry point
├── command_plugin.go             # Plugin command structure
├── command_plugin_create.go      # Plugin create command
├── command_plugin_delete.go      # Plugin delete command
├── command_plugin_describe.go    # Plugin describe command
├── command_plugin_list.go        # Plugin list command
├── command_plugin_update.go      # Plugin update command
├── command_version.go            # Version command structure
└── README.md                     # This documentation
```

## Integration Points

- **Main Command**: Added to `internal/command.go`
- **Client**: Placeholder in `pkg/ccloudv2/ccpm.go`
- **Authentication**: Uses existing authentication flow via `pcmd.AuthenticatedCLICommand`
- **Output**: Uses existing output formatting via `pkg/output` 