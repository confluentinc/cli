//go:build live_test && (all || kafka)

package live

import (
	"os"
	"testing"
)

func (s *CLILiveTestSuite) TestKafkaQuotaCRUDLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	clusterID := os.Getenv("KAFKA_STANDARD_AWS_CLUSTER_ID")
	if envID == "" || clusterID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID and KAFKA_STANDARD_AWS_CLUSTER_ID must be set")
	}

	state := s.setupTestContext(t)

	// Register cleanup
	s.registerCleanup(t, "kafka quota delete {{.quota_id}} --cluster "+clusterID+" --environment "+envID+" --force", state)

	saName := uniqueName("quota-sa")
	s.registerCleanup(t, "iam service-account delete {{.quota_sa_id}} --force", state)

	steps := []CLILiveTest{
		{
			Name: "Use environment",
			Args: "environment use " + envID,
		},
		{
			Name: "Use kafka cluster",
			Args: "kafka cluster use " + clusterID,
		},
		// Create a service account to use as the quota principal
		{
			Name:      "Create service account for quota",
			Args:      `iam service-account create ` + saName + ` --description "SA for quota live test" -o json`,
			CaptureID: "quota_sa_id",
		},
		{
			Name:         "Create client quota",
			Args:         "kafka quota create --ingress 1048576 --egress 1048576 --principals sa:{{.quota_sa_id}} --cluster " + clusterID + " --environment " + envID + ` --description "Live test quota" -o json`,
			UseStateVars: true,
			CaptureID:    "quota_id",
			JSONFieldsExist: []string{"id"},
		},
		{
			Name:         "Describe client quota",
			Args:         "kafka quota describe {{.quota_id}} --cluster " + clusterID + " --environment " + envID + " -o json",
			UseStateVars: true,
			JSONFieldsExist: []string{"id"},
		},
		{
			Name: "List client quotas",
			Args: "kafka quota list --cluster " + clusterID + " --environment " + envID,
		},
		{
			Name:         "Update client quota",
			Args:         "kafka quota update {{.quota_id}} --ingress 2097152 --egress 2097152 --cluster " + clusterID + " --environment " + envID + ` --description "Updated live test quota"`,
			UseStateVars: true,
		},
		{
			Name:         "Delete client quota",
			Args:         "kafka quota delete {{.quota_id}} --cluster " + clusterID + " --environment " + envID + " --force",
			UseStateVars: true,
		},
		{
			Name:         "Delete service account",
			Args:         "iam service-account delete {{.quota_sa_id}} --force",
			UseStateVars: true,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
