package testserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

// Helper function to create a Flink application.
func createApplication(name string, environment string) cmfsdk.Application {
	return cmfsdk.Application{
		ApiVersion: "cmf.confluent.io/v1alpha1",
		Kind:       "FlinkApplication",
		Metadata: map[string]interface{}{
			"name": name,
		},
		Spec: map[string]interface{}{
			"flinkEnvironmentName": environment,
			"image":                "confluentinc/cp-flink:1.19.1-cp1",
			"flinkVersion":         "v1_19",
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
	}
}

// Helper function to create a Flink environment.
func createEnvironment(name string, namespace string) cmfsdk.Environment {
	return cmfsdk.Environment{
		Name:                name,
		KubernetesNamespace: namespace,
		CreatedTime:         time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
		UpdatedTime:         time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
	}
}

// There are a number of request and responses for each path depending on the test case.
// We assume the following set of existing environments and applications as already existing:
// default: default-application-1, default-application-2
// test: [empty environment]
// update-failure: update-failure-application (Only used by environment/application update failure test)
// All other environments and applications don't exist.
// In case an environment or application name has the substring "failure", the operation will fail with a 422 status code.
// There is some special handling required to make update-failure-application work, but that's pretty much the only special case.

// Global level handlers which dispatch specific handlers as required.

// Handler for "cmf/api/v1/environments"
// Used to list, create and update environments.
func handleCmfEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			page := r.URL.Query().Get("page")
			environments := []cmfsdk.Environment{
				createEnvironment("default", "default-namespace"),
				createEnvironment("test", "test-namespace"),
				createEnvironment("update-failure", "update-failure-namespace"),
			}
			environmentPage := map[string]interface{}{
				"items": []cmfsdk.Environment{},
			}
			// Only return the environments on page 0, otherwise return an empty list.
			if page == "0" {
				environmentPage = map[string]interface{}{
					"items": environments,
				}
			}
			err := json.NewEncoder(w).Encode(environmentPage)
			require.NoError(t, err)
			return
		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var environment cmfsdk.PostEnvironment
			err = json.Unmarshal(reqBody, &environment)
			require.NoError(t, err)

			if strings.Contains(environment.Name, "failure") {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			// Already existing environment: update
			if environment.Name == "default" || environment.Name == "test" {
				outputEnvironment := createEnvironment(environment.Name, environment.Name+"-namespace")
				// This is a dummy update - only the defaults can be updated anyway.
				outputEnvironment.FlinkApplicationDefaults = environment.FlinkApplicationDefaults
				err = json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			// New environment: create
			outputEnvironment := createEnvironment(environment.Name, environment.KubernetesNamespace)
			outputEnvironment.FlinkApplicationDefaults = environment.FlinkApplicationDefaults
			err = json.NewEncoder(w).Encode(outputEnvironment)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "cmf/api/v1/environments/{environment}"
// Used by create and update (to validate existence).
// Used by describe and delete.
func handleCmfEnvironment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environment := mux.Vars(r)["environment"]
		switch r.Method {
		case http.MethodGet:
			if environment == "default" || environment == "test" || environment == "update-failure" {
				outputEnvironment := createEnvironment(environment, environment+"-namespace")
				err := json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		case http.MethodDelete:
			if strings.Contains(environment, "failure") {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			if environment == "default" || environment == "test" {
				w.WriteHeader(http.StatusOK)
				return
			}

			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "cmf/api/v1/environments/{environment}/applications"
// Used by list, create and update applications.
func handleCmfApplications(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		environment := vars["environment"]
		switch r.Method {
		case http.MethodGet:
			if environment != "default" && environment != "test" && environment != "update-failure" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}

			// For the 'test' environment, return an empty list.
			// For the 'default' environment, return applications but only on page 0.
			// For the 'update-failure' environment, return the 'update-failure-application' application.
			applicationsPage := map[string]interface{}{
				"items": []cmfsdk.Application{},
			}

			page := r.URL.Query().Get("page")

			if environment == "default" && page == "0" {
				items := []cmfsdk.Application{createApplication("default-application-1", "default"), createApplication("default-application-2", "default")}
				applicationsPage = map[string]interface{}{
					"items": items,
				}
			}

			if environment == "update-failure" && page == "0" {
				items := []cmfsdk.Application{createApplication("update-failure-application", "update-failure")}
				applicationsPage = map[string]interface{}{
					"items": items,
				}
			}

			err := json.NewEncoder(w).Encode(applicationsPage)
			require.NoError(t, err)
			return

		case http.MethodPost:
			if environment != "default" && environment != "test" && environment != "update-failure" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}

			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var application cmfsdk.Application
			err = json.Unmarshal(reqBody, &application)
			require.NoError(t, err)

			applicationName := application.Metadata["name"].(string)
			if strings.Contains(applicationName, "failure") {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			// If application already exists, return the application with the 'update'.
			if applicationName == "default-application-1" || applicationName == "default-application-2" {
				// The 'update' is going to be spec.serviceAccount. This is just a dummy update,
				// and we don't do any actual merge logic.
				outputApplication := createApplication(applicationName, environment)
				outputApplication.Spec["serviceAccount"] = application.Spec["serviceAccount"]
				err = json.NewEncoder(w).Encode(outputApplication)
				require.NoError(t, err)
				return
			}

			// If application does not exist, return the application newly 'created'.
			err = json.NewEncoder(w).Encode(application)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "cmf/api/v1/environments/{environment}/applications/{application}"
// Used by create and update (to validate existence).
// Used by describe and delete applications.
func handleCmfApplication(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		environment := vars["environment"]
		application := vars["application"]

		if environment != "default" && environment != "test" && environment != "update-failure" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			// In case the application actually exists, let the handler return the application.
			if (application == "default-application-1" || application == "default-application-2") && environment == "default" {
				outputApplication := createApplication(application, environment)
				err := json.NewEncoder(w).Encode(outputApplication)
				require.NoError(t, err)
				return
			}

			if application == "update-failure-application" && environment == "update-failure" {
				outputApplication := createApplication(application, environment)
				err := json.NewEncoder(w).Encode(outputApplication)
				require.NoError(t, err)
				return
			}

			http.Error(w, "Application not found", http.StatusNotFound)
		case http.MethodDelete:
			if strings.Contains(application, "failure") {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			if (application == "default-application-1" || application == "default-application-2") && environment == "default" {
				w.WriteHeader(http.StatusOK)
				return
			}

			http.Error(w, "Application not found", http.StatusNotFound)
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}
