package main

import (
	"testing"

	"github.com/confluentinc/cli/v4/pkg/skillgen"
)

func TestValidateManifestData_Valid(t *testing.T) {
	manifest := &skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "v1.0.0",
			GeneratedAt:      "2026-03-12T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       2,
			CommandCount:     2,
		},
		Tools: []skillgen.Tool{
			{
				Name:        "confluent-kafka-topic-list",
				Title:       "List Kafka Topics",
				Description: "Lists all Kafka topics in the cluster",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
			{
				Name:        "confluent-kafka-cluster-list",
				Title:       "List Kafka Clusters",
				Description: "Lists all Kafka clusters",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
		},
	}

	err := validateManifestData(manifest)
	if err != nil {
		t.Errorf("validateManifestData() with valid manifest returned error: %v", err)
	}
}

func TestValidateManifestData_MissingVersion(t *testing.T) {
	manifest := &skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "", // Missing version
			GeneratedAt:      "2026-03-12T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       1,
			CommandCount:     1,
		},
		Tools: []skillgen.Tool{
			{
				Name:        "confluent-kafka-topic-list",
				Title:       "List Kafka Topics",
				Description: "Lists all Kafka topics in the cluster",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
		},
	}

	err := validateManifestData(manifest)
	if err == nil {
		t.Error("validateManifestData() with missing version should return error, got nil")
	}
}

func TestValidateManifestData_CountMismatch(t *testing.T) {
	manifest := &skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "v1.0.0",
			GeneratedAt:      "2026-03-12T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       5, // Mismatch: claims 5 but only has 2
			CommandCount:     2,
		},
		Tools: []skillgen.Tool{
			{
				Name:        "confluent-kafka-topic-list",
				Title:       "List Kafka Topics",
				Description: "Lists all Kafka topics in the cluster",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
			{
				Name:        "confluent-kafka-cluster-list",
				Title:       "List Kafka Clusters",
				Description: "Lists all Kafka clusters",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
		},
	}

	err := validateManifestData(manifest)
	if err == nil {
		t.Error("validateManifestData() with count mismatch should return error, got nil")
	}
}

func TestValidateManifestData_ExceedsLimit(t *testing.T) {
	// Create a manifest with 500 tools (should fail, must be < 500)
	tools := make([]skillgen.Tool, 500)
	for i := 0; i < 500; i++ {
		tools[i] = skillgen.Tool{
			Name:        "tool-" + string(rune(i)),
			Title:       "Tool Title",
			Description: "Tool Description",
			InputSchema: skillgen.InputSchema{
				Type:       "object",
				Properties: map[string]skillgen.JSONSchema{},
			},
		}
	}

	manifest := &skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "v1.0.0",
			GeneratedAt:      "2026-03-12T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       500,
			CommandCount:     500,
		},
		Tools: tools,
	}

	err := validateManifestData(manifest)
	if err == nil {
		t.Error("validateManifestData() with 500 tools should return error (limit is < 500), got nil")
	}
}

func TestValidateManifestData_DuplicateNames(t *testing.T) {
	manifest := &skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "v1.0.0",
			GeneratedAt:      "2026-03-12T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       2,
			CommandCount:     2,
		},
		Tools: []skillgen.Tool{
			{
				Name:        "confluent-kafka-topic-list", // Duplicate name
				Title:       "List Kafka Topics",
				Description: "Lists all Kafka topics in the cluster",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
			{
				Name:        "confluent-kafka-topic-list", // Duplicate name
				Title:       "List Kafka Topics Again",
				Description: "Another list command",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
		},
	}

	err := validateManifestData(manifest)
	if err == nil {
		t.Error("validateManifestData() with duplicate tool names should return error, got nil")
	}
}

func TestValidateManifestData_MissingToolFields(t *testing.T) {
	testCases := []struct {
		name     string
		tool     skillgen.Tool
		errorMsg string
	}{
		{
			name: "missing name",
			tool: skillgen.Tool{
				Name:        "", // Missing name
				Title:       "Title",
				Description: "Description",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
			errorMsg: "tool missing name",
		},
		{
			name: "missing description",
			tool: skillgen.Tool{
				Name:        "confluent-kafka-topic-list",
				Title:       "Title",
				Description: "", // Missing description
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
				},
			},
			errorMsg: "tool missing description",
		},
		{
			name: "nil input schema",
			tool: skillgen.Tool{
				Name:        "confluent-kafka-topic-list",
				Title:       "Title",
				Description: "Description",
				InputSchema: skillgen.InputSchema{}, // Empty input schema (Type should be set)
			},
			errorMsg: "tool missing input_schema",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := &skillgen.SkillManifest{
				Metadata: skillgen.ManifestMetadata{
					CLIVersion:       "v1.0.0",
					GeneratedAt:      "2026-03-12T00:00:00Z",
					GeneratorVersion: "1.0.0",
					SkillCount:       1,
					CommandCount:     1,
				},
				Tools: []skillgen.Tool{tc.tool},
			}

			err := validateManifestData(manifest)
			if err == nil {
				t.Errorf("validateManifestData() with %s should return error, got nil", tc.errorMsg)
			}
		})
	}
}

func TestValidateManifestData_EmptyTools(t *testing.T) {
	manifest := &skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "v1.0.0",
			GeneratedAt:      "2026-03-12T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       0,
			CommandCount:     0,
		},
		Tools: []skillgen.Tool{}, // Empty tools array
	}

	err := validateManifestData(manifest)
	if err == nil {
		t.Error("validateManifestData() with empty tools array should return error, got nil")
	}
}
