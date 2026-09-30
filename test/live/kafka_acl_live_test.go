//go:build live_test && (all || kafka)

package live

import (
	"os"
	"testing"
)

func (s *CLILiveTestSuite) TestKafkaACLCRUDLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	clusterID := os.Getenv("KAFKA_STANDARD_AWS_CLUSTER_ID")
	if envID == "" || clusterID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID and KAFKA_STANDARD_AWS_CLUSTER_ID must be set")
	}

	state := s.setupTestContext(t)

	saName := uniqueName("acl-sa")
	topicName := uniqueName("acl-topic")

	// Cleanup in LIFO order: delete ACL, then service account
	s.registerCleanup(t, "iam service-account delete {{.sa_id}} --force", state)
	s.registerCleanup(t, "kafka acl delete --allow --service-account {{.sa_id}} --operations read,describe --topic "+topicName+" --cluster "+clusterID+" --environment "+envID+" --force", state)

	steps := []CLILiveTest{
		{
			Name:            "Create service account for ACL test",
			Args:            `iam service-account create ` + saName + ` --description "SA for ACL live test" -o json`,
			JSONFieldsExist: []string{"id"},
			CaptureID:       "sa_id",
		},
		{
			Name: "Use environment",
			Args: "environment use " + envID,
		},
		{
			Name: "Use kafka cluster",
			Args: "kafka cluster use " + clusterID,
		},
		{
			Name:         "Create ACL allow read+describe on topic",
			Args:         "kafka acl create --allow --service-account {{.sa_id}} --operations read,describe --topic " + topicName,
			UseStateVars: true,
			Retries:      3,
		},
		{
			Name:         "List ACLs for service account",
			Args:         "kafka acl list --service-account {{.sa_id}}",
			UseStateVars: true,
			Contains:     []string{topicName},
		},
		{
			Name:         "Delete ACL",
			Args:         "kafka acl delete --allow --service-account {{.sa_id}} --operations read,describe --topic " + topicName + " --force",
			UseStateVars: true,
		},
		{
			Name:         "Verify ACL deleted",
			Args:         "kafka acl list --service-account {{.sa_id}}",
			UseStateVars: true,
			NotContains:  []string{topicName},
			Retries:      3,
		},
		{
			Name:         "Delete service account",
			Args:         "iam service-account delete {{.sa_id}} --force",
			UseStateVars: true,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
