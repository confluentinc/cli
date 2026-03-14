package skillgen

import (
	"regexp"
	"strings"
)

// BuildToolName constructs an MCP tool name based on namespace, resource, operation, and tier.
//
// Tool naming follows confluent_namespace_resource_operation pattern with tier-specific rules:
//   - High tier: confluent_namespace_resource_operation (e.g., confluent_kafka_topic_list)
//   - Medium tier: confluent_namespace_operation (e.g., confluent_schema_registry_create)
//   - Low tier: confluent_namespace (e.g., confluent_plugin)
//
// Special handling:
//   - Hyphens in namespace/resource names are converted to underscores
//   - Special commands (login, logout, version, update) get dedicated names (confluent_login, etc.)
//
// All names are validated against MCP SEP-986: [a-zA-Z0-9_\-./]{1,64}
func BuildToolName(namespace, resource, operation, tier string) string {
	// Sanitize all components (replace hyphens with underscores)
	namespace = SanitizeName(namespace)
	resource = SanitizeName(resource)
	operation = SanitizeName(operation)

	var name string

	// Check for special commands that get dedicated names
	if tier == "low" && (namespace == "login" || namespace == "logout" ||
		namespace == "version" || namespace == "update") {
		name = "confluent_" + namespace
	} else {
		// Build name based on tier
		switch tier {
		case "high":
			// High tier: confluent_namespace_resource_operation
			name = "confluent_" + namespace + "_" + resource + "_" + operation
		case "medium":
			// Medium tier: confluent_namespace_operation
			name = "confluent_" + namespace + "_" + operation
		case "low":
			// Low tier: confluent_namespace
			name = "confluent_" + namespace
		default:
			// Fallback to high tier format
			name = "confluent_" + namespace + "_" + resource + "_" + operation
		}
	}

	// Truncate to 64 characters (MCP SEP-986 limit), preserving the
	// operation suffix to avoid collisions between sibling commands.
	if len(name) > 64 {
		lastUnderscore := strings.LastIndex(name, "_")
		if lastUnderscore > 0 {
			suffix := name[lastUnderscore:]
			prefix := name[:64-len(suffix)]
			name = prefix + suffix
		} else {
			name = name[:64]
		}
	}

	return name
}

// SanitizeName converts hyphens to underscores to comply with MCP naming conventions.
// MCP tool names can contain hyphens, but we use underscores for consistency with
// the confluent_namespace_resource_operation pattern.
func SanitizeName(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

// ValidateToolName checks if a tool name complies with MCP SEP-986 specification.
//
// MCP tool name requirements:
//   - Length: 1-64 characters
//   - Allowed characters: a-z A-Z 0-9 _ - . /
//   - Case-sensitive
//
// Returns true if the name is valid, false otherwise.
func ValidateToolName(name string) bool {
	// Check length
	if len(name) < 1 || len(name) > 64 {
		return false
	}

	// Check characters (MCP SEP-986: [a-zA-Z0-9_\-./])
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-./]+$`)
	return validPattern.MatchString(name)
}
