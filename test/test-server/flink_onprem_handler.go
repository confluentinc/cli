package testserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

// Helper function to create a Flink application.
func createApplication(name string) cmfsdk.FlinkApplication {
	status := map[string]interface{}{
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
	}

	return cmfsdk.FlinkApplication{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "FlinkApplication",
		Metadata: map[string]interface{}{
			"name": name,
		},
		Spec: map[string]interface{}{
			"image":        "confluentinc/cp-flink:1.19.1-cp1",
			"flinkVersion": "v1_19",
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
		Status: &status,
	}
}

// Helper function to create a Flink environment.
func createEnvironment(name string, namespace string) cmfsdk.Environment {
	createdTime := time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2024, time.September, 10, 23, 0, 0, 0, time.UTC)

	return cmfsdk.Environment{
		Name:                name,
		KubernetesNamespace: namespace,
		CreatedTime:         &createdTime,
		UpdatedTime:         &updatedTime,
	}
}

// Helper function to create a Flink environment with default application, compute pool and statement.
func createEnvironmentWithDefaults(name string, namespace string) cmfsdk.Environment {
	createdTime := time.Date(2025, time.September, 25, 12, 29, 0, 0, time.UTC)
	updatedTime := time.Date(2025, time.September, 25, 12, 29, 0, 0, time.UTC)

	applicationDefaults := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"fmc.platform.confluent.io/intra-cluster-ssl": "false",
			},
		},
		"spec": map[string]interface{}{
			"flinkConfiguration": map[string]interface{}{
				"taskmanager.numberOfTaskSlots": "8",
			},
		},
	}

	computePoolDefaults := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "test-pool",
		},
		"spec": map[string]interface{}{
			"type": "DEDICATED",
		},
	}

	detachedConfig := map[string]string{"key1": "value1"}
	interactiveConfig := map[string]string{"key2": "value2"}

	statementDefaults := cmfsdk.AllStatementDefaults1{
		Detached: &cmfsdk.StatementDefaults{
			FlinkConfiguration: &detachedConfig,
		},
		Interactive: &cmfsdk.StatementDefaults{
			FlinkConfiguration: &interactiveConfig,
		},
	}

	return cmfsdk.Environment{
		Name:                     name,
		KubernetesNamespace:      namespace,
		CreatedTime:              &createdTime,
		UpdatedTime:              &updatedTime,
		FlinkApplicationDefaults: &applicationDefaults,
		ComputePoolDefaults:      &computePoolDefaults,
		StatementDefaults:        &statementDefaults,
	}
}

func createComputePool(poolName, phase string) cmfsdk.ComputePool {
	timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()

	status := cmfsdk.ComputePoolStatus{
		Phase: phase,
	}

	return cmfsdk.ComputePool{
		Metadata: cmfsdk.ComputePoolMetadata{
			Name:              poolName,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.ComputePoolSpec{
			Type: "DEDICATED",
		},
		Status: &status,
	}
}

func createKafkaCatalog(catName string) cmfsdk.KafkaCatalog {
	timeStamp := time.Date(2025, time.August, 5, 12, 00, 0, 0, time.UTC).String()
	return cmfsdk.KafkaCatalog{
		Metadata: cmfsdk.CatalogMetadata{
			Name:              catName,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.KafkaCatalogSpec{
			KafkaClusters: []cmfsdk.KafkaCatalogSpecKafkaClusters{
				{
					DatabaseName: "test-database",
				},
				{
					DatabaseName: "test-database-2",
				},
			},
		},
	}
}

func createFlinkStatement(stmtName string, stopped bool, parallelism int32) cmfsdk.Statement {
	timeStamp := time.Date(2025, time.August, 5, 12, 00, 0, 0, time.UTC).String()
	status := cmfsdk.StatementStatus{
		Phase:  "PENDING",
		Detail: cmfsdk.PtrString("Statement is pending execution."),
		Traits: &cmfsdk.StatementTraits{
			SqlKind:      cmfsdk.PtrString("SELECT"),
			IsAppendOnly: cmfsdk.PtrBool(false),
			IsBounded:    cmfsdk.PtrBool(false),
		},
	}

	return cmfsdk.Statement{
		Metadata: cmfsdk.StatementMetadata{
			Name:              stmtName,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.StatementSpec{
			Statement:       "SELECT * FROM test_table",
			ComputePoolName: "test-pool",
			Parallelism:     cmfsdk.PtrInt32(parallelism),
			Stopped:         cmfsdk.PtrBool(stopped),
		},
		Status: &status,
	}
}

func createStatementExceptionData() []cmfsdk.StatementException {
	return []cmfsdk.StatementException{
		{
			Name:      "statement-exception-1",
			Message:   "This is a test exception message.",
			Timestamp: "2025-08-05T12:00:00Z",
		},
		{
			Name:      "statement-exception-2",
			Message:   "This is another test exception message.",
			Timestamp: "2025-08-05T12:01:00Z",
		},
	}
}

// Helper function to check that the login type is either empty or onprem, and if it's onprem,
// that the headers are correct.
func handleLoginType(t *testing.T, r *http.Request) {
	loginType := os.Getenv("LOGIN_TYPE")

	// Depending on the login type, we need to check if the headers are correct
	if loginType == "" {
		return
	}
	if loginType != "onprem" {
		require.Fail(t, "LOGIN_TYPE besides onprem and not-logged in are not allowed - the test has a bug.")
		return
	}

	authValue := r.Header.Get("Authorization")
	require.NotEqual(t, "", authValue)
	require.True(t, strings.HasPrefix(authValue, "Bearer "))
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
		handleLoginType(t, r)

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
				outputEnvironment.ComputePoolDefaults = environment.ComputePoolDefaults
				outputEnvironment.StatementDefaults = environment.StatementDefaults
				err = json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			// New environment: create
			outputEnvironment := createEnvironment(environment.Name, environment.GetKubernetesNamespace())
			outputEnvironment.FlinkApplicationDefaults = environment.FlinkApplicationDefaults
			outputEnvironment.ComputePoolDefaults = environment.ComputePoolDefaults
			outputEnvironment.StatementDefaults = environment.StatementDefaults
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
		handleLoginType(t, r)

		environment := mux.Vars(r)["environment"]
		switch r.Method {
		case http.MethodGet:
			if environment == "default" || environment == "test" || environment == "update-failure" {
				outputEnvironment := createEnvironment(environment, environment+"-namespace")
				err := json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			if environment == "defaults-all" {
				outputEnvironment := createEnvironmentWithDefaults(environment, environment+"-namespace")
				err := json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			if strings.Contains(environment, "failure") {
				http.Error(w, "", http.StatusUnprocessableEntity)
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
		handleLoginType(t, r)

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
				"items": []cmfsdk.FlinkApplication{},
			}

			page := r.URL.Query().Get("page")

			if environment == "default" && page == "0" {
				items := []cmfsdk.FlinkApplication{createApplication("default-application-1"), createApplication("default-application-2")}
				applicationsPage = map[string]interface{}{
					"items": items,
				}
			}

			if environment == "update-failure" && page == "0" {
				items := []cmfsdk.FlinkApplication{createApplication("update-failure-application")}
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
			var application cmfsdk.FlinkApplication
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
				outputApplication := createApplication(applicationName)
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
		handleLoginType(t, r)

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
				outputApplication := createApplication(application)
				err := json.NewEncoder(w).Encode(outputApplication)
				require.NoError(t, err)
				return
			}

			if application == "update-failure-application" && environment == "update-failure" {
				outputApplication := createApplication(application)
				err := json.NewEncoder(w).Encode(outputApplication)
				require.NoError(t, err)
				return
			}

			if strings.Contains(application, "failure") {
				http.Error(w, "", http.StatusUnprocessableEntity)
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

// Handler for "cmf/api/v1/environments/{envName}/compute-pools"
// Used by list, create compute pools, no update compute pools.
func handleCmfComputePools(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["environment"]

		if environment == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			computePool1 := createComputePool("test-pool1", "RUNNING")
			computePool2 := createComputePool("test-pool2", "PENDING")
			computePool3 := createComputePool("test-pool3", "COMPLETE")

			computePools := []cmfsdk.ComputePool{computePool1, computePool2, computePool3}
			computePoolsPage := cmfsdk.ComputePoolsPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				computePoolsPage.SetItems(computePools)
			}

			err := json.NewEncoder(w).Encode(computePoolsPage)
			require.NoError(t, err)
			return

		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var computePool cmfsdk.ComputePool
			err = json.Unmarshal(reqBody, &computePool)
			require.NoError(t, err)

			poolName := computePool.GetMetadata().Name

			if poolName == "invalid-pool" {
				http.Error(w, "The compute pool object from resource file is invalid", http.StatusUnprocessableEntity)
				return
			}
			if poolName == "existing-pool" {
				http.Error(w, "The compute pool name already exists, please try with another compute pool name", http.StatusConflict)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			computePool.Metadata.CreationTimestamp = &timeStamp
			err = json.NewEncoder(w).Encode(computePool)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "cmf/api/v1/environments/{envName}/compute-pools/{poolName}"
// Used by describe, delete compute pools, no update compute pools.
func handleCmfComputePool(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["environment"]
		poolName := vars["poolName"]

		if environment == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if poolName == "invalid-pool" {
				http.Error(w, "The compute pool is invalid", http.StatusNotFound)
				return
			}

			computePool := createComputePool(poolName, "RUNNING")
			err := json.NewEncoder(w).Encode(computePool)
			require.NoError(t, err)
			return
		case http.MethodDelete:
			if poolName == "non-exist-pool" {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "cmf/api/v1/catalogs/kafka"
// Used by list, create Kafka catalogs
func handleCmfCatalogs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)
		switch r.Method {
		case http.MethodGet:
			catalog1 := createKafkaCatalog("test-catalog1")
			catalog2 := createKafkaCatalog("test-catalog2")
			catalog3 := createKafkaCatalog("test-catalog3")

			catalogs := []cmfsdk.KafkaCatalog{catalog1, catalog2, catalog3}
			catalogsPage := cmfsdk.KafkaCatalogsPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				catalogsPage.SetItems(catalogs)
			}

			err := json.NewEncoder(w).Encode(catalogsPage)
			require.NoError(t, err)
			return
		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var catalog cmfsdk.KafkaCatalog
			err = json.Unmarshal(reqBody, &catalog)
			require.NoError(t, err)

			catName := catalog.GetMetadata().Name

			if catName == "invalid-catalog" {
				http.Error(w, "The Kafka catalog object from resource file is invalid", http.StatusUnprocessableEntity)
				return
			}
			if catName == "existing-catalog" {
				http.Error(w, "The Kafka catalog name already exists, please try with another catalog name", http.StatusConflict)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			catalog.Metadata.CreationTimestamp = &timeStamp
			err = json.NewEncoder(w).Encode(catalog)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "cmf/api/v1/catalogs/kafka/{catName}"
// Used by describe, delete catalog, no update catalog.
func handleCmfCatalog(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		catalogName := vars["catName"]

		switch r.Method {
		case http.MethodGet:
			if catalogName == "invalid-catalog" {
				http.Error(w, "The catalog name is invalid", http.StatusNotFound)
				return
			}

			catalog := createKafkaCatalog(catalogName)
			err := json.NewEncoder(w).Encode(catalog)
			require.NoError(t, err)
			return
		case http.MethodDelete:
			if catalogName == "non-exist-catalog" {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "/cmf/api/v1/environments/{envName}/statements/{stmtName}"
// Used by describe, delete or update Flink statement.
func handleCmfStatement(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		stmtName := vars["stmtName"]
		environment := vars["environment"]

		if environment == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if stmtName == "invalid-stmt" {
				http.Error(w, "The statement name is invalid", http.StatusNotFound)
				return
			}

			if stmtName == "shell-test-stmt" {
				stmt := cmfsdk.Statement{
					Metadata: cmfsdk.StatementMetadata{
						Name: stmtName,
					},
					Status: &cmfsdk.StatementStatus{
						Phase:  "COMPLETED",
						Detail: cmfsdk.PtrString("Statement execution completed."),
						Traits: &cmfsdk.StatementTraits{
							SqlKind:       cmfsdk.PtrString("DESCRIBE"),
							UpsertColumns: nil,
							Schema: &cmfsdk.ResultSchema{
								Columns: []cmfsdk.ResultSchemaColumn{
									{
										Name: "name",
										Type: cmfsdk.DataType{
											Type:        "VARCHAR",
											Nullable:    true,
											KeyType:     nil,
											ValueType:   nil,
											ElementType: nil,
											Fields:      nil,
										},
									},
									{
										Name: "type",
										Type: cmfsdk.DataType{
											Type:        "VARCHAR",
											Nullable:    true,
											KeyType:     nil,
											ValueType:   nil,
											ElementType: nil,
											Fields:      nil,
										},
									},
									{
										Name: "null",
										Type: cmfsdk.DataType{
											Type:        "BOOLEAN",
											Nullable:    true,
											KeyType:     nil,
											ValueType:   nil,
											ElementType: nil,
											Fields:      nil,
										},
									},
								},
							},
						},
					},
					Result: &cmfsdk.StatementResult{
						Results: cmfsdk.StatementResults{
							Data: &[]map[string]interface{}{
								{
									"op":  0,
									"row": []string{"click_id", "STRING", "false"},
								},
								{
									"op":  0,
									"row": []string{"USER_id", "INT", "false"},
								},
								{
									"op":  0,
									"row": []string{"url", "STRING", "false"},
								},
								{
									"op":  0,
									"row": []string{"user_agent", "STRING", "false"},
								},
								{
									"op":  0,
									"row": []string{"view_time", "INT", "false"},
								},
							},
						},
					},
				}
				err := json.NewEncoder(w).Encode(stmt)
				require.NoError(t, err)
			} else {
				stmt := createFlinkStatement(stmtName, false, 1)
				err := json.NewEncoder(w).Encode(stmt)
				require.NoError(t, err)
			}
			return
		case http.MethodDelete:
			if stmtName == "non-exist-stmt" {
				http.Error(w, "The statement name can't be found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodPut:
			if stmtName == "non-exist-stmt" {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			// Read the existing statement from the request body and return it as the updated statement.
			req := new(cmfsdk.Statement)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(cmfsdk.Statement{})
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "/cmf/api/v1/environments/{envName}/statements"
// Used by list, create Flink statements
func handleCmfStatements(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["environment"]

		if environment == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			stmt1 := createFlinkStatement("test-stmt1", false, 1)
			stmt2 := createFlinkStatement("test-stmt2", false, 2)
			stmt3 := createFlinkStatement("test-stmt3", true, 4)

			stmts := []cmfsdk.Statement{stmt1, stmt2, stmt3}
			stmtsPage := cmfsdk.StatementsPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				stmtsPage.SetItems(stmts)
			}

			err := json.NewEncoder(w).Encode(stmtsPage)
			require.NoError(t, err)
			return
		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var stmt cmfsdk.Statement
			err = json.Unmarshal(reqBody, &stmt)
			require.NoError(t, err)

			stmtName := stmt.GetMetadata().Name

			if stmtName == "invalid-stmt" {
				http.Error(w, "The Flink statement is invalid", http.StatusUnprocessableEntity)
				return
			}
			if stmtName == "existing-stmt" {
				http.Error(w, "The Flink statement name already exists, please try with another statement name", http.StatusConflict)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			stmt.Metadata.CreationTimestamp = &timeStamp

			status := cmfsdk.StatementStatus{}
			if stmtName == "shell-test-stmt" {
				status = cmfsdk.StatementStatus{
					Phase:  "COMPLETED",
					Detail: cmfsdk.PtrString("Statement execution completed."),
					Traits: &cmfsdk.StatementTraits{
						SqlKind:       cmfsdk.PtrString("DESCRIBE"),
						UpsertColumns: nil,
						Schema: &cmfsdk.ResultSchema{
							Columns: []cmfsdk.ResultSchemaColumn{
								{
									Name: "name",
									Type: cmfsdk.DataType{
										Type:        "VARCHAR",
										Nullable:    true,
										KeyType:     nil,
										ValueType:   nil,
										ElementType: nil,
										Fields:      nil,
									},
								},
								{
									Name: "type",
									Type: cmfsdk.DataType{
										Type:        "VARCHAR",
										Nullable:    true,
										KeyType:     nil,
										ValueType:   nil,
										ElementType: nil,
										Fields:      nil,
									},
								},
								{
									Name: "null",
									Type: cmfsdk.DataType{
										Type:        "BOOLEAN",
										Nullable:    true,
										KeyType:     nil,
										ValueType:   nil,
										ElementType: nil,
										Fields:      nil,
									},
								},
							},
						},
					},
				}
			} else {
				status = cmfsdk.StatementStatus{
					Phase:  "PENDING",
					Detail: cmfsdk.PtrString("Statement is pending execution."),
					Traits: &cmfsdk.StatementTraits{
						SqlKind:      cmfsdk.PtrString("SELECT"),
						IsAppendOnly: cmfsdk.PtrBool(false),
						IsBounded:    cmfsdk.PtrBool(false),
					},
				}
			}
			stmt.Status = &status
			err = json.NewEncoder(w).Encode(stmt)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

func handleCmfStatementExceptions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["environment"]
		stmtName := vars["stmtName"]

		if environment == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if stmtName == "invalid-stmt" {
				http.Error(w, "The Flink statement is invalid", http.StatusUnprocessableEntity)
				return
			}
			data := createStatementExceptionData()
			exceptions := cmfsdk.StatementExceptionList{
				Data: data,
			}
			err := json.NewEncoder(w).Encode(exceptions)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}
