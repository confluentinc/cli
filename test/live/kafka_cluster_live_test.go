//go:build live_test && (all || kafka)

package live

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func (s *CLILiveTestSuite) TestKafkaClusterCRUDLive() {
	t := s.T()
	t.Parallel()

	for _, variant := range liveTestVariants() {
		variant := variant
		t.Run(variant.String(), func(t *testing.T) {
			state := s.setupTestContext(t)

			cloud := variant.Cloud
			region := variant.Region
			clusterType := variant.ClusterType
			clusterName := uniqueName("cluster")
			updatedName := clusterName + "-updated"
			envName := uniqueName("kafka-env")

			// Cleanup in LIFO order: delete cluster first, then environment
			s.registerCleanup(t, "environment delete {{.env_id}} --force", state)
			s.registerCleanup(t, "kafka cluster delete {{.cluster_id}} --force --environment {{.env_id}}", state)

			// Phase 1: Create environment and cluster
			createSteps := []CLILiveTest{
				{
					Name:            "Create environment for kafka test",
					Args:            "environment create " + envName + " -o json",
					JSONFieldsExist: []string{"id"},
					CaptureID:       "env_id",
				},
				{
					Name:         "Use environment",
					Args:         "environment use {{.env_id}}",
					UseStateVars: true,
				},
				{
					Name:         "Create kafka cluster",
					Args:         fmt.Sprintf("kafka cluster create %s --cloud %s --region %s --type %s -o json", clusterName, cloud, region, clusterType),
					UseStateVars: true,
					CaptureID:    "cluster_id",
					JSONFields: map[string]string{
						"name":   clusterName,
						"cloud":  cloud,
						"region": region,
					},
					JSONFieldsExist: []string{"id", "status"},
				},
			}

			for _, step := range createSteps {
				t.Run(step.Name, func(t *testing.T) {
					s.runLiveCommand(t, step, state)
				})
			}

			// Phase 2: Wait for cluster to be provisioned (~5 min for basic)
			t.Run("Wait for cluster provisioned", func(t *testing.T) {
				s.waitForCondition(t,
					"kafka cluster describe {{.cluster_id}} --environment {{.env_id}} -o json",
					state,
					func(output string) bool {
						status := extractJSONField(t, output, "status")
						return strings.EqualFold(status, "UP")
					},
					30*time.Second,
					10*time.Minute,
				)
			})

			// Phase 3: CRUD operations on the provisioned cluster
			crudSteps := []CLILiveTest{
				{
					Name:         "List kafka clusters",
					Args:         "kafka cluster list --environment {{.env_id}}",
					UseStateVars: true,
					Contains:     []string{clusterName},
				},
				{
					Name:         "Update cluster name",
					Args:         "kafka cluster update {{.cluster_id}} --name " + updatedName + " --environment {{.env_id}}",
					UseStateVars: true,
				},
				{
					Name:         "Describe updated cluster",
					Args:         "kafka cluster describe {{.cluster_id}} --environment {{.env_id}} -o json",
					UseStateVars: true,
					JSONFields: map[string]string{
						"name":   updatedName,
						"cloud":  cloud,
						"region": region,
					},
				},
				{
					Name:         "Delete cluster",
					Args:         "kafka cluster delete {{.cluster_id}} --force --environment {{.env_id}}",
					UseStateVars: true,
				},
			}

			for _, step := range crudSteps {
				t.Run(step.Name, func(t *testing.T) {
					s.runLiveCommand(t, step, state)
				})
			}
		})
	}
}
