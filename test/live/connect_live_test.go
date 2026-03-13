//go:build live_test && (all || connect)

package live

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func (s *CLILiveTestSuite) TestConnectClusterCRUDLive() {
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

	connectorName := uniqueName("datagen")
	topicName := uniqueName("connect-topic")

	// Create connector config file for DatagenSource (managed connector)
	configContent := fmt.Sprintf(`{
  "connector.class": "DatagenSource",
  "name": "%s",
  "kafka.auth.mode": "KAFKA_API_KEY",
  "kafka.api.key": "%s",
  "kafka.api.secret": "%s",
  "kafka.topic": "%s",
  "output.data.format": "JSON",
  "quickstart": "ORDERS",
  "tasks.max": "1"
}`, connectorName, apiKey, apiSecret, topicName)

	configDir, err := os.MkdirTemp("", "cli-live-connect-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(configDir) })

	configFile := filepath.Join(configDir, "connector-config.json")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Register cleanups in LIFO order: topic last, connector first
	s.registerCleanup(t, "kafka topic delete "+topicName+" --cluster "+clusterID+" --environment "+envID+" --force", state)
	s.registerCleanup(t, "connect cluster delete {{.connector_id}} --cluster "+clusterID+" --environment "+envID+" --force", state)

	// Phase 1: Setup and create connector
	setupSteps := []CLILiveTest{
		{
			Name: "Use environment",
			Args: "environment use " + envID,
		},
		{
			Name: "Use kafka cluster",
			Args: "kafka cluster use " + clusterID,
		},
		{
			Name: "Create topic for connector",
			Args: "kafka topic create " + topicName + " --partitions 1",
		},
		{
			Name:      "Create connector",
			Args:      "connect cluster create --config-file " + configFile + " -o json",
			CaptureID: "connector_id",
			WantFunc: func(t *testing.T, output string, state *LiveTestState) {
				t.Helper()
				// Also capture the connector name for list verification
				name := extractJSONField(t, output, "name")
				if name != "" {
					state.Set("connector_name", name)
				}
				t.Logf("Connector create output: %s", output)
			},
		},
	}

	for _, step := range setupSteps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}

	// Phase 2: Wait for connector to reach a usable state
	t.Run("Wait for connector provisioned", func(t *testing.T) {
		s.waitForCondition(t,
			"connect cluster describe {{.connector_id}} -o json",
			state,
			func(output string) bool {
				return strings.Contains(output, "RUNNING") || strings.Contains(output, "PROVISIONING")
			},
			15*time.Second,
			5*time.Minute,
		)
	})

	// Phase 3: CRUD and lifecycle operations
	crudSteps := []CLILiveTest{
		{
			Name:     "List connectors",
			Args:     "connect cluster list",
			Contains: []string{connectorName},
			Retries:  3,
		},
		{
			Name:         "Describe connector",
			Args:         "connect cluster describe {{.connector_id}}",
			UseStateVars: true,
			Contains:     []string{connectorName},
		},
		{
			Name:         "Describe connector offsets",
			Args:         "connect offset describe {{.connector_id}} --cluster " + clusterID + " --environment " + envID,
			UseStateVars: true,
			Retries:      3,
		},
		{
			Name:         "View connector logs",
			Args:         "connect event describe {{.connector_id}} --cluster " + clusterID + " --environment " + envID,
			UseStateVars: true,
			Retries:      3,
		},
		{
			Name:          "Pause connector",
			Args:          "connect cluster pause {{.connector_id}}",
			UseStateVars:  true,
			Retries:       5,
			RetryInterval: 15 * time.Second,
		},
		{
			Name:          "Resume connector",
			Args:          "connect cluster resume {{.connector_id}}",
			UseStateVars:  true,
			Retries:       5,
			RetryInterval: 15 * time.Second,
		},
		{
			Name:         "Delete connector",
			Args:         "connect cluster delete {{.connector_id}} --force",
			UseStateVars: true,
		},
	}

	for _, step := range crudSteps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
