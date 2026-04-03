//go:build live_test && (all || kafka)

package live

import (
	"os"
	"testing"
)

func (s *CLILiveTestSuite) TestKafkaShareGroupListLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	clusterID := os.Getenv("KAFKA_DEDICATED_AWS_CLUSTER_ID")
	if envID == "" || clusterID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID and KAFKA_DEDICATED_AWS_CLUSTER_ID must be set")
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
			Name: "List share groups",
			Args: "kafka share-group list --cluster " + clusterID + " --environment " + envID,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
