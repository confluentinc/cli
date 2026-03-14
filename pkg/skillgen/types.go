// Package skillgen provides command metadata extraction from the Confluent CLI's
// Cobra command tree to produce an intermediate representation (IR) for Claude
// skill generation.
//
// The IR serves as a bridge between the CLI's internal command structure and the
// final Claude skill JSON format. It captures all necessary metadata including
// command paths, flags, help text, annotations, and inferred semantic information
// (operation type, resource type).
package skillgen

import "time"

// IR is the root structure containing all extracted command metadata along with
// generation metadata. This is the top-level structure that gets serialized to
// JSON for consumption by the skill generator.
type IR struct {
	Metadata Metadata    `json:"metadata"`
	Commands []CommandIR `json:"commands"`
}

// Metadata contains information about the CLI version and when the IR was
// generated. This helps track which version of the CLI the skills were
// generated from.
type Metadata struct {
	CLIVersion  string    `json:"cli_version"`
	GeneratedAt time.Time `json:"generated_at"`
}

// CommandIR represents a single CLI command's metadata including its path,
// help text, flags, and inferred semantic information (operation, resource).
// This is the intermediate representation that gets populated by the parser
// and used by the skill generator.
type CommandIR struct {
	// CommandPath is the full command path (e.g., "confluent kafka topic list")
	CommandPath string `json:"command_path"`

	// Short is the short help text from Cobra's Short field
	Short string `json:"short"`

	// Long is the long help text from Cobra's Long field
	Long string `json:"long"`

	// Example is the usage example from Cobra's Example field
	Example string `json:"example"`

	// Flags contains metadata for all flags defined on this command
	Flags []FlagIR `json:"flags"`

	// Annotations contains raw Cobra annotations (e.g., run_requirement)
	Annotations map[string]string `json:"annotations"`

	// Operation is the inferred operation type (e.g., "list", "create", "delete")
	// populated by the classifier
	Operation string `json:"operation"`

	// Resource is the inferred resource type (e.g., "kafka-topic", "iam-user")
	// populated by the classifier
	Resource string `json:"resource"`

	// Mode indicates whether this command works in "cloud", "onprem", or "both"
	// modes, derived from Cobra annotations
	Mode string `json:"mode"`
}

// FlagIR represents metadata for a single command flag.
type FlagIR struct {
	// Name is the flag name (without -- prefix)
	Name string `json:"name"`

	// Type is the flag value type (from flag.Value.Type())
	Type string `json:"type"`

	// Required indicates whether this flag is required
	Required bool `json:"required"`

	// Default is the default value as a string
	Default string `json:"default"`

	// Usage is the help text for this flag
	Usage string `json:"usage"`
}

// Tool represents an MCP (Model Context Protocol) tool definition for Claude.
// This structure defines a skill that Claude can invoke.
type Tool struct {
	// Name is the unique tool identifier (e.g., "confluent-kafka-topic-list")
	Name string `json:"name"`

	// Title is the human-readable tool name
	Title string `json:"title"`

	// Description explains what the tool does
	Description string `json:"description"`

	// CommandPath is the CLI command path (e.g., "kafka cluster list").
	// Used by the MCP server to map tool invocations to CLI commands.
	CommandPath string `json:"command_path"`

	// InputSchema defines the parameters this tool accepts
	InputSchema InputSchema `json:"inputSchema"`

	// Annotations contains metadata about the tool
	Annotations Annotations `json:"annotations"`
}

// InputSchema defines the JSON Schema for tool parameters.
// It follows the JSON Schema specification for object validation.
type InputSchema struct {
	// Type is always "object" for MCP tools
	Type string `json:"type"`

	// Properties maps parameter names to their JSON Schema definitions
	Properties map[string]JSONSchema `json:"properties"`

	// Required lists the names of required parameters
	Required []string `json:"required"`
}

// JSONSchema represents a JSON Schema definition for a single parameter.
// Supports primitive types (string, boolean, integer), arrays, and objects.
type JSONSchema struct {
	// Type is the JSON Schema type (string, boolean, integer, array, object)
	Type string `json:"type"`

	// Default is the default value (type must match Type field)
	Default interface{} `json:"default,omitempty"`

	// Description is the parameter description (from flag usage)
	Description string `json:"description,omitempty"`

	// Items defines the schema for array elements (only for type="array")
	Items *JSONSchema `json:"items,omitempty"`

	// AdditionalProperties defines the schema for object values (only for type="object")
	AdditionalProperties *JSONSchema `json:"additionalProperties,omitempty"`
}

// Annotations contains metadata about a tool.
type Annotations struct {
	// Audience indicates who should use this tool (e.g., "customers", "internal")
	Audience string `json:"audience,omitempty"`

	// Priority indicates tool importance (e.g., "high", "medium", "low")
	Priority string `json:"priority,omitempty"`

	// LastModified is the last modification date
	LastModified string `json:"lastModified,omitempty"`
}

// SkillManifest represents the complete output of skill generation.
// It contains all generated tools plus metadata about the generation process.
type SkillManifest struct {
	// Metadata contains provenance information about the generated skills
	Metadata ManifestMetadata `json:"metadata"`

	// Tools is the list of all generated MCP tool definitions
	Tools []Tool `json:"tools"`
}

// ManifestMetadata contains provenance and statistics about skill generation.
type ManifestMetadata struct {
	// CLIVersion is the version of the CLI the skills were generated from
	CLIVersion string `json:"cli_version"`

	// GeneratedAt is when the skills were generated
	GeneratedAt string `json:"generated_at"`

	// GeneratorVersion is the version of the skill generator
	GeneratorVersion string `json:"generator_version"`

	// SkillCount is the total number of skills generated
	SkillCount int `json:"skill_count"`

	// CommandCount is the total number of CLI commands processed
	CommandCount int `json:"command_count"`
}
