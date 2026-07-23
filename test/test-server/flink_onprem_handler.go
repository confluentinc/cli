package testserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

const invalidSecretMappingName = "invalid-secret-mapping"

func createSecretMapping(name string) cmfsdk.EnvironmentSecretMapping {
	timeStamp := time.Date(2025, time.August, 5, 12, 0, 0, 0, time.UTC).String()
	mappingName := name
	secretName := "my-actual-secret"
	return cmfsdk.EnvironmentSecretMapping{
		ApiVersion: "cmf/v1",
		Kind:       "EnvironmentSecretMapping",
		Metadata: &cmfsdk.EnvironmentSecretMappingMetadata{
			Name:              &mappingName,
			CreationTimestamp: &timeStamp,
		},
		Spec: &cmfsdk.EnvironmentSecretMappingSpec{
			SecretName: secretName,
		},
	}
}

const invalidSecretName = "invalid-secret"
const invalidDatabaseName = "invalid-database"

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

func createSavepoint(name string) cmfsdk.Savepoint {
	timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
	path := "abc/def"
	backLimit := int32(10)
	format := "CANONICAL"

	return cmfsdk.Savepoint{
		Metadata: cmfsdk.SavepointMetadata{
			Name:              &name,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.SavepointSpec{
			Path:         &path,
			BackoffLimit: &backLimit,
			FormatType:   &format,
		},
	}
}

func createComputePool(poolName, phase string) cmfsdk.ComputePool {
	timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()

	status := map[string]interface{}{
		"phase": phase,
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
	timeStamp := time.Date(2025, time.August, 5, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	return cmfsdk.KafkaCatalog{
		ApiVersion: "cmf/api/v1/catalog",
		Kind:       "KafkaCatalog",
		Metadata: cmfsdk.CatalogMetadata{
			Name:              catName,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.KafkaCatalogSpec{
			SrInstance: cmfsdk.KafkaCatalogSpecSrInstance{
				ConnectionConfig: map[string]string{
					"schema.registry.url": "http://localhost:8081",
				},
			},
			KafkaClusters: &[]cmfsdk.KafkaCatalogSpecKafkaClusters{
				{
					DatabaseName: "test-database",
					ConnectionConfig: map[string]string{
						"bootstrap.servers": "localhost:9092",
					},
				},
				{
					DatabaseName: "test-database-2",
					ConnectionConfig: map[string]string{
						"bootstrap.servers": "localhost:9092",
					},
				},
			},
		},
	}
}

func createKafkaDatabase(dbName string) cmfsdk.KafkaDatabase {
	timeStamp := time.Date(2025, time.August, 5, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	return cmfsdk.KafkaDatabase{
		ApiVersion: "cmf/api/v1/database",
		Kind:       "KafkaDatabase",
		Metadata: cmfsdk.DatabaseMetadata{
			Name:              dbName,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.KafkaDatabaseSpec{
			KafkaCluster: cmfsdk.KafkaDatabaseSpecKafkaCluster{
				ConnectionConfig: map[string]string{
					"bootstrap.servers": "localhost:9092",
				},
			},
		},
	}
}

func handleCmfCatalogDatabases(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)
		switch r.Method {
		case http.MethodGet:
			databases := []cmfsdk.KafkaDatabase{
				createKafkaDatabase("test-database-1"),
				createKafkaDatabase("test-database-2"),
			}
			databasesPage := cmfsdk.KafkaDatabasesPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				databasesPage.SetItems(databases)
			}

			err := json.NewEncoder(w).Encode(databasesPage)
			require.NoError(t, err)
		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var database cmfsdk.KafkaDatabase
			err = json.Unmarshal(reqBody, &database)
			require.NoError(t, err)

			dbName := database.GetMetadata().Name

			if dbName == invalidDatabaseName {
				http.Error(w, "The Kafka database object from resource file is invalid", http.StatusUnprocessableEntity)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			database.Metadata.CreationTimestamp = &timeStamp
			err = json.NewEncoder(w).Encode(database)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

func handleCmfCatalogDatabase(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		dbName := vars["dbName"]

		switch r.Method {
		case http.MethodGet:
			if dbName == invalidDatabaseName {
				http.Error(w, "The database name is invalid", http.StatusNotFound)
				return
			}

			database := createKafkaDatabase(dbName)
			err := json.NewEncoder(w).Encode(database)
			require.NoError(t, err)
			return
		case http.MethodPut:
			if dbName == invalidDatabaseName {
				http.Error(w, "The database name is invalid", http.StatusNotFound)
				return
			}

			// Read and validate the request body.
			req := new(cmfsdk.KafkaDatabase)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
			return
		case http.MethodDelete:
			if dbName == "non-exist-database" {
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

func createFlinkStatementSavepoint(stmtName string, stopped bool, parallelism int32) cmfsdk.Statement {
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

	savepointPath := "savepointPath"
	allowNonRestored := true
	savepoint := cmfsdk.StatementStartFromSavepoint{
		InitialSavepointPath:  &savepointPath,
		AllowNonRestoredState: &allowNonRestored,
	}

	return cmfsdk.Statement{
		Metadata: cmfsdk.StatementMetadata{
			Name:              stmtName,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.StatementSpec{
			Statement:          "SELECT * FROM test_table",
			ComputePoolName:    "test-pool",
			Parallelism:        cmfsdk.PtrInt32(parallelism),
			Stopped:            cmfsdk.PtrBool(stopped),
			StartFromSavepoint: &savepoint,
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

			envName := environment.GetName()
			if strings.Contains(envName, "failure") {
				http.Error(w, "", http.StatusUnprocessableEntity)
				return
			}

			// Already existing environment: update
			if envName == "default" || envName == "test" {
				outputEnvironment := createEnvironment(envName, envName+"-namespace")
				// This is a dummy update - only the defaults can be updated anyway.
				outputEnvironment.FlinkApplicationDefaults = environment.FlinkApplicationDefaults
				outputEnvironment.ComputePoolDefaults = environment.ComputePoolDefaults
				outputEnvironment.StatementDefaults = environment.StatementDefaults
				err = json.NewEncoder(w).Encode(outputEnvironment)
				require.NoError(t, err)
				return
			}

			// New environment: create
			outputEnvironment := createEnvironment(envName, environment.GetKubernetesNamespace())
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

// Handler for "cmf/api/v1alpha1/environments/{envName}/applications/{appName}/events"
func handleCmfApplicationEvents(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		envName := vars["envName"]
		appName := vars["appName"]

		if r.Method != http.MethodGet {
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
			return
		}

		if envName != "default" && envName != "test" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		// Validate application exists, matching handleCmfApplication behavior.
		validApps := map[string]bool{
			"default-application-1": true,
			"default-application-2": true,
			"default-application-s": true,
		}
		if envName == "default" && !validApps[appName] {
			http.Error(w, "Application not found", http.StatusNotFound)
			return
		}

		eventsPage := map[string]interface{}{
			"items": []cmfsdk.FlinkApplicationEvent{},
		}

		page := r.URL.Query().Get("page")

		if envName == "default" && appName == "default-application-1" && page == "0" {
			timestamp1 := "2024-01-15T10:30:00Z"
			timestamp2 := "2024-01-15T10:31:00Z"
			instance := "default-application-1-instance-1"
			eventType1 := "Normal"
			eventType2 := "Warning"
			message1 := "Application started successfully"
			message2 := "Application restarting due to failure"
			name1 := "event-001"
			name2 := "event-002"
			newStatus := "DEPLOYED"
			exceptionString := "java.lang.RuntimeException: Job execution failed"
			newStatusData := cmfsdk.EventDataNewStatusAsEventData(&cmfsdk.EventDataNewStatus{NewStatus: &newStatus})
			jobExceptionData := cmfsdk.EventDataJobExceptionAsEventData(&cmfsdk.EventDataJobException{ExceptionString: &exceptionString})

			events := []cmfsdk.FlinkApplicationEvent{
				{
					ApiVersion: "cmf.confluent.io/v1alpha1",
					Kind:       "FlinkApplicationEvent",
					Metadata: cmfsdk.EventMetadata{
						Name:                     &name1,
						CreationTimestamp:        &timestamp1,
						FlinkApplicationInstance: &instance,
					},
					Status: cmfsdk.EventStatus{
						Type:    &eventType1,
						Message: &message1,
						Data:    &newStatusData,
					},
				},
				{
					ApiVersion: "cmf.confluent.io/v1alpha1",
					Kind:       "FlinkApplicationEvent",
					Metadata: cmfsdk.EventMetadata{
						Name:                     &name2,
						CreationTimestamp:        &timestamp2,
						FlinkApplicationInstance: &instance,
					},
					Status: cmfsdk.EventStatus{
						Type:    &eventType2,
						Message: &message2,
						Data:    &jobExceptionData,
					},
				},
			}
			eventsPage = map[string]interface{}{
				"items": events,
			}
		}

		err := json.NewEncoder(w).Encode(eventsPage)
		require.NoError(t, err)
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
				items := []cmfsdk.FlinkApplication{createApplication("default-application-1"), createApplication("default-application-2"), createApplication("default-application-s")}
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

			if applicationName == "default-application-s" {
				// The 'update' is going to be spec.serviceAccount. This is just a dummy update,
				// and we don't do any actual merge logic.
				outputApplication := createApplication(applicationName)
				outputApplication.Spec["serviceAccount"] = application.Spec["serviceAccount"]
				outputApplication.Spec["job"] = application.Spec["job"]
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
			if (application == "default-application-1" || application == "default-application-2" || application == "default-application-s") && environment == "default" {
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

// Helper function to create a Flink application instance.
func createApplicationInstance(name, jobId, state, creationTimestamp string) cmfsdk.FlinkApplicationInstance {
	instance := cmfsdk.FlinkApplicationInstance{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "FlinkApplicationInstance",
	}

	metadata := cmfsdk.ApplicationInstanceMetadata{
		Name:              &name,
		Uid:               &name,
		CreationTimestamp: &creationTimestamp,
	}
	instance.Metadata = &metadata

	jobStatus := cmfsdk.ApplicationInstanceStatusJobStatus{
		JobId: &jobId,
		State: &state,
	}
	status := cmfsdk.ApplicationInstanceStatus{
		JobStatus: &jobStatus,
	}
	instance.Status = &status

	return instance
}

// Handler for "cmf/api/v1/environments/{environment}/applications/{application}/instances"
// Used to list application instances.
func handleCmfApplicationInstances(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["environment"]
		application := vars["application"]

		switch r.Method {
		case http.MethodGet:
			if environment != "default" && environment != "test" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}

			if application != "default-application-1" && application != "default-application-2" {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}

			instancesPage := map[string]interface{}{
				"items": []cmfsdk.FlinkApplicationInstance{},
			}

			page := r.URL.Query().Get("page")

			if application == "default-application-1" && page == "0" {
				items := []cmfsdk.FlinkApplicationInstance{
					createApplicationInstance("inst-001", "job-abc123", "RUNNING", "2025-09-18T10:00:00Z"),
					createApplicationInstance("inst-002", "job-def456", "FINISHED", "2025-09-17T08:30:00Z"),
				}
				instancesPage = map[string]interface{}{
					"items": items,
				}
			}

			err := json.NewEncoder(w).Encode(instancesPage)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "cmf/api/v1/environments/{environment}/applications/{application}/instances/{instName}"
// Used to describe a specific application instance.
func handleCmfApplicationInstance(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["environment"]
		application := vars["application"]
		instName := vars["instName"]

		switch r.Method {
		case http.MethodGet:
			if environment != "default" && environment != "test" {
				http.Error(w, "Environment not found", http.StatusNotFound)
				return
			}

			if application != "default-application-1" && application != "default-application-2" {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}

			if application == "default-application-1" && instName == "inst-001" {
				instance := createApplicationInstance("inst-001", "job-abc123", "RUNNING", "2025-09-18T10:00:00Z")
				err := json.NewEncoder(w).Encode(instance)
				require.NoError(t, err)
				return
			}

			http.Error(w, "Instance not found", http.StatusNotFound)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "/cmf/api/v1/environments/{envName}/applications/{appName}/savepoints"
// Used by list, create savepoints.
func handleCmfSavepoints(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["envName"]

		if environment == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			savepoint1 := createSavepoint("test-savepoint1")
			savepoint2 := createSavepoint("test-savepoint2")
			savepoint3 := createSavepoint("test-savepoint3")

			savepoints := []cmfsdk.Savepoint{savepoint1, savepoint2, savepoint3}
			savepointsPage := cmfsdk.SavepointsPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				savepointsPage.SetItems(savepoints)
			}

			err := json.NewEncoder(w).Encode(savepointsPage)
			require.NoError(t, err)
			return

		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var savepoint cmfsdk.Savepoint
			err = json.Unmarshal(reqBody, &savepoint)
			require.NoError(t, err)

			savepointName := savepoint.Metadata.GetName()

			if savepointName == "invalid-pool" {
				http.Error(w, "The savepoint object from resource file is invalid", http.StatusUnprocessableEntity)
				return
			}
			if savepointName == "existing-pool" {
				http.Error(w, "The savepoint name already exists, please try with another savepoint name", http.StatusConflict)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			savepoint.Metadata.CreationTimestamp = &timeStamp
			err = json.NewEncoder(w).Encode(savepoint)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

func handleCmfSavepoint(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		environment := vars["envName"]
		savepointName := vars["savepointName"]

		if environment == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if savepointName == "invalid-savepoint" {
				http.Error(w, "The savepoint is invalid", http.StatusNotFound)
				return
			}

			savepoint := createSavepoint(savepointName)
			err := json.NewEncoder(w).Encode(savepoint)
			require.NoError(t, err)
			return
		case http.MethodDelete:
			if savepointName == "non-exist-savepoint" {
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

func handleCmfDetachedSavepoints(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		switch r.Method {
		case http.MethodGet:
			savepoint1 := createSavepoint("test-savepoint1")
			savepoint2 := createSavepoint("test-savepoint2")
			savepoint3 := createSavepoint("test-savepoint3")

			savepoints := []cmfsdk.Savepoint{savepoint1, savepoint2, savepoint3}
			savepointsPage := cmfsdk.SavepointsPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				savepointsPage.SetItems(savepoints)
			}

			err := json.NewEncoder(w).Encode(savepointsPage)
			require.NoError(t, err)
			return

		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var savepoint cmfsdk.Savepoint
			err = json.Unmarshal(reqBody, &savepoint)
			require.NoError(t, err)

			savepointName := savepoint.Metadata.GetName()

			if savepointName == "invalid-pool" {
				http.Error(w, "The savepoint object from resource file is invalid", http.StatusUnprocessableEntity)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			savepoint.Metadata.CreationTimestamp = &timeStamp
			savepoint.Metadata.SetUid("id1")
			savepoint.Spec.SetFormatType("Canonical")
			savepoint.Spec.SetBackoffLimit(10)
			err = json.NewEncoder(w).Encode(savepoint)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

func handleCmfDetachedSavepoint(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)
		vars := mux.Vars(r)
		savepointName := vars["detachedSavepointName"]
		switch r.Method {
		case http.MethodGet:
			savepoint := createSavepoint("savepoint1")
			if savepointName == "invalid-savepoint" {
				http.Error(w, "The savepoint is invalid", http.StatusNotFound)
				return
			}
			err := json.NewEncoder(w).Encode(savepoint)
			require.NoError(t, err)
			return
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			return
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
// Used by describe, update, delete catalog.
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
		case http.MethodPut:
			if catalogName == "invalid-catalog" {
				http.Error(w, "The catalog name is invalid", http.StatusNotFound)
				return
			}

			req := new(cmfsdk.KafkaCatalog)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
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
			} else if stmtName == "test-stmt-savepoint" {
				stmt := createFlinkStatementSavepoint(stmtName, false, 1)
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
			stmt4 := createFlinkStatementSavepoint("test-stmt4", true, 4)

			stmts := []cmfsdk.Statement{stmt1, stmt2, stmt3, stmt4}
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

			if stmtName == "stmt-savepoint" {
				savepointName := "savepoint1"
				allowNonRestored := false
				savepoint := cmfsdk.StatementStartFromSavepoint{
					SavepointName:         &savepointName,
					AllowNonRestoredState: &allowNonRestored,
				}
				stmt.Spec.SetStartFromSavepoint(savepoint)
			}
			if stmtName == "stmt-savepoint2" {
				savepointPath := "savepointPath"
				allowNonRestored := true
				savepoint := cmfsdk.StatementStartFromSavepoint{
					InitialSavepointPath:  &savepointPath,
					AllowNonRestoredState: &allowNonRestored,
				}
				stmt.Spec.SetStartFromSavepoint(savepoint)
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

func handleCmfSystemInformation(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		switch r.Method {
		case http.MethodGet:
			sysInfo := map[string]interface{}{
				"status": map[string]interface{}{
					"version":  "1.0.0",
					"revision": "abc1234def5678",
				},
			}
			err := json.NewEncoder(w).Encode(sysInfo)
			require.NoError(t, err)
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

func handleCmfSecretMappings(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)
		switch r.Method {
		case http.MethodGet:
			mappings := []cmfsdk.EnvironmentSecretMapping{
				createSecretMapping("test-mapping-1"),
				createSecretMapping("test-mapping-2"),
			}
			mappingsPage := cmfsdk.EnvironmentSecretMappingsPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				mappingsPage.SetItems(mappings)
			}

			err := json.NewEncoder(w).Encode(mappingsPage)
			require.NoError(t, err)
		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var mapping cmfsdk.EnvironmentSecretMapping
			err = json.Unmarshal(reqBody, &mapping)
			require.NoError(t, err)

			var mappingName string
			if mapping.Metadata != nil && mapping.Metadata.Name != nil {
				mappingName = *mapping.Metadata.Name
			}

			if mappingName == invalidSecretMappingName {
				http.Error(w, "The EnvironmentSecretMapping object from resource file is invalid", http.StatusUnprocessableEntity)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			if mapping.Metadata == nil {
				mapping.Metadata = &cmfsdk.EnvironmentSecretMappingMetadata{}
			}
			mapping.Metadata.CreationTimestamp = &timeStamp
			err = json.NewEncoder(w).Encode(mapping)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

func handleCmfSecretMapping(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		name := vars["name"]

		switch r.Method {
		case http.MethodGet:
			if name == invalidSecretMappingName {
				http.Error(w, "The secret mapping name is invalid", http.StatusNotFound)
				return
			}

			mapping := createSecretMapping(name)
			err := json.NewEncoder(w).Encode(mapping)
			require.NoError(t, err)
			return
		case http.MethodPut:
			if name == invalidSecretMappingName {
				http.Error(w, "The secret mapping name is invalid", http.StatusNotFound)
				return
			}

			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var mapping cmfsdk.EnvironmentSecretMapping
			err = json.Unmarshal(reqBody, &mapping)
			require.NoError(t, err)

			timeStamp := time.Date(2025, time.August, 5, 12, 0, 0, 0, time.UTC).String()
			if mapping.Metadata == nil {
				mapping.Metadata = &cmfsdk.EnvironmentSecretMappingMetadata{}
			}
			mapping.Metadata.CreationTimestamp = &timeStamp
			err = json.NewEncoder(w).Encode(mapping)
			require.NoError(t, err)
			return
		case http.MethodDelete:
			if name == "non-exist-secret-mapping" {
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

func createSecret(secretName string) cmfsdk.Secret {
	timeStamp := time.Date(2025, time.August, 5, 12, 0, 0, 0, time.UTC).String()
	maskedData := map[string]string{
		"bootstrap.servers": "****",
		"sasl.jaas.config":  "****",
	}
	return cmfsdk.Secret{
		ApiVersion: "cmf/v1",
		Kind:       "Secret",
		Metadata: cmfsdk.SecretMetadata{
			Name:              secretName,
			CreationTimestamp: &timeStamp,
		},
		Spec: cmfsdk.SecretSpec{
			Data: &maskedData,
		},
	}
}

func handleCmfSecrets(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)
		switch r.Method {
		case http.MethodGet:
			secrets := []cmfsdk.Secret{
				createSecret("test-secret-1"),
				createSecret("test-secret-2"),
			}
			secretsPage := cmfsdk.SecretsPage{}
			page := r.URL.Query().Get("page")

			if page == "0" {
				secretsPage.SetItems(secrets)
			}

			err := json.NewEncoder(w).Encode(secretsPage)
			require.NoError(t, err)
		case http.MethodPost:
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var secret cmfsdk.Secret
			err = json.Unmarshal(reqBody, &secret)
			require.NoError(t, err)

			secretName := secret.Metadata.Name

			if secretName == invalidSecretName {
				http.Error(w, "The Secret object from resource file is invalid", http.StatusUnprocessableEntity)
				return
			}

			timeStamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
			secret.Metadata.CreationTimestamp = &timeStamp
			// Mask the secret data in response
			maskedData := make(map[string]string)
			if secret.Spec.Data != nil {
				for k := range *secret.Spec.Data {
					maskedData[k] = "****"
				}
			}
			secret.Spec.Data = &maskedData
			err = json.NewEncoder(w).Encode(secret)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

func handleCmfSecret(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		secretName := vars["secretName"]

		switch r.Method {
		case http.MethodGet:
			if secretName == invalidSecretName {
				http.Error(w, "The secret name is invalid", http.StatusNotFound)
				return
			}

			secret := createSecret(secretName)
			err := json.NewEncoder(w).Encode(secret)
			require.NoError(t, err)
			return
		case http.MethodPut:
			if secretName == invalidSecretName {
				http.Error(w, "The secret name is invalid", http.StatusNotFound)
				return
			}

			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var secret cmfsdk.Secret
			err = json.Unmarshal(reqBody, &secret)
			require.NoError(t, err)

			timeStamp := time.Date(2025, time.August, 5, 12, 0, 0, 0, time.UTC).String()
			secret.Metadata.CreationTimestamp = &timeStamp
			// Mask the secret data in response
			maskedData := make(map[string]string)
			if secret.Spec.Data != nil {
				for k := range *secret.Spec.Data {
					maskedData[k] = "****"
				}
			}
			secret.Spec.Data = &maskedData
			err = json.NewEncoder(w).Encode(secret)
			require.NoError(t, err)
			return
		case http.MethodDelete:
			if secretName == "non-exist-secret" {
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

// assertArtifactLabelContract enforces the CLI's label contract on the submitted artifact JSON: the "labels" field must
// be omitted entirely to preserve existing labels (nil map / --label not passed), and the CLI must never send an empty
// object, which CMF interprets as "clear all labels". When present, labels must carry the caller's entries.
func assertArtifactLabelContract(t *testing.T, rawArtifact string) {
	var probe struct {
		Metadata struct {
			Labels *map[string]string `json:"labels"`
		} `json:"metadata"`
	}
	require.NoError(t, json.Unmarshal([]byte(rawArtifact), &probe))
	if probe.Metadata.Labels != nil {
		require.NotEmpty(t, *probe.Metadata.Labels, "CLI must omit the labels field to preserve labels, never send an empty object")
	}
}

// createArtifactObject builds a fully-populated Artifact for the given name and version with deterministic field values.
func createArtifactObject(name string, version int32) cmfsdk.Artifact {
	timestamp := time.Date(2025, time.March, 12, 23, 42, 0, 0, time.UTC).String()
	return cmfsdk.Artifact{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Artifact",
		Metadata: cmfsdk.ArtifactMetadata{
			Name:              name,
			Uid:               cmfsdk.PtrString("11111111-1111-1111-1111-111111111111"),
			CreationTimestamp: &timestamp,
			UpdateTimestamp:   &timestamp,
		},
		Spec: map[string]interface{}{},
		Status: &cmfsdk.ArtifactStatus{
			Version:           cmfsdk.PtrInt32(version),
			CreationTimestamp: &timestamp,
			Path:              cmfsdk.PtrString(fmt.Sprintf("artifacts/%s/%d", name, version)),
			Size:              cmfsdk.PtrInt64(1024),
			Checksum:          cmfsdk.PtrString("d41d8cd98f00b204e9800998ecf8427e"),
			Phase:             cmfsdk.PtrString("READY"),
		},
	}
}

// buildArtifactResponse echoes the submitted name, labels, and annotations while attaching a deterministic server-side status.
func buildArtifactResponse(submitted cmfsdk.Artifact, version int32) cmfsdk.Artifact {
	artifact := createArtifactObject(submitted.Metadata.Name, version)
	if submitted.Metadata.Labels != nil {
		artifact.Metadata.Labels = submitted.Metadata.Labels
	}
	if submitted.Metadata.Annotations != nil {
		artifact.Metadata.Annotations = submitted.Metadata.Annotations
	}
	return artifact
}

// Handler for "/cmf/api/v1/environments/{environment}/artifacts"
// Used by list and create Flink artifacts.
func handleCmfArtifacts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		if mux.Vars(r)["environment"] == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			artifactsPage := cmfsdk.ArtifactsPage{}
			if r.URL.Query().Get("page") == "0" {
				artifactsPage.SetItems([]cmfsdk.Artifact{
					createArtifactObject("test-artifact-1", 1),
					createArtifactObject("test-artifact-2", 3),
				})
			}
			err := json.NewEncoder(w).Encode(artifactsPage)
			require.NoError(t, err)
			return
		case http.MethodPost:
			require.NoError(t, r.ParseMultipartForm(32<<20))
			rawArtifact := r.FormValue("artifact")
			assertArtifactLabelContract(t, rawArtifact)
			var artifact cmfsdk.Artifact
			require.NoError(t, json.Unmarshal([]byte(rawArtifact), &artifact))

			if artifact.Metadata.Name == "existing-artifact" {
				http.Error(w, "The artifact name already exists, please try with another artifact name", http.StatusConflict)
				return
			}

			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(buildArtifactResponse(artifact, 1))
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "/cmf/api/v1/environments/{environment}/artifacts/{artifactName}"
// Used by describe, update (metadata-only), version create (with file), and delete a Flink artifact.
func handleCmfArtifact(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		artifactName := vars["artifactName"]

		if vars["environment"] == "non-exist" {
			http.Error(w, "Environment not found", http.StatusNotFound)
			return
		}
		if artifactName == "invalid-artifact" || artifactName == "non-exist-artifact" {
			http.Error(w, "The artifact name is invalid", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			version := int32(3)
			if requested := r.URL.Query().Get("version"); requested != "" {
				parsed, err := strconv.Atoi(requested)
				if err != nil {
					http.Error(w, "invalid version", http.StatusBadRequest)
					return
				}
				version = int32(parsed)
			}
			artifact := createArtifactObject(artifactName, version)
			// Exercise the human describe view's Labels/Annotations rows for this fixture name.
			if artifactName == "labeled-artifact" {
				artifact.Metadata.Labels = map[string]string{"owner": "team-a", "tier": "gold"}
				artifact.Metadata.Annotations = map[string]string{"note": "managed-externally"}
			}
			err := json.NewEncoder(w).Encode(artifact)
			require.NoError(t, err)
			return
		case http.MethodPut:
			require.NoError(t, r.ParseMultipartForm(32<<20))
			rawArtifact := r.FormValue("artifact")
			assertArtifactLabelContract(t, rawArtifact)
			var artifact cmfsdk.Artifact
			require.NoError(t, json.Unmarshal([]byte(rawArtifact), &artifact))

			// A "file" part indicates a new version upload; its absence is a metadata-only update.
			version := int32(1)
			if r.MultipartForm != nil && len(r.MultipartForm.File["file"]) > 0 {
				version = 2
			}

			err := json.NewEncoder(w).Encode(buildArtifactResponse(artifact, version))
			require.NoError(t, err)
			return
		case http.MethodDelete:
			// Assert the CLI forwarded the --version flag verbatim: a version-scoped delete sends ?version=<value>,
			// a whole-artifact delete sends none. (These fixture names are used only by the version-delete tests.)
			version := r.URL.Query().Get("version")
			switch artifactName {
			case "delete-version-2":
				require.Equal(t, "2", version)
			case "delete-version-all":
				require.Equal(t, "all", version)
			default:
				require.Empty(t, version, "whole-artifact delete must not send a version query param")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "/cmf/api/v1/environments/{environment}/artifacts/{artifactName}/versions"
// Used by list the versions of a Flink artifact.
func handleCmfArtifactVersions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		artifactName := vars["artifactName"]

		if artifactName == "invalid-artifact" || artifactName == "non-exist-artifact" {
			http.Error(w, "The artifact name is invalid", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			versionsPage := cmfsdk.ArtifactsPage{}
			if r.URL.Query().Get("page") == "0" {
				versionsPage.SetItems([]cmfsdk.Artifact{
					createArtifactObject(artifactName, 3),
					createArtifactObject(artifactName, 2),
					createArtifactObject(artifactName, 1),
				})
			}
			err := json.NewEncoder(w).Encode(versionsPage)
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}

// Handler for "/cmf/api/v1/environments/{environment}/artifacts/{artifactName}/content"
// Used by download the content of a Flink artifact.
func handleCmfArtifactContent(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleLoginType(t, r)

		vars := mux.Vars(r)
		artifactName := vars["artifactName"]

		if artifactName == "invalid-artifact" || artifactName == "non-exist-artifact" {
			http.Error(w, "The artifact name is invalid", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.jar"`, artifactName))
			_, err := w.Write([]byte("dummy artifact content"))
			require.NoError(t, err)
			return
		default:
			require.Fail(t, fmt.Sprintf("Unexpected method %s", r.Method))
		}
	}
}
