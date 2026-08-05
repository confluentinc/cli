//go:build live_test && (all || kafka)

package live

import (
	"os"
	"testing"
)

func (s *CLILiveTestSuite) TestKafkaConsumerGroupListLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	clusterID := os.Getenv("KAFKA_STANDARD_AWS_CLUSTER_ID")
	if envID == "" || clusterID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID and KAFKA_STANDARD_AWS_CLUSTER_ID must be set")
	}

	state := s.setupTestContext(t)

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
			Name: "List consumer groups",
			Args: "kafka consumer group list",
			// Verify command succeeds — there may or may not be active consumer groups
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
