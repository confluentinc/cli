package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

// Handler for GET "cmf/api/v1/environments/{environment}/applications"
func handleCmfApplications(t *testing.T) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		environmentName := vars["environment"]
		page := r.URL.Query().Get("page")
		if environmentName == "non-existent" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}
		if environmentName == "empty-environment" {
			applicationPage := map[string]interface{}{
				"items": []cmfsdk.Application{},
			}
			err := json.NewEncoder(w).Encode(applicationPage)
			require.NoError(t, err)
			return
		}
		items := []cmfsdk.Application{
			{
				ApiVersion: "cmf.confluent.io/v1alpha1",
				Kind:       "FlinkApplication",
				Metadata: map[string]interface{}{
					"name": "state-machine-example",
				},
				Spec: map[string]interface{}{
					"flinkEnvironment": "default",
					"image":            "confluentinc/cp-flink:1.19.1-cp1",
					"flinkVersion":     "v1_19",
					"flinkConfiguration": map[string]interface{}{
						"taskmanager.numberOfTaskSlots":       "8",
						"metrics.reporter.prom.factory.class": "org.apache.flink.metrics.prometheus.PrometheusReporterFactory",
						"metrics.reporter.prom.port":          "9249-9250",
					},
					"serviceAccount": "flink",
					"jobManager": map[string]interface{}{
						"resource": map[string]interface{}{
							"memory": "1048m",
							"cpu":    1,
						},
					},
					"taskManager": map[string]interface{}{
						"resource": map[string]interface{}{
							"memory": "1048m",
							"cpu":    1,
						},
					},
					"job": map[string]interface{}{
						"jarURI":      "local:///opt/flink/examples/streaming/StateMachineExample.jar",
						"state":       "running",
						"parallelism": 3,
						"upgradeMode": "stateless",
					},
				},
				Status: map[string]interface{}{
					"jobStatus": map[string]interface{}{
						"jobName":    "State machine job",
						"jobId":      "dcabb1ad6c40495bc2d7fa7a0097c5aa",
						"state":      "RECONCILING",
						"startTime":  "1726640263746",
						"updateTime": "1726640280561",
						"savepointInfo": map[string]interface{}{
							"lastSavepoint":                  nil,
							"triggerId":                      nil,
							"triggerTimestamp":               nil,
							"triggerType":                    nil,
							"formatType":                     nil,
							"savepointHistory":               []interface{}{},
							"lastPeriodicSavepointTimestamp": 0,
						},
						"checkpointInfo": map[string]interface{}{
							"lastCheckpoint":                  nil,
							"triggerId":                       nil,
							"triggerTimestamp":                nil,
							"triggerType":                     nil,
							"formatType":                      nil,
							"lastPeriodicCheckpointTimestamp": 0,
						},
					},
					"error":              nil,
					"observedGeneration": 4,
					"lifecycleState":     "DEPLOYED",
					"clusterInfo": map[string]interface{}{
						"flink-revision": "89d0b8f @ 2024-06-22T13:19:31+02:00",
						"flink-version":  "1.19.1-cp1",
						"total-cpu":      "3.0",
						"total-memory":   "3296722944",
					},
					"jobManagerDeploymentStatus": "DEPLOYING",
					"reconciliationStatus": map[string]interface{}{
						"reconciliationTimestamp": 1726640346899,
						"lastReconciledSpec":      "",
						"lastStableSpec":          "",
						"state":                   "DEPLOYED",
					},
					"taskManager": map[string]interface{}{
						"labelSelector": "component=taskmanager,app=basic-example",
						"replicas":      1,
					},
				},
			},
		}
		// Return empty list of applications for pages other than 0
		applicationPage := map[string]interface{}{
			"items": []cmfsdk.Application{},
		}
		if page == "0" {
			applicationPage = map[string]interface{}{
				"items": items,
			}
		}
		err := json.NewEncoder(w).Encode(applicationPage)
		require.NoError(t, err)
	}
}
