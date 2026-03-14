package skillgen

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIRMarshaling(t *testing.T) {
	t.Run("CommandIR marshals to JSON with snake_case", func(t *testing.T) {
		cmd := CommandIR{
			CommandPath:   "confluent kafka topic list",
			Short:         "List Kafka topics",
			Long:          "List all Kafka topics in the cluster",
			Example:       "confluent kafka topic list --cluster lkc-123456",
			Flags: []FlagIR{
				{
					Name:     "cluster",
					Type:     "string",
					Required: true,
					Default:  "",
					Usage:    "Kafka cluster ID",
				},
			},
			Annotations: map[string]string{
				"run_requirement": "cloud_login",
			},
			Operation:     "list",
			Resource:      "kafka-topic",
			Mode:          "cloud",
		}

		data, err := json.Marshal(cmd)
		require.NoError(t, err)

		// Verify snake_case field names
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "command_path")
		assert.Contains(t, result, "short")
		assert.Contains(t, result, "long")
		assert.Contains(t, result, "example")
		assert.Contains(t, result, "flags")
		assert.Contains(t, result, "annotations")
		assert.Contains(t, result, "operation")
		assert.Contains(t, result, "resource")
		assert.Contains(t, result, "mode")

		assert.Equal(t, "confluent kafka topic list", result["command_path"])
		assert.Equal(t, "list", result["operation"])
		assert.Equal(t, "kafka-topic", result["resource"])
	})

	t.Run("FlagIR marshals to JSON with snake_case", func(t *testing.T) {
		flag := FlagIR{
			Name:     "cluster",
			Type:     "string",
			Required: true,
			Default:  "default-cluster",
			Usage:    "Kafka cluster ID",
		}

		data, err := json.Marshal(flag)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "name")
		assert.Contains(t, result, "type")
		assert.Contains(t, result, "required")
		assert.Contains(t, result, "default")
		assert.Contains(t, result, "usage")

		assert.Equal(t, "cluster", result["name"])
		assert.Equal(t, true, result["required"])
	})

	t.Run("Metadata marshals to JSON with snake_case", func(t *testing.T) {
		meta := Metadata{
			CLIVersion:  "3.0.0",
			GeneratedAt: time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(meta)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "cli_version")
		assert.Contains(t, result, "generated_at")

		assert.Equal(t, "3.0.0", result["cli_version"])
	})

	t.Run("IR root struct marshals complete structure", func(t *testing.T) {
		ir := IR{
			Metadata: Metadata{
				CLIVersion:  "3.0.0",
				GeneratedAt: time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
			},
			Commands: []CommandIR{
				{
					CommandPath: "confluent kafka topic list",
					Short:       "List Kafka topics",
					Operation:   "list",
					Resource:    "kafka-topic",
					Mode:        "cloud",
				},
				{
					CommandPath: "confluent iam user create",
					Short:       "Create IAM user",
					Operation:   "create",
					Resource:    "iam-user",
					Mode:        "cloud",
				},
			},
		}

		data, err := json.Marshal(ir)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "metadata")
		assert.Contains(t, result, "commands")

		commands := result["commands"].([]interface{})
		assert.Len(t, commands, 2)
	})

	t.Run("Round-trip marshal/unmarshal preserves data", func(t *testing.T) {
		original := IR{
			Metadata: Metadata{
				CLIVersion:  "3.0.0",
				GeneratedAt: time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
			},
			Commands: []CommandIR{
				{
					CommandPath:   "confluent kafka topic list",
					Short:         "List Kafka topics",
					Long:          "List all Kafka topics in the cluster",
					Example:       "confluent kafka topic list --cluster lkc-123456",
					Flags: []FlagIR{
						{
							Name:     "cluster",
							Type:     "string",
							Required: true,
							Default:  "",
							Usage:    "Kafka cluster ID",
						},
					},
					Annotations: map[string]string{
						"run_requirement": "cloud_login",
					},
					Operation:     "list",
					Resource:      "kafka-topic",
					Mode:          "cloud",
				},
			},
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var restored IR
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		// Verify data is preserved
		assert.Equal(t, original.Metadata.CLIVersion, restored.Metadata.CLIVersion)
		assert.Equal(t, original.Metadata.GeneratedAt.Unix(), restored.Metadata.GeneratedAt.Unix())
		assert.Len(t, restored.Commands, 1)
		assert.Equal(t, original.Commands[0].CommandPath, restored.Commands[0].CommandPath)
		assert.Equal(t, original.Commands[0].Operation, restored.Commands[0].Operation)
		assert.Equal(t, original.Commands[0].Resource, restored.Commands[0].Resource)
		assert.Len(t, restored.Commands[0].Flags, 1)
		assert.Equal(t, original.Commands[0].Flags[0].Name, restored.Commands[0].Flags[0].Name)
		assert.Equal(t, original.Commands[0].Flags[0].Required, restored.Commands[0].Flags[0].Required)
	})
}
