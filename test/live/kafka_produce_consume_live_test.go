//go:build live_test && (all || kafka)

package live

import (
	"os"
	"testing"
)

func (s *CLILiveTestSuite) TestKafkaProduceConsumeLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	clusterID := os.Getenv("KAFKA_STANDARD_AWS_CLUSTER_ID")
	apiKey := os.Getenv("KAFKA_STANDARD_AWS_API_KEY")
	apiSecret := os.Getenv("KAFKA_STANDARD_AWS_API_SECRET")
	if envID == "" || clusterID == "" || apiKey == "" || apiSecret == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID, KAFKA_STANDARD_AWS_CLUSTER_ID, KAFKA_STANDARD_AWS_API_KEY, and KAFKA_STANDARD_AWS_API_SECRET must be set")
	}

	state := s.setupTestContext(t)

	topicName := uniqueName("produce-consume")
	state.Set("topic_name", topicName)

	s.registerCleanup(t, "kafka topic delete "+topicName+" --cluster "+clusterID+" --environment "+envID+" --force", state)

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
			Name: "Store API key",
			Args: "api-key store " + apiKey + " " + apiSecret + " --resource " + clusterID + " --force",
		},
		{
			Name: "Use API key",
			Args: "api-key use " + apiKey + " --resource " + clusterID,
		},
		{
			Name: "Create topic for produce/consume",
			Args: "kafka topic create " + topicName + " --partitions 1",
		},
		{
			Name:  "Produce message to topic",
			Args:  "kafka topic produce " + topicName,
			Input: "test-key:test-value\n",
		},
		{
			Name:     "Consume message from topic",
			Args:     "kafka topic consume " + topicName + " --from-beginning --exit",
			Contains: []string{"test-value"},
			Retries:  5,
		},
		{
			Name: "List consumer groups",
			Args: "kafka consumer group list",
		},
		{
			Name: "Create client config",
			Args: "kafka client-config create java --environment " + envID + " --cluster " + clusterID,
			Contains: []string{"bootstrap.servers"},
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
