//go:build live_test && (all || schema_registry)

package live

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *CLILiveTestSuite) TestSchemaRegistryExtendedLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	if envID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID must be set")
	}

	state := s.setupTestContext(t)

	subjectName := uniqueName("sr-ext") + "-value"

	// Create temp schema files for compatibility validation
	schemaContent := `{"type":"record","name":"ExtTestRecord","namespace":"io.confluent.test","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"}]}`
	schemaV2Content := `{"type":"record","name":"ExtTestRecord","namespace":"io.confluent.test","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"email","type":["null","string"],"default":null}]}`

	schemaDir, err := os.MkdirTemp("", "cli-live-sr-ext-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(schemaDir) })

	schemaFile := filepath.Join(schemaDir, "test.avsc")
	require.NoError(t, os.WriteFile(schemaFile, []byte(schemaContent), 0644))
	schemaFileV2 := filepath.Join(schemaDir, "test_v2.avsc")
	require.NoError(t, os.WriteFile(schemaFileV2, []byte(schemaV2Content), 0644))

	// Cleanup: delete schema versions
	s.registerCleanup(t, "schema-registry schema delete --subject "+subjectName+" --version all --force --environment "+envID, state)

	steps := []CLILiveTest{
		{
			Name: "Use environment",
			Args: "environment use " + envID,
		},
		{
			Name:            "Describe SR cluster",
			Args:            "schema-registry cluster describe --environment " + envID + " -o json",
			JSONFieldsExist: []string{"cluster_id"},
		},
		{
			Name:            "Register schema for subject",
			Args:            "schema-registry schema create --subject " + subjectName + " --schema " + schemaFile + " --type avro --environment " + envID + " -o json",
			JSONFieldsExist: []string{"id"},
		},
		{
			Name:     "Validate schema compatibility",
			Args:     "schema-registry schema compatibility validate --subject " + subjectName + " --schema " + schemaFileV2 + " --type avro --environment " + envID,
			Contains: []string{"compatible"},
		},
		{
			Name: "Update subject compatibility",
			Args: "schema-registry subject update " + subjectName + " --compatibility FULL --environment " + envID,
		},
		{
			Name:            "Describe SR global config",
			Args:            "schema-registry configuration describe --environment " + envID + " -o json",
			JSONFieldsExist: []string{"compatibility_level"},
		},
		{
			Name: "Delete schema versions",
			Args: "schema-registry schema delete --subject " + subjectName + " --version all --force --environment " + envID,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
