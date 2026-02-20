//go:build live_test && (all || kafka)

package live

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *CLILiveTestSuite) TestKafkaTopicCRUDLive() {
	t := s.T()
	t.Parallel()

	// This test requires a pre-existing cluster to avoid the ~5 min provisioning wait.
	// Skip if env vars are not set (allows local dev without persistent infra).
	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	clusterID := os.Getenv("KAFKA_STANDARD_AWS_CLUSTER_ID")
	if envID == "" || clusterID == "" {
		t.Skip("Skipping kafka topic test: LIVE_TEST_ENVIRONMENT_ID and KAFKA_STANDARD_AWS_CLUSTER_ID must be set")
	}

	state := s.setupTestContext(t)

	topicName := uniqueName("topic")

	// Best-effort cleanup: delete topic if it still exists
	s.registerCleanup(t, "kafka topic delete {{.topic_name}} --cluster "+clusterID+" --environment "+envID+" --force", state)
	state.Set("topic_name", topicName)

	steps := []CLILiveTest{
		{
			Name: "Use environment",
			Args: "environment use " + envID,
		},
		{
			Name: "Use kafka cluster",
			Args: "kafka cluster use " + clusterID,
		},
		{
			Name: "Create topic",
			Args: "kafka topic create " + topicName + " --partitions 3",
		},
		{
			Name:     "List topics",
			Args:     "kafka topic list",
			Contains: []string{topicName},
		},
		{
			Name: "Describe topic",
			Args: "kafka topic describe " + topicName + " -o json",
			JSONFields: map[string]string{
				"name": topicName,
			},
			WantFunc: func(t *testing.T, output string, state *LiveTestState) {
				t.Helper()
				partitions := extractJSONField(t, output, "partition_count")
				require.Equal(t, "3", partitions, "topic should have 3 partitions")
			},
		},
		{
			Name: "Update topic config",
			Args: "kafka topic update " + topicName + " --config retention.ms=86400000",
		},
		{
			Name: "Describe updated topic config",
			Args: "kafka topic configuration list " + topicName + " -o json",
			WantFunc: func(t *testing.T, output string, _ *LiveTestState) {
				t.Helper()
				require.Contains(t, output, "retention.ms", "output should contain retention.ms config")
				require.Contains(t, output, "86400000", "retention.ms should be 86400000")
			},
		},
		{
			Name: "Delete topic",
			Args: "kafka topic delete " + topicName + " --force",
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
