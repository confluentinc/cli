//go:build live_test && (all || schema_registry)

package live

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *CLILiveTestSuite) TestSchemaRegistrySchemaCRUDLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	if envID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID must be set")
	}

	state := s.setupTestContext(t)

	subjectName := uniqueName("sr-schema") + "-value"

	// Create temp avro schema files
	schemaContent := `{"type":"record","name":"TestRecord","namespace":"io.confluent.test","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"}]}`

	schemaDir, err := os.MkdirTemp("", "cli-live-schema-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(schemaDir) })

	schemaFile := filepath.Join(schemaDir, "test.avsc")
	require.NoError(t, os.WriteFile(schemaFile, []byte(schemaContent), 0644))

	// Create a backward-compatible v2 schema (adds optional field)
	schemaV2Content := `{"type":"record","name":"TestRecord","namespace":"io.confluent.test","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"email","type":["null","string"],"default":null}]}`
	schemaFileV2 := filepath.Join(schemaDir, "test_v2.avsc")
	require.NoError(t, os.WriteFile(schemaFileV2, []byte(schemaV2Content), 0644))

	// Cleanup: soft-delete all schema versions under this subject
	s.registerCleanup(t, "schema-registry schema delete --subject "+subjectName+" --version all --force --environment "+envID, state)

	steps := []CLILiveTest{
		{
			Name: "Use environment",
			Args: "environment use " + envID,
		},
		{
			Name:            "Register schema v1",
			Args:            "schema-registry schema create --subject " + subjectName + " --schema " + schemaFile + " --type avro --environment " + envID + " -o json",
			JSONFieldsExist: []string{"id"},
			WantFunc: func(t *testing.T, output string, state *LiveTestState) {
				t.Helper()
				id := extractJSONField(t, output, "id")
				require.NotEmpty(t, id, "failed to extract schema id")
				state.Set("schema_id", id)
				t.Logf("Captured schema_id = %s", id)
			},
		},
		{
			Name:            "Register schema v2",
			Args:            "schema-registry schema create --subject " + subjectName + " --schema " + schemaFileV2 + " --type avro --environment " + envID + " -o json",
			JSONFieldsExist: []string{"id"},
		},
		{
			Name:         "Describe schema by ID",
			Args:         "schema-registry schema describe {{.schema_id}} --environment " + envID,
			UseStateVars: true,
			Contains:     []string{"TestRecord"},
		},
		{
			Name:     "Describe schema by subject and version",
			Args:     "schema-registry schema describe --subject " + subjectName + " --version latest --environment " + envID,
			Contains: []string{"TestRecord", "email"},
		},
		{
			Name:     "List schemas for subject",
			Args:     "schema-registry schema list --subject-prefix " + subjectName + " --environment " + envID,
			Contains: []string{subjectName},
		},
		{
			Name:     "List subjects",
			Args:     "schema-registry subject list --environment " + envID,
			Contains: []string{subjectName},
			Retries:  3,
		},
		{
			Name:     "Describe subject versions",
			Args:     "schema-registry subject describe " + subjectName + " --environment " + envID,
			Contains: []string{"1"},
		},
		{
			Name:            "Describe global SR configuration",
			Args:            "schema-registry configuration describe --environment " + envID + " -o json",
			JSONFieldsExist: []string{"compatibility_level"},
		},
		{
			Name: "Delete all schema versions (soft delete)",
			Args: "schema-registry schema delete --subject " + subjectName + " --version all --force --environment " + envID,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
