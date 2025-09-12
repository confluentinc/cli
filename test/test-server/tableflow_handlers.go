package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2/tableflow/v1"
)

// Handler for: "/tableflow/v1/tableflow-topics/{display_name}"
func handleTableflowTopic(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		displayName := mux.Vars(r)["display_name"]
		environmentId := r.URL.Query().Get("environment")
		clusterId := r.URL.Query().Get("spec.kafka_cluster")
		switch r.Method {
		case http.MethodGet:
			handleTableflowTopicGet(t, environmentId, clusterId, displayName)(w, r)
		case http.MethodDelete:
			handleTableflowTopicDelete(t, displayName)(w, r)
		case http.MethodPatch:
			handleTableflowTopicUpdate(t, displayName)(w, r)
		}
	}
}

// Handler for: "/tableflow/v1/tableflow-topics"
func handleTableflowTopics(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environmentId := r.URL.Query().Get("environment")
		clusterId := r.URL.Query().Get("spec.kafka_cluster")
		switch r.Method {
		case http.MethodGet:
			handleTableflowTopicsList(t, environmentId, clusterId)(w, r)
		case http.MethodPost:
			handleTableflowTopicsCreate(t, environmentId)(w, r)
		}
	}
}

// Handler for: "/tableflow/v1/catalog-integrations/{id}"
func handleCatalogIntegration(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		environmentId := r.URL.Query().Get("environment")
		clusterId := r.URL.Query().Get("spec.kafka_cluster")
		switch r.Method {
		case http.MethodGet:
			handleCatalogIntegrationGet(t, environmentId, clusterId, id)(w, r)
		case http.MethodDelete:
			handleCatalogIntegrationDelete(t)(w, r)
		case http.MethodPatch:
			handleCatalogIntegrationUpdate(t, id)(w, r)
		}
	}
}

// Handler for: "/tableflow/v1/catalog-integrations"
func handleCatalogIntegrations(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environmentId := r.URL.Query().Get("environment")
		clusterId := r.URL.Query().Get("spec.kafka_cluster")
		switch r.Method {
		case http.MethodGet:
			handleCatalogIntegrationList(t, environmentId, clusterId)(w, r)
		case http.MethodPost:
			handleCatalogIntegrationCreate(t)(w, r)
		}
	}
}

func handleTableflowTopicsCreate(t *testing.T, environment string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tableflowTopic := &tableflowv1.TableflowV1TableflowTopic{}
		err := json.NewDecoder(r.Body).Decode(tableflowTopic)
		require.NoError(t, err)

		tableflowTopic.Spec.Config.SetEnableCompaction(true) //readonly attribute, always true
		tableflowTopic.Status = tableflowv1.NewTableflowV1TableflowTopicStatus("APPEND")
		tableflowTopic.Status.SetPhase("RUNNING")
		tableflowTopic.Spec.SetEnvironment(tableflowv1.GlobalObjectReference{Id: environment})

		if tableflowTopic.Spec.Storage.TableflowV1ByobAwsSpec != nil {
			tableflowTopic.Spec.Storage.TableflowV1ByobAwsSpec.SetBucketRegion("us-east-1")
			tableflowTopic.Spec.Storage.TableflowV1ByobAwsSpec.SetTablePath("s3://dummy-bucket-name-1//10011010/11101100/org-1/env-2/lkc-3/v1/tableId")
		} else if tableflowTopic.Spec.Storage.TableflowV1ManagedStorageSpec != nil {
			tableflowTopic.Spec.Storage.TableflowV1ManagedStorageSpec.SetTablePath("s3://dummy-bucket-name-1//10011010/11101100/org-1/env-2/lkc-3/v1/tableId")
		}

		err = json.NewEncoder(w).Encode(tableflowTopic)
		require.NoError(t, err)
	}
}

func handleTableflowTopicGet(t *testing.T, environmentId, clusterId, display_name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tableflowTopic tableflowv1.TableflowV1TableflowTopic

		switch display_name {
		case "topic-invalid":
			w.WriteHeader(http.StatusNotFound)
		case "topic-byob":
			tableflowTopic = getTopicByob("topic-byob", environmentId, clusterId)
		case "topic-managed":
			tableflowTopic = getTopicManaged("topic-managed", environmentId, clusterId)
		}
		err := json.NewEncoder(w).Encode(tableflowTopic)
		require.NoError(t, err)
	}
}

func handleTableflowTopicsList(t *testing.T, environmentId, clusterId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicOne := getTopicByob("topic-byob", environmentId, clusterId)
		topicTwo := getTopicManaged("topic-managed", environmentId, clusterId)

		recordList := tableflowv1.TableflowV1TableflowTopicList{Data: []tableflowv1.TableflowV1TableflowTopic{topicOne, topicTwo}}
		setPageToken(&recordList, &recordList.Metadata, r.URL)
		err := json.NewEncoder(w).Encode(recordList)
		require.NoError(t, err)
	}
}

func handleTableflowTopicUpdate(t *testing.T, display_name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &tableflowv1.TableflowV1TableflowTopicUpdate{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		var tableflowTopic tableflowv1.TableflowV1TableflowTopic
		switch display_name {
		case "topic-invalid":
			w.WriteHeader(http.StatusNotFound)
		case "topic-byob":
			tableflowTopic = getTopicByob("topic-byob", "env-596", "lkc-123456")
		case "topic-managed":
			tableflowTopic = getTopicManaged("topic-managed", "env-596", "lkc-123456")
		}

		if body.Spec.Config.GetRetentionMs() != "" {
			tableflowTopic.Spec.Config.SetRetentionMs(body.Spec.Config.GetRetentionMs())
		}

		err = json.NewEncoder(w).Encode(tableflowTopic)
		require.NoError(t, err)
	}
}

func handleTableflowTopicDelete(t *testing.T, display_name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch display_name {
		case "invalid-topic":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The Tableflow topic was not found.")
			require.NoError(t, err)
		case "topic-byob", "topic-managed":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func getTopicByob(display_name, environmentId, clusterId string) tableflowv1.TableflowV1TableflowTopic {
	return tableflowv1.TableflowV1TableflowTopic{
		Spec: &tableflowv1.TableflowV1TableflowTopicSpec{
			DisplayName: tableflowv1.PtrString(display_name),
			Suspended:   tableflowv1.PtrBool(false),
			Storage: &tableflowv1.TableflowV1TableflowTopicSpecStorageOneOf{
				TableflowV1ByobAwsSpec: &tableflowv1.TableflowV1ByobAwsSpec{
					Kind:                  "ByobAws",
					BucketName:            "bucket_1",
					BucketRegion:          tableflowv1.PtrString("us-east-1"),
					ProviderIntegrationId: "cspi-stgce89r7",
					TablePath:             tableflowv1.PtrString("s3://dummy-bucket-name-1//10011010/11101100/org-1/env-2/lkc-3/v1/tableId"),
				},
			},
			Config: &tableflowv1.TableflowV1TableFlowTopicConfigsSpec{
				EnableCompaction:      tableflowv1.PtrBool(true),
				EnablePartitioning:    tableflowv1.PtrBool(true),          // ready-only property that needs confirmation, assuming constantly true for now
				RetentionMs:           tableflowv1.PtrString("604800000"), // 7 days to miliseconds
				RecordFailureStrategy: tableflowv1.PtrString("SKIP"),
			},
			TableFormats: &[]string{"ICEBERG"},
			Environment:  &tableflowv1.GlobalObjectReference{Id: environmentId},
			KafkaCluster: &tableflowv1.EnvScopedObjectReference{Id: clusterId, Environment: tableflowv1.PtrString(environmentId)},
		},
		Status: &tableflowv1.TableflowV1TableflowTopicStatus{
			Phase: tableflowv1.PtrString("RUNNING"),
			//ErrorMessage: tableflowv1.PtrString(""),
			CatalogSyncStatuses: tableflowv1.PtrString(),
			FailingTableFormats: tableflowv1.PtrString(),
		},
	}
}

func getTopicManaged(display_name, environmentId, clusterId string) tableflowv1.TableflowV1TableflowTopic {
	return tableflowv1.TableflowV1TableflowTopic{
		Spec: &tableflowv1.TableflowV1TableflowTopicSpec{
			DisplayName: tableflowv1.PtrString(display_name),
			Suspended:   tableflowv1.PtrBool(false),
			Storage: &tableflowv1.TableflowV1TableflowTopicSpecStorageOneOf{
				TableflowV1ManagedStorageSpec: &tableflowv1.TableflowV1ManagedStorageSpec{
					Kind:      "Managed",
					TablePath: tableflowv1.PtrString("s3://dummy-bucket-name-1//10011010/11101100/org-1/env-2/lkc-3/v1/tableId"),
				},
			},
			Config: &tableflowv1.TableflowV1TableFlowTopicConfigsSpec{
				EnableCompaction:      tableflowv1.PtrBool(true),
				EnablePartitioning:    tableflowv1.PtrBool(true),          // ready-only property that needs confirmation, assuming constantly true for now
				RetentionMs:           tableflowv1.PtrString("604800000"), // 7 days to miliseconds
				RecordFailureStrategy: tableflowv1.PtrString("SUSPEND"),
			},
			TableFormats: &[]string{"DELTA"},
			Environment:  &tableflowv1.GlobalObjectReference{Id: environmentId},
			KafkaCluster: &tableflowv1.EnvScopedObjectReference{Id: clusterId},
		},
		Status: &tableflowv1.TableflowV1TableflowTopicStatus{
			Phase: tableflowv1.PtrString("RUNNING"),
			//ErrorMessage: tableflowv1.PtrString(""),
			CatalogSyncStatuses: tableflowv1.PtrString(),
			FailingTableFormats: tableflowv1.PtrString(),
		},
	}
}

func handleCatalogIntegrationGet(t *testing.T, environmentId, clusterId, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch id {
		case "tci-invalid":
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The catalog integration tci-invalid was not found.")
			require.NoError(t, err)
		case "tci-abc123":
			catalogIntegration := getCatalogIntegration(id, environmentId, clusterId, "my-aws-glue-ci", "AwsGlue")
			err := json.NewEncoder(w).Encode(catalogIntegration)
			require.NoError(t, err)
		case "tci-def456":
			catalogIntegration := getCatalogIntegration(id, environmentId, clusterId, "my-snowflake-ci", "Snowflake")
			err := json.NewEncoder(w).Encode(catalogIntegration)
			require.NoError(t, err)
		case "tci-abc456":
			catalogIntegration := getCatalogIntegration(id, environmentId, clusterId, "my-aws-glue-ci", "AwsGlue")
			err := json.NewEncoder(w).Encode(catalogIntegration)
			require.NoError(t, err)
		}
	}
}

func handleCatalogIntegrationDelete(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleCatalogIntegrationUpdate(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &tableflowv1.TableflowV1CatalogIntegrationUpdateRequest{}
		err := json.NewDecoder(r.Body).Decode(body)
		require.NoError(t, err)

		var catalogIntegration tableflowv1.TableflowV1CatalogIntegration
		switch id {
		case "tci-abc123":
			catalogIntegration = getCatalogIntegration(id, body.GetSpec().Environment.Id, body.GetSpec().KafkaCluster.Id, "my-aws-glue-ci", "AwsGlue")
		case "tci-def456":
			catalogIntegration = getCatalogIntegration(id, body.GetSpec().Environment.Id, body.GetSpec().KafkaCluster.Id, "my-snowflake-ci", "Snowflake")
			catalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeSpec.SetEndpoint(body.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeUpdateSpec.GetEndpoint())
			catalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeSpec.SetClientId(body.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeUpdateSpec.GetClientId())
			catalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeSpec.SetClientSecret(body.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeUpdateSpec.GetClientSecret())
			catalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeSpec.SetWarehouse(body.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeUpdateSpec.GetWarehouse())
			catalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeSpec.SetAllowedScope(body.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeUpdateSpec.GetAllowedScope())
		default:
			catalogIntegration = getCatalogIntegration(id, body.GetSpec().Environment.Id, body.GetSpec().KafkaCluster.Id, "my-aws-glue-ci", "AwsGlue")
		}

		if body.Spec.DisplayName != nil {
			catalogIntegration.Spec.SetDisplayName(body.Spec.GetDisplayName())
		}

		err = json.NewEncoder(w).Encode(catalogIntegration)
		require.NoError(t, err)
	}
}

func handleCatalogIntegrationList(t *testing.T, environment, clusterId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalogIntegrationOne := getCatalogIntegration("tci-abc123", environment, clusterId, "my-aws-glue-ci", "AwsGlue")
		catalogIntegrationTwo := getCatalogIntegration("tci-def456", environment, clusterId, "my-snowflake-ci", "Snowflake")
		catalogIntegrationTwo.Status.SetPhase("PENDING")

		recordList := tableflowv1.TableflowV1CatalogIntegrationList{Data: []tableflowv1.TableflowV1CatalogIntegration{catalogIntegrationOne, catalogIntegrationTwo}}
		setPageToken(&recordList, &recordList.Metadata, r.URL)
		err := json.NewEncoder(w).Encode(recordList)
		require.NoError(t, err)
	}
}

func handleCatalogIntegrationCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalogIntegration := &tableflowv1.TableflowV1CatalogIntegration{}
		err := json.NewDecoder(r.Body).Decode(catalogIntegration)
		require.NoError(t, err)

		catalogIntegration.SetId("tci-abc123")
		catalogIntegration.Status = &tableflowv1.TableflowV1CatalogIntegrationStatus{Phase: tableflowv1.PtrString("PENDING")}

		err = json.NewEncoder(w).Encode(catalogIntegration)
		require.NoError(t, err)
	}
}

func getCatalogIntegration(id, environment, cluster, name, specConfigKind string) tableflowv1.TableflowV1CatalogIntegration {
	catalogIntegration := tableflowv1.TableflowV1CatalogIntegration{
		Id: tableflowv1.PtrString(id),
		Spec: &tableflowv1.TableflowV1CatalogIntegrationSpec{
			DisplayName:  tableflowv1.PtrString(name),
			Suspended:    tableflowv1.PtrBool(false),
			Environment:  &tableflowv1.GlobalObjectReference{Id: environment},
			KafkaCluster: &tableflowv1.EnvScopedObjectReference{Id: cluster, Environment: tableflowv1.PtrString(environment)},
		},
		Status: &tableflowv1.TableflowV1CatalogIntegrationStatus{
			Phase:      tableflowv1.PtrString("CONNECTED"),
			LastSyncAt: tableflowv1.PtrString("2024-02-01T22:25:50.415274Z"),
		},
	}

	switch specConfigKind {
	case "AwsGlue":
		catalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationAwsGlueSpecAsTableflowV1CatalogIntegrationSpecConfigOneOf(&tableflowv1.TableflowV1CatalogIntegrationAwsGlueSpec{
			Kind:                  specConfigKind,
			ProviderIntegrationId: "cspi-stgce89r7",
		}))
	case "Snowflake":
		catalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationSnowflakeSpecAsTableflowV1CatalogIntegrationSpecConfigOneOf(&tableflowv1.TableflowV1CatalogIntegrationSnowflakeSpec{
			Kind:         specConfigKind,
			Endpoint:     "https://vuser1_polaris.snowflakecomputing.com/",
			ClientId:     "client-id",
			ClientSecret: "client-secret",
			Warehouse:    "warehouse",
			AllowedScope: "allowed-scope",
		}))
	}

	return catalogIntegration
}
