package skillgen

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateTool tests basic tool generation from CommandIR
func TestGenerateTool(t *testing.T) {
	cmd := CommandIR{
		CommandPath: "confluent kafka topic list",
		Short:       "List Kafka topics.",
		Long:        "List Kafka topics in the cluster.",
		Example:     "",
		Flags: []FlagIR{
			{
				Name:     "cluster",
				Type:     "string",
				Required: true,
				Default:  "",
				Usage:    "Kafka cluster ID",
			},
			{
				Name:     "limit",
				Type:     "int",
				Required: false,
				Default:  "100",
				Usage:    "Maximum number of topics to list",
			},
		},
		Operation: "list",
		Resource:  "topic",
		Mode:      "cloud",
	}

	tool := GenerateTool(cmd, "kafka", "high")

	// Verify basic structure
	assert.Equal(t, "confluent_kafka_topic_list", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "List Kafka topics")

	// Verify InputSchema
	assert.Equal(t, "object", tool.InputSchema.Type)
	assert.Len(t, tool.InputSchema.Properties, 2)
	assert.Contains(t, tool.InputSchema.Properties, "cluster")
	assert.Contains(t, tool.InputSchema.Properties, "limit")
	assert.Equal(t, []string{"cluster"}, tool.InputSchema.Required)

	// Verify parameter schemas
	clusterSchema := tool.InputSchema.Properties["cluster"]
	assert.Equal(t, "string", clusterSchema.Type)
	assert.Equal(t, "Kafka cluster ID", clusterSchema.Description)

	limitSchema := tool.InputSchema.Properties["limit"]
	assert.Equal(t, "integer", limitSchema.Type)
	assert.Equal(t, 100, limitSchema.Default)
}

// TestGenerateTool_WithExample tests description formatting with example
func TestGenerateTool_WithExample(t *testing.T) {
	cmd := CommandIR{
		CommandPath: "confluent kafka topic create",
		Short:       "Create a Kafka topic.",
		Example:     "  $ confluent kafka topic create my-topic --cluster lkc-123456",
		Operation:   "create",
		Resource:    "topic",
	}

	tool := GenerateTool(cmd, "kafka", "high")

	// Verify description includes both Short and Example
	assert.Contains(t, tool.Description, "Create a Kafka topic")
	assert.Contains(t, tool.Description, "Example:")
	assert.Contains(t, tool.Description, "confluent kafka topic create my-topic")
}

// TestGenerateTool_ModeBoth tests single tool for dual-mode commands
func TestGenerateTool_ModeBoth(t *testing.T) {
	cmd := CommandIR{
		CommandPath: "confluent kafka topic list",
		Short:       "List Kafka topics.",
		Operation:   "list",
		Resource:    "topic",
		Mode:        "both",
	}

	tool := GenerateTool(cmd, "kafka", "high")

	// Should create single tool, not separate cloud/onprem versions
	assert.Equal(t, "confluent_kafka_topic_list", tool.Name)
	assert.NotContains(t, tool.Name, "_cloud")
	assert.NotContains(t, tool.Name, "_onprem")

	// Should have mode annotation (we'll verify structure exists)
	assert.NotNil(t, tool.Annotations)
}

// TestGenerateTool_ModeCloud tests cloud-only command
func TestGenerateTool_ModeCloud(t *testing.T) {
	cmd := CommandIR{
		CommandPath: "confluent kafka topic list",
		Short:       "List Kafka topics.",
		Operation:   "list",
		Resource:    "topic",
		Mode:        "cloud",
	}

	tool := GenerateTool(cmd, "kafka", "high")

	assert.Equal(t, "confluent_kafka_topic_list", tool.Name)
	assert.NotNil(t, tool.Annotations)
}

// TestGenerateTool_Annotations tests annotation setting
func TestGenerateTool_Annotations(t *testing.T) {
	tests := []struct {
		name              string
		operation         string
		expectedPriority  string
		expectedAudience  string
	}{
		{
			name:              "list operation high priority",
			operation:         "list",
			expectedPriority:  "high",
			expectedAudience:  "user,assistant",
		},
		{
			name:              "describe operation high priority",
			operation:         "describe",
			expectedPriority:  "high",
			expectedAudience:  "user,assistant",
		},
		{
			name:              "create operation medium priority",
			operation:         "create",
			expectedPriority:  "medium",
			expectedAudience:  "user,assistant",
		},
		{
			name:              "delete operation medium priority",
			operation:         "delete",
			expectedPriority:  "medium",
			expectedAudience:  "user,assistant",
		},
		{
			name:              "update operation medium priority",
			operation:         "update",
			expectedPriority:  "medium",
			expectedAudience:  "user,assistant",
		},
		{
			name:              "other operation low priority",
			operation:         "other",
			expectedPriority:  "low",
			expectedAudience:  "user,assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CommandIR{
				CommandPath: "confluent kafka topic " + tt.operation,
				Short:       "Test command",
				Operation:   tt.operation,
				Resource:    "topic",
			}

			tool := GenerateTool(cmd, "kafka", "high")

			assert.Equal(t, tt.expectedPriority, tool.Annotations.Priority)
			assert.Equal(t, tt.expectedAudience, tool.Annotations.Audience)
		})
	}
}

// TestGenerateTool_HighTier tests high tier tool generation
func TestGenerateTool_HighTier(t *testing.T) {
	cmd := CommandIR{
		CommandPath: "confluent kafka topic list",
		Short:       "List Kafka topics.",
		Operation:   "list",
		Resource:    "topic",
	}

	tool := GenerateTool(cmd, "kafka", "high")

	// High tier: confluent_namespace_resource_operation
	assert.Equal(t, "confluent_kafka_topic_list", tool.Name)
}

// TestGenerateTool_MediumTier tests medium tier tool generation
func TestGenerateTool_MediumTier(t *testing.T) {
	cmd := CommandIR{
		CommandPath: "confluent schema-registry create",
		Short:       "Create a Schema Registry.",
		Operation:   "create",
		Resource:    "",
	}

	tool := GenerateTool(cmd, "schema-registry", "medium")

	// Medium tier: confluent_namespace_operation
	assert.Equal(t, "confluent_schema_registry_create", tool.Name)
}

// TestGenerateTool_LowTier tests low tier tool generation
func TestGenerateTool_LowTier(t *testing.T) {
	cmd := CommandIR{
		CommandPath: "confluent plugin install",
		Short:       "Install a plugin.",
		Operation:   "other",
		Resource:    "",
	}

	tool := GenerateTool(cmd, "plugin", "low")

	// Low tier: confluent_namespace
	assert.Equal(t, "confluent_plugin", tool.Name)
}

// TestGenerateManifest tests full manifest generation
func TestGenerateManifest(t *testing.T) {
	ir := IR{
		Metadata: Metadata{
			CLIVersion:  "v3.50.0",
			GeneratedAt: time.Now(),
		},
		Commands: []CommandIR{
			{
				CommandPath: "confluent kafka topic list",
				Short:       "List Kafka topics.",
				Operation:   "list",
				Resource:    "topic",
				Mode:        "cloud",
			},
			{
				CommandPath: "confluent kafka topic create",
				Short:       "Create a Kafka topic.",
				Operation:   "create",
				Resource:    "topic",
				Mode:        "cloud",
			},
			{
				CommandPath: "confluent kafka topic delete",
				Short:       "Delete a Kafka topic.",
				Operation:   "delete",
				Resource:    "topic",
				Mode:        "cloud",
			},
			{
				CommandPath: "confluent plugin install",
				Short:       "Install a plugin.",
				Operation:   "other",
				Resource:    "",
				Mode:        "both",
			},
		},
	}

	manifest := GenerateManifest(ir)

	// Verify metadata
	assert.NotNil(t, manifest.Metadata)
	assert.Equal(t, "v3.50.0", manifest.Metadata.CLIVersion)
	assert.NotEmpty(t, manifest.Metadata.GeneratedAt)
	assert.NotEmpty(t, manifest.Metadata.GeneratorVersion)
	assert.Greater(t, manifest.Metadata.SkillCount, 0)
	assert.Equal(t, 4, manifest.Metadata.CommandCount)

	// Verify tools were generated
	assert.NotEmpty(t, manifest.Tools)
	assert.Greater(t, len(manifest.Tools), 0)
}

// TestGenerateManifest_CLIVersionProvenance tests GEN-05 requirement
func TestGenerateManifest_CLIVersionProvenance(t *testing.T) {
	testVersion := "v3.50.0"
	ir := IR{
		Metadata: Metadata{
			CLIVersion:  testVersion,
			GeneratedAt: time.Now(),
		},
		Commands: []CommandIR{
			{
				CommandPath: "confluent kafka topic list",
				Short:       "List topics",
				Operation:   "list",
				Resource:    "topic",
			},
		},
	}

	manifest := GenerateManifest(ir)

	// CRITICAL: ManifestMetadata.CLIVersion must match IR.Metadata.CLIVersion
	require.NotNil(t, manifest.Metadata)
	assert.Equal(t, testVersion, manifest.Metadata.CLIVersion,
		"ManifestMetadata.CLIVersion must be populated from IR.Metadata.CLIVersion (GEN-05)")
}

// TestGenerateManifest_TierGrouping tests tier-based grouping
func TestGenerateManifest_TierGrouping(t *testing.T) {
	// Create commands for high-tier namespace (many commands with resource variations)
	highTierCommands := []CommandIR{}
	for _, op := range []string{"list", "create", "delete"} {
		for _, res := range []string{"topic", "acl"} {
			highTierCommands = append(highTierCommands, CommandIR{
				CommandPath: "confluent kafka " + res + " " + op,
				Short:       op + " " + res,
				Operation:   op,
				Resource:    res,
			})
		}
	}

	// Create commands for low-tier namespace (few commands)
	lowTierCommands := []CommandIR{
		{
			CommandPath: "confluent plugin install",
			Short:       "Install plugin",
			Operation:   "other",
			Resource:    "",
		},
	}

	ir := IR{
		Metadata: Metadata{
			CLIVersion:  "v3.50.0",
			GeneratedAt: time.Now(),
		},
		Commands: append(highTierCommands, lowTierCommands...),
	}

	manifest := GenerateManifest(ir)

	// Verify tools exist
	assert.NotEmpty(t, manifest.Tools)

	// Find tools by namespace
	kafkaTools := []Tool{}
	pluginTools := []Tool{}
	for _, tool := range manifest.Tools {
		if Contains(tool.Name, "kafka") {
			kafkaTools = append(kafkaTools, tool)
		}
		if Contains(tool.Name, "plugin") {
			pluginTools = append(pluginTools, tool)
		}
	}

	// High tier should create multiple tools (one per resource+operation)
	assert.Greater(t, len(kafkaTools), 1, "High tier namespace should create multiple tools")

	// Low tier should create single tool
	assert.Equal(t, 1, len(pluginTools), "Low tier namespace should create single tool")
}

// Helper function for string contains check
func Contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestGenerateToolFromCommand_Operations validates generator produces valid skill definitions for all operation types (TEST-02)
func TestGenerateToolFromCommand_Operations(t *testing.T) {
	tests := []struct {
		name              string
		operation         string
		resource          string
		expectedPriority  string
		descriptionCheck  string
	}{
		{
			name:              "list operation",
			operation:         "list",
			resource:          "topic",
			expectedPriority:  "high",
			descriptionCheck:  "List",
		},
		{
			name:              "create operation",
			operation:         "create",
			resource:          "topic",
			expectedPriority:  "medium",
			descriptionCheck:  "Create",
		},
		{
			name:              "delete operation",
			operation:         "delete",
			resource:          "topic",
			expectedPriority:  "medium",
			descriptionCheck:  "Delete",
		},
		{
			name:              "describe operation",
			operation:         "describe",
			resource:          "cluster",
			expectedPriority:  "high",
			descriptionCheck:  "Describe",
		},
		{
			name:              "update operation",
			operation:         "update",
			resource:          "topic",
			expectedPriority:  "medium",
			descriptionCheck:  "Update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CommandIR{
				CommandPath: "confluent kafka " + tt.resource + " " + tt.operation,
				Short:       tt.descriptionCheck + " a " + tt.resource,
				Operation:   tt.operation,
				Resource:    tt.resource,
				Mode:        "cloud",
				Flags: []FlagIR{
					{
						Name:     "cluster",
						Type:     "string",
						Required: true,
						Usage:    "Cluster ID",
					},
				},
			}

			tool := GenerateTool(cmd, "kafka", "high")

			// Verify tool name follows convention
			assert.NotEmpty(t, tool.Name, "Tool name should not be empty")
			assert.Contains(t, tool.Name, "confluent", "Tool name should start with confluent")
			assert.Contains(t, tool.Name, tt.operation, "Tool name should contain operation")

			// Verify description
			assert.Contains(t, tool.Description, tt.descriptionCheck, "Description should mention operation")

			// Verify InputSchema structure
			assert.Equal(t, "object", tool.InputSchema.Type, "InputSchema should be object type")
			assert.NotEmpty(t, tool.InputSchema.Properties, "InputSchema should have properties")

			// Verify annotations
			assert.Equal(t, tt.expectedPriority, tool.Annotations.Priority, "Priority should match operation type")
			assert.Equal(t, "user,assistant", tool.Annotations.Audience, "Audience should be user,assistant")
		})
	}
}

// TestGenerateInputSchema_RequiredVsOptional validates required parameter detection in JSON Schema output (TEST-02)
func TestGenerateInputSchema_RequiredVsOptional(t *testing.T) {
	flags := []FlagIR{
		{
			Name:     "cluster",
			Type:     "string",
			Required: true,
			Usage:    "Cluster ID",
		},
		{
			Name:     "environment",
			Type:     "string",
			Required: true,
			Usage:    "Environment ID",
		},
		{
			Name:     "limit",
			Type:     "int",
			Required: false,
			Default:  "100",
			Usage:    "Result limit",
		},
		{
			Name:     "output",
			Type:     "string",
			Required: false,
			Default:  "human",
			Usage:    "Output format",
		},
	}

	schema := BuildInputSchema(flags)

	// Verify schema structure
	assert.Equal(t, "object", schema.Type, "Schema should be object type")
	assert.Len(t, schema.Properties, 4, "Schema should have 4 properties")

	// Verify required array contains only required flags
	assert.Len(t, schema.Required, 2, "Should have exactly 2 required flags")
	assert.Contains(t, schema.Required, "cluster", "cluster should be required")
	assert.Contains(t, schema.Required, "environment", "environment should be required")
	assert.NotContains(t, schema.Required, "limit", "limit should not be required")
	assert.NotContains(t, schema.Required, "output", "output should not be required")

	// Verify property schemas
	assert.Equal(t, "string", schema.Properties["cluster"].Type)
	assert.Equal(t, "Cluster ID", schema.Properties["cluster"].Description)

	assert.Equal(t, "integer", schema.Properties["limit"].Type)
	assert.Equal(t, 100, schema.Properties["limit"].Default)

	assert.Equal(t, "string", schema.Properties["output"].Type)
	assert.Equal(t, "human", schema.Properties["output"].Default)
}

// TestApplyNamingConvention_HyphenatedResources validates naming convention handles hyphenated resources correctly (TEST-02)
func TestApplyNamingConvention_HyphenatedResources(t *testing.T) {
	tests := []struct {
		name         string
		commandPath  string
		namespace    string
		resource     string
		operation    string
		tier         string
		expectedName string
	}{
		{
			name:         "service-account to service_account",
			commandPath:  "confluent iam service-account list",
			namespace:    "iam",
			resource:     "service-account",
			operation:    "list",
			tier:         "high",
			expectedName: "confluent_iam_service_account_list",
		},
		{
			name:         "api-key to api_key",
			commandPath:  "confluent api-key create",
			namespace:    "api-key",
			resource:     "api-key",
			operation:    "create",
			tier:         "medium",
			expectedName: "confluent_api_key_create",
		},
		{
			name:         "schema-registry to schema_registry",
			commandPath:  "confluent schema-registry cluster describe",
			namespace:    "schema-registry",
			resource:     "cluster",
			operation:    "describe",
			tier:         "high",
			expectedName: "confluent_schema_registry_cluster_describe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolName := BuildToolName(tt.namespace, tt.resource, tt.operation, tt.tier)

			// Verify hyphens are converted to underscores
			assert.Equal(t, tt.expectedName, toolName, "Tool name should follow naming convention")
			assert.NotContains(t, toolName, "-", "Tool name should not contain hyphens")
			assert.Contains(t, toolName, "_", "Tool name should use underscores")
		})
	}
}
