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

func createEnvironment(name string) cmfsdk.Environment {
	return cmfsdk.Environment{
		Name:        name,
		CreatedTime: time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
		UpdatedTime: time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC),
	}
}

// There are a number of request and responses for each path depending on the test case.
// They can be uniquely distinguished by a per API-path tuple (request method, [environment name prefix, [application name]]).

func commandTypeByEnvironment(environmentName string) string {
	if strings.HasPrefix(environmentName, "create") {
		return "create"
	} else if strings.HasPrefix(environmentName, "update") {
		return "update"
	} else if strings.HasPrefix(environmentName, "list") {
		return "list"
	} else if strings.HasPrefix(environmentName, "delete") {
		return "delete"
	}
	return "unknown"
}

// Global level handlers which dispatch specific handlers as required.

// Handler for GET "cmf/api/v1/environments"
// Used by TestListFlinkEnvironments (GET, nil, nil)
// Used by TestCreateFlinkEnvironments (POST, "create-", nil)
// Used by TestUpdateFlinkEnvironments (POST, "update-", nil)
func handleCmfEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			environments := []cmfsdk.Environment{
				createEnvironment("default"),
				createEnvironment("etl-team"),
			}
			environmentPage := map[string]interface{}{
				"items": environments,
			}
			err := json.NewEncoder(w).Encode(environmentPage)
			require.NoError(t, err)
			return
		}

		if r.Method == http.MethodPost {
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var environment cmfsdk.PostEnvironment
			err = json.Unmarshal(reqBody, &environment)
			require.NoError(t, err)

			environmentName := environment.Name
			if environmentName == "create-success" || environmentName == "create-success-with-defaults" {
				outputEnvironment := createEnvironment(environmentName)
				outputEnvironment.Defaults = environment.Defaults
				err = json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			if environmentName == "create-failure" {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			if environmentName == "update-success" {
				outputEnvironment := createEnvironment("update-success")
				// This is a dummy update - only the defaults can be updated anyway.
				outputEnvironment.Defaults = environment.Defaults
				err = json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			if environmentName == "update-failure" {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			require.Fail(t, fmt.Sprintf("Unexpected environment name %s", environmentName))
		}
	}
}

// Handler for "cmf/api/v1/environments/{environment}"
// Used by TestDeleteFlinkEnvironments (DELETE, "delete-", nil)
// Used by TestCreateeFlinkEnvironments (POST, "create-", nil) for listing before create
// Used by TestUpdateFlinkEnvironments (POST, "update-", nil) for listing before update
func handleCmfEnvironment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environment := mux.Vars(r)["environment"]
		if r.Method == http.MethodDelete && commandTypeByEnvironment(environment) == "delete" {
			if environment == "delete-non-existent" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodGet && commandTypeByEnvironment(environment) == "create" {
			if environment == "create-existing" {
				outputEnvironment := createEnvironment("create-existing")
				err := json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			if environment == "create-success" || environment == "create-failure" || environment == "create-success-with-defaults" {
				http.Error(w, "", http.StatusNotFound)
				return
			}
		}

		if r.Method == http.MethodGet && commandTypeByEnvironment(environment) == "update" {
			if environment == "update-success" || environment == "update-failure" {
				outputEnvironment := createEnvironment(environment)
				err := json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			if environment == "update-non-existent" {
				http.Error(w, "", http.StatusNotFound)
				return
			}
		}

		require.Fail(t, fmt.Sprintf("Unexpected method %s or environment name %s", r.Method, environment))
	}
}

// Handler for "cmf/api/v1/environments/{environment}/applications"
// Used by TestListFlinkApplications (GET, "list-", nil)
// Used by TestCreateFlinkApplications (POST, "create-", applicationName)
// Used by TestUpdateFlinkApplications (POST, "update-", applicationName)
func handleCmfApplications(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		environment := vars["environment"]
		if r.Method == http.MethodGet && commandTypeByEnvironment(environment) == "list" {
			if environment == "list-non-existent" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}

			if environment == "list-empty-environment" {
				applicationPage := map[string]interface{}{
					"items": []cmfsdk.Application{},
				}
				err := json.NewEncoder(w).Encode(applicationPage)
				require.NoError(t, err)
				return
			}

			items := []cmfsdk.Application{createApplication("list-application-application", "list-test")}
			applicationPage := map[string]interface{}{
				"items": items,
			}
			err := json.NewEncoder(w).Encode(applicationPage)
			require.NoError(t, err)

			return
		}

		if r.Method == http.MethodPost && commandTypeByEnvironment(environment) == "create" {
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var application cmfsdk.Application
			err = json.Unmarshal(reqBody, &application)
			require.NoError(t, err)

			applicationName := application.Metadata["name"].(string)
			if applicationName == "create-non-existent-successfully" {
				// Success case - just echo the application back to the user.
				err = json.NewEncoder(w).Encode(application)
				require.NoError(t, err)
				return
			}
			if environment == "create-with-non-existent-environment" || applicationName == "create-non-existent-unsuccessfully" {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			require.Fail(t, fmt.Sprintf("Unexpected application name %s", applicationName))
		}

		if r.Method == http.MethodPost && commandTypeByEnvironment(environment) == "update" {
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var application cmfsdk.Application
			err = json.Unmarshal(reqBody, &application)
			require.NoError(t, err)

			applicationName := application.Metadata["name"].(string)
			if applicationName == "update-successful" {
				// Success case - send the application back to the user with the 'update'.
				// The 'update' is going to be spec.serviceAccount. This is just a dummy update,
				// and we don't do any actual merge logic.
				outputApplication := createApplication("update-successful", "update-test")
				err = json.NewEncoder(w).Encode(outputApplication)
				require.NoError(t, err)
				return
			}

			if applicationName == "update-failure" {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			require.Fail(t, fmt.Sprintf("Unexpected application name %s", applicationName))
		}

		require.Fail(t, fmt.Sprintf("Unexpected method %s or environment %s", r.Method, environment))
	}
}

// Handler for "cmf/api/v1/environments/{environment}/applications/{application}"
// Used by TestCreateFlinkApplications (GET, "create-", applicationName) for listing before create
// Used by TestUpdateFlinkApplications (GET, "update-", applicationName) for listing before update
// Used bt TestDeleteFlinkApplications (DELETE, "delete-", applicationName)
func handleCmfApplication(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		environment := vars["environment"]
		application := vars["application"]
		if r.Method == http.MethodGet && commandTypeByEnvironment(environment) == "create" {
			if environment == "create-with-non-existent-environment" || application == "create-non-existent-unsuccessfully" || application == "create-non-existent-successfully" {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}

			if application == "create-already-existing" {
				application := createApplication("create-already-existing", "create-test")
				err := json.NewEncoder(w).Encode(application)
				require.NoError(t, err)
				return
			}

			require.Fail(t, fmt.Sprintf("Unexpected application name %s", application))
			return
		}

		if r.Method == http.MethodGet && commandTypeByEnvironment(environment) == "update" {
			if environment == "update-with-non-existent-environment" || application == "update-non-existent" {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}

			if application == "update-successful" || application == "update-failure" {
				application := createApplication(application, "update-test")
				err := json.NewEncoder(w).Encode(application)
				require.NoError(t, err)
				return
			}

			require.Fail(t, fmt.Sprintf("Unexpected application name %s", application))
			return
		}

		if r.Method == http.MethodDelete && commandTypeByEnvironment(environment) == "delete" {
			require.Fail(t, "Not implemented.")
		}
	}
}
