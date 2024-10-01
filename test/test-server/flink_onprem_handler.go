package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
			applicationPage := map[string]any{
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
				Metadata: map[string]any{
					"name": "state-machine-example",
				},
				Spec: map[string]any{
					"flinkEnvironment": "default",
					"image":            "confluentinc/cp-flink:1.19.1-cp1",
					"flinkVersion":     "v1_19",
					"flinkConfiguration": map[string]any{
						"taskmanager.numberOfTaskSlots":       "8",
						"metrics.reporter.prom.factory.class": "org.apache.flink.metrics.prometheus.PrometheusReporterFactory",
						"metrics.reporter.prom.port":          "9249-9250",
					},
					"serviceAccount": "flink",
					"jobManager": map[string]any{
						"resource": map[string]any{
							"memory": "1048m",
							"cpu":    1,
						},
					},
					"taskManager": map[string]any{
						"resource": map[string]any{
							"memory": "1048m",
							"cpu":    1,
						},
					},
					"job": map[string]any{
						"jarURI":      "local:///opt/flink/examples/streaming/StateMachineExample.jar",
						"state":       "running",
						"parallelism": 3,
						"upgradeMode": "stateless",
					},
				},
				Status: map[string]any{
					"jobStatus": map[string]any{
						"jobName":    "State machine job",
						"jobId":      "dcabb1ad6c40495bc2d7fa7a0097c5aa",
						"state":      "RECONCILING",
						"startTime":  "1726640263746",
						"updateTime": "1726640280561",
						"savepointInfo": map[string]any{
							"lastSavepoint":                  nil,
							"triggerId":                      nil,
							"triggerTimestamp":               nil,
							"triggerType":                    nil,
							"formatType":                     nil,
							"savepointHistory":               []any{},
							"lastPeriodicSavepointTimestamp": 0,
						},
						"checkpointInfo": map[string]any{
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
					"clusterInfo": map[string]any{
						"flink-revision": "89d0b8f @ 2024-06-22T13:19:31+02:00",
						"flink-version":  "1.19.1-cp1",
						"total-cpu":      "3.0",
						"total-memory":   "3296722944",
					},
					"jobManagerDeploymentStatus": "DEPLOYING",
					"reconciliationStatus": map[string]any{
						"reconciliationTimestamp": 1726640346899,
						"lastReconciledSpec":      "",
						"lastStableSpec":          "",
						"state":                   "DEPLOYED",
					},
					"taskManager": map[string]any{
						"labelSelector": "component=taskmanager,app=basic-example",
						"replicas":      1,
					},
				},
			},
		}
		// Return empty list of applications for pages other than 0
		applicationPage := map[string]any{
			"items": []cmfsdk.Application{},
		}
		if page == "0" {
			applicationPage = map[string]any{
				"items": items,
			}
		}
		err := json.NewEncoder(w).Encode(applicationPage)
		require.NoError(t, err)
	}
}

// Handler for "cmf/api/v1/environments/{environment}/applications/{application}"
func handleCmfApplication(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			vars := mux.Vars(r)
			if vars["environment"] == "non-existent" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}
			if vars["application"] == "non-existent" {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodPatch:
			http.Error(w, "Not implemented", http.StatusNotImplemented)
		case http.MethodGet:
			vars := mux.Vars(r)
			if vars["application"] == "non-existent" {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}
			// Just write the status code for now, more details will be added later.
			w.WriteHeader(http.StatusOK)
		}
	}
}

// Handler for GET "cmf/api/v1/environments"
func handleCmfEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		environments := []cmfsdk.Environment{
			{
				Name:        "default",
				CreatedTime: time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
				UpdatedTime: time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
			},
			{
				Name:        "etl-team",
				CreatedTime: time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
				UpdatedTime: time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
			},
		}
		// Return empty list of applications for pages other than 0
		environmentPage := map[string]interface{}{
			"items": []cmfsdk.Environment{},
		}
		if page == "0" {
			environmentPage = map[string]interface{}{
				"items": environments,
			}
		}
		err := json.NewEncoder(w).Encode(environmentPage)
		require.NoError(t, err)
	}
}

// Handler for "cmf/api/v1/environments/{environment}"
func handleCmfEnvironment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			vars := mux.Vars(r)
			environmentName := vars["environment"]
			if environmentName == "non-existent" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodPatch:
			http.Error(w, "Not implemented", http.StatusNotImplemented)
		case http.MethodGet:
			vars := mux.Vars(r)
			environmentName := vars["environment"]
			if environmentName == "non-existent" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}
			// Just write the status code for now, more details will be added later.
			w.WriteHeader(http.StatusOK)
		}
	}
}
