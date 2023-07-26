package testserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cckafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	cpkafkarestv3 "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

type route struct {
	path    string
	handler func(t *testing.T) http.HandlerFunc
}

var kafkaRestRoutes = []route{
	{"/kafka/v3/clusters", handleKafkaRPClusters},
	{"/kafka/v3/clusters/{cluster_id}/acls:batch", handleKafkaRPACLsBatch},
	{"/kafka/v3/clusters/{cluster_id}/broker-configs", handleKafkaBrokerConfigs},
	{"/kafka/v3/clusters/{cluster_id}/broker-configs/{name}", handleKafkaBrokerConfigsName},
	{"/kafka/v3/clusters/{cluster_id}/broker-configs:alter", handleKafkaBrokerConfigsAlter},
	{"/kafka/v3/clusters/{cluster_id}/brokers", handleKafkaBrokers},
	{"/kafka/v3/clusters/{cluster_id}/brokers/-/tasks", handleKafkaClustersClusterIdBrokersTasksGet},
	{"/kafka/v3/clusters/{cluster_id}/brokers/-/tasks/{task_type}", handleKafkaClustersClusterIdBrokersTasksTaskTypeGet},
	{"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}", handleKafkaBrokersBrokerId},
	{"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs", handleKafkaBrokerIdConfigs},
	{"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs/{name}", handleKafkaBrokerIdConfigsName},
	{"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs:alter", handleKafkaBrokerIdConfigsAlter},
	{"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks", handleKafkaClustersClusterIdBrokersBrokerIdTasksGet},
	{"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks/{task_type}", handleKafkaClustersClusterIdBrokersBrokerIdTasksTaskTypeGet},
	{"/kafka/v3/clusters/{cluster_id}/consumer-groups", handleKafkaRPConsumerGroups},
	{"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}", handleKafkaRPConsumerGroup},
	{"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/consumers", handleKafkaRPConsumers},
	{"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lag-summary", handleKafkaRPLagSummary},
	{"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags", handleKafkaRPLags},
	{"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags/{topic_name}/partitions/{partition_id}", handleKafkaRPLag},
	{"/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions", handleKafkaTopicPartitions},
	{"/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}", handleKafkaTopicPartitionId},
	{"/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}/reassignment", handleKafkaTopicPartitionIdReassignment},
	{"/kafka/v3/clusters/{cluster_id}/topics/{topic}/partitions/-/replica-status", handleKafkaRPReplicaStatus},
	{"/kafka/v3/clusters/{cluster}/acls", handleKafkaRPACLs},
	{"/kafka/v3/clusters/{cluster}/links", handleKafkaRPLinks},
	{"/kafka/v3/clusters/{cluster}/links/-/mirrors", handleKafkaRPAllMirrors},
	{"/kafka/v3/clusters/{cluster}/links/{link}", handleKafkaRPLink},
	{"/kafka/v3/clusters/{cluster}/links/{link}/configs", handleKafkaRPLinkConfigs},
	{"/kafka/v3/clusters/{cluster}/links/{link}/mirrors", handleKafkaRPMirrors},
	{"/kafka/v3/clusters/{cluster}/links/{link}/mirrors/{mirror_topic_name}", handleKafkaRPMirror},
	{"/kafka/v3/clusters/{cluster}/links/{link}/mirrors:promote", handleKafkaRPMirrorsPromote},
	{"/kafka/v3/clusters/{cluster}/topic/{topic}/partitions/-/replica-status", handleClustersClusterIdTopicsTopicsNamePartitionsReplicaStatus},
	{"/kafka/v3/clusters/{cluster}/topics", handleKafkaRPTopics},
	{"/kafka/v3/clusters/{cluster}/topics/{topic}", handleKafkaRPTopic},
	{"/kafka/v3/clusters/{cluster}/topics/{topic}/configs", handleKafkaRPTopicConfigs},
	{"/kafka/v3/clusters/{cluster}/topics/{topic}/configs:alter", handleKafkaRPConfigsAlter},
	{"/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replica-status", handleClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicaStatus},
	{"/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replicas", handleKafkaRPPartitionReplicas},
}

func NewKafkaRestProxyRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range kafkaRestRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}

// Handler for: "/kafka/v3/clusters"
func handleKafkaRPClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// List Clusters
		if r.Method == http.MethodGet {
			// Response jsons are created by editting examples from the Kafka REST Proxy OpenApi Spec
			response := `{
				"kind": "KafkaClusterList",
				"metadata": { "self": "http://localhost:8082/v3/clusters", "next": null},
				"data": [
				  {
					"kind": "KafkaCluster","metadata": {"self": "http://localhost:8082/v3/clusters/cluster-1","resource_name": "crn:///kafka=cluster-1"},
					"cluster_id": "cluster-1",
					"controller": {"related": "http://localhost:8082/v3/clusters/cluster-1/brokers/1"},
					"acls": {"related": "http://localhost:8082/v3/clusters/cluster-1/acls"},
					"brokers": {"related": "http://localhost:8082/v3/clusters/cluster-1/brokers"},
					"broker_configs": {"related": "http://localhost:8082/v3/clusters/cluster-1/broker-configs"},
					"consumer_groups": {"related": "http://localhost:8082/v3/clusters/cluster-1/consumer-groups"},
					"topics": {"related": "http://localhost:8082/v3/clusters/cluster-1/topics"},
					"partition_reassignments": {"related": "http://localhost:8082/v3/clusters/cluster-1/topics/-/partitions/-/reassignment"}
				  }
				]
			}`
			_, err := io.WriteString(w, response)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/acls"
func handleKafkaRPACLs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := []cckafkarestv3.AclData{{
			ResourceType: cckafkarestv3.TOPIC,
			ResourceName: "test-topic",
			Operation:    "READ",
			Permission:   "ALLOW",
			Host:         "*",
			Principal:    "User:sa-12345",
			PatternType:  "LITERAL",
		}}

		var res any

		switch r.Method {
		case http.MethodGet:
			res = cckafkarestv3.AclDataList{Data: data}
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			res = cckafkarestv3.AclData{}
		case http.MethodDelete:
			res = cckafkarestv3.InlineResponse200{Data: data}
		}

		err := json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/acls:batch"
func handleKafkaRPACLsBatch(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics"
func handleKafkaRPTopics(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			response := `{
				"kind": "KafkaTopicList",
				"metadata": {"self": "http://localhost:9391/v3/clusters/cluster-1/topics","next": null},
				"data": [
					{
					"kind": "KafkaTopic",
					"metadata": {"self": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1","resource_name": "crn:///kafka=cluster-1/topic=topic-1"},
					"cluster_id": "cluster-1",
					"topic_name": "topic1",
					"is_internal": false,
					"replication_factor": 3,
					"partitions": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1/partitions"},
					"configs": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1/configs"},
					"partition_reassignments": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1/partitions/-/reassignments"}
					},
					{
					"kind": "KafkaTopic",
					"metadata": {"self": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-2","resource_name": "crn:///kafka=cluster-1/topic=topic-2"},
					"cluster_id": "cluster-1",
					"topic_name": "topic2",
					"is_internal": true,
					"replication_factor": 4,
					"partitions": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-2/partitions"},
					"configs": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-2/configs"},
					"partition_reassignments": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-2/partitions/-/reassignments"}
					}
				]
			}`
			_, err := io.WriteString(w, response)
			require.NoError(t, err)
		case http.MethodPost:
			// Parse Create Args
			reqBody, _ := io.ReadAll(r.Body)
			var requestData cpkafkarestv3.CreateTopicRequestData
			err := json.Unmarshal(reqBody, &requestData)
			require.NoError(t, err)
			if requestData.TopicName == "topic-exist" { // check topic
				require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, "Topic 'topic-exist' already exists."))
				return
			} else if requestData.TopicName == "topic-exceed-limit" {
				require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, "Adding the requested number of partitions will exceed 9000 total partitions."))
				return
			} else if requestData.ReplicationFactor > 3 {
				require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, "Replication factor: 4 larger than available brokers: 3."))
				return
			}
			// check configs
			for _, config := range requestData.Configs {
				if config.Name != "retention.ms" && config.Name != "compression.type" {
					require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, fmt.Sprintf("Unknown topic config name: %s", config.Name)))
					return
				} else if config.Name == "retention.ms" {
					if config.Value == nil { // if retention.ms but value null
						require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, "Null value not supported for topic configs : retention.ms"))
						return
					} else if _, err := strconv.Atoi(*config.Value); err != nil { // if retention.ms but value invalid
						require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, fmt.Sprintf("Invalid value %s for configuration retention.ms: Not a number of type LONG", *config.Value)))
						return
					}
				}
				// TODO: check for compression.type
			}
			// no errors = successfully created
			w.WriteHeader(http.StatusCreated)
			response := fmt.Sprintf(`{
					"kind": "KafkaTopic",
					"metadata": {
					  "self": "http://localhost:9391/v3/clusters/cluster-1/topics/%[1]s",
					  "resource_name": "crn:///kafka=cluster-1/topic=%[1]s"
					},
					"cluster_id": "cluster-1",
					"topic_name": "%[1]s",
					"is_internal": false,
					"replication_factor": %[2]d,
					"partitions": {
					  "related": "http://localhost:9391/v3/clusters/cluster-1/topics/%[1]s/partitions"
					},
					"configs": {
					  "related": "http://localhost:9391/v3/clusters/cluster-1/topics/%[1]s/configs"
					},
					"partition_reassignments": {
					  "related": "http://localhost:9391/v3/clusters/cluster-1/topics/%[1]s/partitions/-/reassignments"
					}
				  }`, requestData.TopicName, requestData.ReplicationFactor)
			_, err = io.WriteString(w, response)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}/configs"
func handleKafkaRPTopicConfigs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topicName := vars["topic"]
		switch r.Method {
		case http.MethodGet:
			// if topic exists
			if topicName == "topic-exist" || topicName == "topic-exist-2" {
				topicConfigList := cpkafkarestv3.TopicConfigDataList{
					Data: []cpkafkarestv3.TopicConfigData{
						{
							Name:  "cleanup.policy",
							Value: cpkafkarestv3.PtrString("delete"),
						},
						{
							Name:       "compression.type",
							Value:      cpkafkarestv3.PtrString("producer"),
							IsReadOnly: true,
						},
						{
							Name:  "retention.ms",
							Value: cpkafkarestv3.PtrString("604800000"),
						},
					},
				}
				reply, err := json.Marshal(topicConfigList)
				require.NoError(t, err)
				_, err = io.WriteString(w, string(reply))
				require.NoError(t, err)
			} else if topicName == "topic-exist-rest" {
				topicConfigList := cpkafkarestv3.TopicConfigDataList{
					Data: []cpkafkarestv3.TopicConfigData{
						{
							Name:       "compression.type",
							Value:      cpkafkarestv3.PtrString("producer"),
							IsReadOnly: true,
						},
						{
							Name:  "retention.ms",
							Value: cpkafkarestv3.PtrString("1"),
						},
					},
				}
				reply, err := json.Marshal(topicConfigList)
				require.NoError(t, err)
				_, err = io.WriteString(w, string(reply))
				require.NoError(t, err)
			} else if topicName == "topic1" {
				topicConfigList := cckafkarestv3.TopicConfigDataList{
					Data: []cckafkarestv3.TopicConfigData{
						{
							Name:  "cleanup.policy",
							Value: *cckafkarestv3.NewNullableString(cckafkarestv3.PtrString("delete")),
						},
						{
							Name:  "delete.retention.ms",
							Value: *cckafkarestv3.NewNullableString(cckafkarestv3.PtrString("86400000")),
						},
					},
				}
				reply, err := json.Marshal(topicConfigList)
				require.NoError(t, err)
				_, err = io.WriteString(w, string(reply))
				require.NoError(t, err)
			} else { // if topic not exist
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/topics/{topic}/partitions/-/replica-status"
func handleKafkaRPReplicaStatus(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topic := vars["topic"]
		switch r.Method {
		case http.MethodGet:
			// if topic does not exist
			if topic != "topic-exist" {
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
				return
			}
			err := json.NewEncoder(w).Encode(cpkafkarestv3.ReplicaStatusDataList{
				Kind:     "",
				Metadata: cpkafkarestv3.ResourceCollectionMetadata{},
				Data: []cpkafkarestv3.ReplicaStatusData{
					{
						TopicName:   "topic-exist",
						BrokerId:    1001,
						PartitionId: 0,
						IsLeader:    true,
						IsInIsr:     true,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1002,
						PartitionId: 0,
						IsInIsr:     true,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1003,
						PartitionId: 0,
						IsInIsr:     true,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1001,
						PartitionId: 1,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1002,
						PartitionId: 1,
						IsLeader:    true,
						IsInIsr:     true,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1003,
						PartitionId: 1,
						IsInIsr:     true,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1001,
						PartitionId: 2,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1002,
						PartitionId: 2,
					},
					{
						TopicName:   "topic-exist",
						BrokerId:    1003,
						PartitionId: 2,
						IsLeader:    true,
						IsInIsr:     true,
					},
				},
			})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replicas"
func handleKafkaRPPartitionReplicas(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topicName := vars["topic"]
		partitionId := vars["partition"]
		switch r.Method {
		case http.MethodGet:
			// if topic exists
			if topicName == "topic-exist" {
				// Define replica & partition info
				type replicaData struct {
					brokerId int
					isLeader bool
					isInSync bool
				}
				partitionInfo := map[string]struct { // TODO: add test for different # of replicas for different partitions
					replicas []replicaData
				}{
					"0": {
						replicas: []replicaData{
							{brokerId: 1001, isLeader: true, isInSync: true},
							{brokerId: 1002, isLeader: false, isInSync: true},
							{brokerId: 1003, isLeader: false, isInSync: true},
						},
					},
					"1": {
						replicas: []replicaData{
							{brokerId: 1001, isLeader: false, isInSync: false},
							{brokerId: 1002, isLeader: true, isInSync: true},
							{brokerId: 1003, isLeader: false, isInSync: true},
						},
					},
					"2": {
						replicas: []replicaData{
							{brokerId: 1001, isLeader: false, isInSync: false},
							{brokerId: 1002, isLeader: false, isInSync: false},
							{brokerId: 1003, isLeader: true, isInSync: true},
						},
					},
				}

				// Build response string
				// Different sets of replica strings for different partitions
				replicaString := make([]string, len(partitionInfo[partitionId].replicas))
				for i := range partitionInfo[partitionId].replicas {
					replicaString[i] = fmt.Sprintf(`{
						"kind": "KafkaReplica",
						"metadata": {
							"self": "http://localhost:8082/v3/clusters/cluster-1/topics/%[1]s/partitions/%[2]s/replicas/1001",
							"resource_name": "crn:///kafka=cluster-1/topic=%[1]s/partition=%[2]s/replica=1001"
						},
						"cluster_id": "cluster-1",
						"topic_name": "%[1]s",
						"partition_id": %[2]s,
						"broker_id": %[3]d,
						"is_leader": %[4]t,
						"is_in_sync": %[5]t,
						"broker": {
							"related": "http://localhost:8082/v3/clusters/cluster-1/brokers/1001"
						}
					}`, "topic-exist", partitionId, partitionInfo[partitionId].replicas[i].brokerId,
						partitionInfo[partitionId].replicas[i].isLeader,
						partitionInfo[partitionId].replicas[i].isInSync)
				}
				responseString := fmt.Sprintf(`{
					"kind": "KafkaReplicaList",
					"metadata": {
						"self": "http://localhost:8082/v3/clusters/cluster-1/topics/%[1]s/partitions/%[2]s/replicas",
						"next": null
					},
					"data": [
						%[3]s,
						%[4]s,
						%[5]s
					]
				}`, "topic-exist", partitionId, replicaString[0], replicaString[1], replicaString[2])

				_, err := io.WriteString(w, responseString)
				require.NoError(t, err)
			} else { // if topic not exist
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}/configs:alter"
func handleKafkaRPConfigsAlter(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topicName := vars["topic"]
		switch r.Method {
		case http.MethodPost:
			if topicName == "topic-exist" || topicName == "topic-exist-rest" {
				// Parse Alter Args
				requestBody, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				var requestData cpkafkarestv3.AlterConfigBatchRequestData
				err = json.Unmarshal(requestBody, &requestData)
				require.NoError(t, err)

				// Check Alter Args if valid
				for _, config := range requestData.Data {
					if config.Name != "retention.ms" && config.Name != "compression.type" { // should be either retention.ms or compression.type
						require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 404, fmt.Sprintf("Config %s cannot be found for TOPIC topic-exist in cluster cluster-1.", config.Name)))
						return
					} else if config.Name == "retention.ms" {
						if config.Value == nil { // if retention.ms but value null
							require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, "Null value not supported for : SET:retention.ms"))
							return
						} else if _, err := strconv.Atoi(*config.Value); err != nil { // if retention.ms but value invalid
							require.NoError(t, writeErrorResponse(w, http.StatusBadRequest, 40002, fmt.Sprintf("Invalid config value for resource ConfigResource(type=TOPIC, name='topic-exist'): Invalid value %s for configuration retention.ms: Not a number of type LONG", *config.Value)))
							return
						}
					}
					// TODO check for compression.type values
				}
				// No error
				w.WriteHeader(http.StatusNoContent)
			} else if topicName == "topic1" {
				w.WriteHeader(http.StatusOK)
			} else { // topic-not-exist
				// not found
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}"
func handleKafkaRPTopic(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := mux.Vars(r)["topic"]
		if topic != "topic-exist" && topic != "topic-exist-2" && topic != "topic-exist-rest" {
			require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
			return
		}
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			data := cckafkarestv3.TopicData{PartitionsCount: 3}
			err := json.NewEncoder(w).Encode(data)
			require.NoError(t, err)
		case http.MethodPatch:
			data := cckafkarestv3.TopicData{PartitionsCount: 6}
			err := json.NewEncoder(w).Encode(data)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/links"
func handleKafkaRPLinks(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
			var req cpkafkarestv3.CreateKafkaLinkOpts
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
		case http.MethodGet:
			topics := make([]string, 2)
			topics = append(topics, "link-1-topic-1", "link-1-topic-2")
			cluster1 := cckafkarestv3.PtrString("cluster-1")
			cluster2 := cckafkarestv3.PtrString("cluster-2")
			linkStateAvailable := cckafkarestv3.PtrString("AVAILABLE")
			linkStateUnavailable := cckafkarestv3.PtrString("UNAVAILABLE")
			linkAuthErr := cckafkarestv3.PtrString("AUTHENTICATION_ERROR")
			noErrorErr := cckafkarestv3.PtrString("NO_ERROR")
			linkAuthErrMsg := cckafkarestv3.PtrString("Please check your API key and secret.")
			err := json.NewEncoder(w).Encode(cckafkarestv3.ListLinksResponseDataList{Data: []cckafkarestv3.ListLinksResponseData{
				{
					SourceClusterId:      *cckafkarestv3.NewNullableString(cluster1),
					DestinationClusterId: *cckafkarestv3.NewNullableString(cluster2),
					LinkName:             "link-1",
					ClusterLinkId:        "LINKID1",
					TopicNames:           topics,
					LinkError:            noErrorErr,
				},
				{
					SourceClusterId:      *cckafkarestv3.NewNullableString(cluster1),
					DestinationClusterId: *cckafkarestv3.NewNullableString(cluster2),
					LinkName:             "link-2",
					ClusterLinkId:        "LINKID2",
					TopicNames:           topics,
					LinkState:            linkStateAvailable,
					LinkError:            noErrorErr,
				},
				{
					SourceClusterId:      *cckafkarestv3.NewNullableString(cluster1),
					DestinationClusterId: *cckafkarestv3.NewNullableString(cluster2),
					LinkName:             "link-3",
					ClusterLinkId:        "LINKID3",
					TopicNames:           topics,
					LinkState:            linkStateUnavailable,
					LinkError:            linkAuthErr,
					LinkErrorMessage:     *cckafkarestv3.NewNullableString(linkAuthErrMsg),
				},
				{
					RemoteClusterId: *cckafkarestv3.NewNullableString(cluster2),
					LinkName:        "link-4",
					ClusterLinkId:   "LINKID4",
					TopicNames:      topics,
					LinkError:       noErrorErr,
				},
			}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/consumer-groups"
func handleKafkaRPConsumerGroups(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(cpkafkarestv3.ConsumerGroupDataList{
				Kind:     "",
				Metadata: cpkafkarestv3.ResourceCollectionMetadata{},
				Data: []cpkafkarestv3.ConsumerGroupData{
					{
						Kind:              "",
						Metadata:          cpkafkarestv3.ResourceMetadata{},
						ClusterId:         "cluster-1",
						ConsumerGroupId:   "consumer-group-1",
						IsSimple:          true,
						PartitionAssignor: "org.apache.kafka.clients.consumer.RoundRobinAssignor",
						State:             "STABLE",
						Coordinator:       cpkafkarestv3.Relationship{},
						Consumer:          cpkafkarestv3.Relationship{},
						LagSummary:        cpkafkarestv3.Relationship{},
					},
					{
						Kind:              "",
						Metadata:          cpkafkarestv3.ResourceMetadata{},
						ClusterId:         "cluster-1",
						ConsumerGroupId:   "consumer-group-2",
						IsSimple:          true,
						PartitionAssignor: "org.apache.kafka.clients.consumer.RoundRobinAssignor",
						State:             "DEAD",
						Coordinator:       cpkafkarestv3.Relationship{},
						Consumer:          cpkafkarestv3.Relationship{},
						LagSummary:        cpkafkarestv3.Relationship{},
					},
				},
			})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/links/{link}"
func handleKafkaRPLink(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cluster := mux.Vars(r)["cluster"]
		link := mux.Vars(r)["link"]
		switch r.Method {
		case http.MethodGet:
			if link == "link-1" {
				switch cluster {
				case "cluster-1":
					err := json.NewEncoder(w).Encode(cpkafkarestv3.ListLinksResponseData{
						Kind:                 "",
						Metadata:             cpkafkarestv3.ResourceMetadata{},
						DestinationClusterId: cpkafkarestv3.PtrString("cluster-2"),
						LinkName:             link,
						ClusterLinkId:        "LINKID1",
						TopicsNames:          []string{"link-1-topic-1", "link-1-topic-2"},
					})
					require.NoError(t, err)
				case "lkc-describe-topic":
					err := json.NewEncoder(w).Encode(cckafkarestv3.ListLinksResponseData{
						Kind:            "",
						Metadata:        cckafkarestv3.ResourceMetadata{},
						SourceClusterId: *cckafkarestv3.NewNullableString(cckafkarestv3.PtrString(cluster)),
						LinkName:        link,
						ClusterLinkId:   "LINKID1",
						TopicNames:      []string{"link-1-topic-1", "link-1-topic-2"},
						LinkState:       cckafkarestv3.PtrString("AVAILABLE"),
					})
					require.NoError(t, err)
				default:
					err := writeResourceNotFoundError(w)
					require.NoError(t, err)
				}
			} else if link == "link-3" {
				err := json.NewEncoder(w).Encode(cckafkarestv3.ListLinksResponseData{
					Kind:                 "",
					Metadata:             cckafkarestv3.ResourceMetadata{},
					SourceClusterId:      *cckafkarestv3.NewNullableString(cckafkarestv3.PtrString(cluster)),
					DestinationClusterId: *cckafkarestv3.NewNullableString(cckafkarestv3.PtrString("cluster-2")),
					LinkName:             link,
					ClusterLinkId:        "LINKID3",
					TopicNames:           []string{"link-1-topic-1", "link-1-topic-2"},
					LinkState:            cckafkarestv3.PtrString("UNAVAILABLE"),
					LinkError:            cckafkarestv3.PtrString("AUTHENTICATION_ERROR"),
					LinkErrorMessage:     *cckafkarestv3.NewNullableString(cckafkarestv3.PtrString("Please check your API key and secret.")),
				})
				require.NoError(t, err)
			} else if link == "link-4" {
				err := json.NewEncoder(w).Encode(cckafkarestv3.ListLinksResponseData{
					Kind:            "",
					Metadata:        cckafkarestv3.ResourceMetadata{},
					RemoteClusterId: *cckafkarestv3.NewNullableString(cckafkarestv3.PtrString("cluster-2")),
					LinkName:        link,
					ClusterLinkId:   "LINKID4",
					TopicNames:      []string{"link-1-topic-1", "link-1-topic-2"},
					LinkState:       cckafkarestv3.PtrString("AVAILABLE"),
				})
				require.NoError(t, err)
			}
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}"
func handleKafkaRPConsumerGroup(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		switch r.Method {
		case http.MethodGet:
			if vars["consumer_group_id"] == "consumer-group-1" {
				err := json.NewEncoder(w).Encode(cpkafkarestv3.ConsumerGroupData{
					Kind:              "",
					Metadata:          cpkafkarestv3.ResourceMetadata{},
					ClusterId:         "cluster-1",
					ConsumerGroupId:   "consumer-group-1",
					IsSimple:          true,
					PartitionAssignor: "RoundRobin",
					State:             "STABLE",
					Coordinator:       cpkafkarestv3.Relationship{Related: "/kafka/v3/clusters/cluster-1/brokers/broker-1"},
					Consumer:          cpkafkarestv3.Relationship{},
					LagSummary:        cpkafkarestv3.Relationship{},
				})
				require.NoError(t, err)
			} else {
				// group not found
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this consumer group."))
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/links/-/mirrors"
func handleKafkaRPAllMirrors(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
			var req cpkafkarestv3.ListKafkaMirrorTopicsOpts
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(cpkafkarestv3.ListMirrorTopicsResponseDataList{Data: []cpkafkarestv3.ListMirrorTopicsResponseData{
				{
					Kind:            "",
					Metadata:        cpkafkarestv3.ResourceMetadata{},
					LinkName:        "link-1",
					MirrorTopicName: "dest-topic-1",
					SourceTopicName: "src-topic-1",
					NumPartitions:   3,
					MirrorLags: []cpkafkarestv3.MirrorLag{
						{
							Partition:             0,
							Lag:                   142857,
							LastSourceFetchOffset: 1293009,
						},
						{
							Partition:             1,
							Lag:                   285714,
							LastSourceFetchOffset: 28340404,
						},
						{
							Partition:             2,
							Lag:                   571428,
							LastSourceFetchOffset: 5739304,
						},
					},
					MirrorStatus: "active",
					StateTimeMs:  111111111,
				},
				{
					Kind:            "",
					Metadata:        cpkafkarestv3.ResourceMetadata{},
					LinkName:        "link-2",
					MirrorTopicName: "dest-topic-2",
					SourceTopicName: "src-topic-2",
					NumPartitions:   2,
					MirrorLags: []cpkafkarestv3.MirrorLag{
						{
							Partition:             0,
							Lag:                   0,
							LastSourceFetchOffset: 0,
						},
						{
							Partition:             1,
							Lag:                   0,
							LastSourceFetchOffset: 0,
						},
					},
					MirrorStatus: "stopped",
					StateTimeMs:  222222222,
				},
			}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/consumers"
func handleKafkaRPConsumers(t *testing.T) http.HandlerFunc {
	instance1 := "instance-1"
	instance2 := "instance-2"
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(cpkafkarestv3.ConsumerDataList{
				Kind:     "",
				Metadata: cpkafkarestv3.ResourceCollectionMetadata{},
				Data: []cpkafkarestv3.ConsumerData{
					{
						Kind:            "",
						Metadata:        cpkafkarestv3.ResourceMetadata{},
						ClusterId:       "cluster-1",
						ConsumerGroupId: "consumer-group-1",
						ConsumerId:      "consumer-1",
						InstanceId:      &instance1,
						ClientId:        "client-1",
						Assignments:     cpkafkarestv3.Relationship{},
					},
					{
						Kind:            "",
						Metadata:        cpkafkarestv3.ResourceMetadata{},
						ClusterId:       "cluster-1",
						ConsumerGroupId: "consumer-group-1",
						ConsumerId:      "consumer-2",
						InstanceId:      &instance2,
						ClientId:        "client-2",
						Assignments:     cpkafkarestv3.Relationship{},
					},
				},
			})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/links/{link_name}/mirrors"
func handleKafkaRPMirrors(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
			var req cpkafkarestv3.ListKafkaMirrorTopicsUnderLinkOpts
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(cpkafkarestv3.ListMirrorTopicsResponseDataList{Data: []cpkafkarestv3.ListMirrorTopicsResponseData{
				{
					Kind:            "",
					Metadata:        cpkafkarestv3.ResourceMetadata{},
					LinkName:        "link-1",
					MirrorTopicName: "dest-topic-1",
					SourceTopicName: "src-topic-1",
					NumPartitions:   3,
					MirrorLags: []cpkafkarestv3.MirrorLag{
						{
							Partition:             0,
							Lag:                   142857,
							LastSourceFetchOffset: 1293009,
						},
						{
							Partition:             1,
							Lag:                   285714,
							LastSourceFetchOffset: 28340404,
						},
						{
							Partition:             2,
							Lag:                   571428,
							LastSourceFetchOffset: 5739304,
						},
					},
					MirrorStatus: "active",
					StateTimeMs:  111111111,
				},
				{
					Kind:            "",
					Metadata:        cpkafkarestv3.ResourceMetadata{},
					LinkName:        "link-2",
					MirrorTopicName: "dest-topic-2",
					SourceTopicName: "src-topic-2",
					NumPartitions:   2,
					MirrorLags: []cpkafkarestv3.MirrorLag{
						{
							Partition:             0,
							Lag:                   0,
							LastSourceFetchOffset: 0,
						},
						{
							Partition:             1,
							Lag:                   0,
							LastSourceFetchOffset: 0,
						},
					},
					MirrorStatus: "stopped",
					StateTimeMs:  222222222,
				},
			}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lag-summary"
func handleKafkaRPLagSummary(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		switch r.Method {
		case http.MethodGet:
			if vars["consumer_group_id"] == "consumer-group-1" {
				instance := "instance-1"
				err := json.NewEncoder(w).Encode(cpkafkarestv3.ConsumerGroupLagSummaryData{
					Kind:              "",
					Metadata:          cpkafkarestv3.ResourceMetadata{},
					ClusterId:         "cluster-1",
					ConsumerGroupId:   "consumer-group-1",
					MaxLagConsumerId:  "consumer-1",
					MaxLagInstanceId:  &instance,
					MaxLagClientId:    "client-1",
					MaxLagTopicName:   "topic-1",
					MaxLagPartitionId: 1,
					MaxLag:            100,
					TotalLag:          110,
					MaxLagConsumer:    cpkafkarestv3.Relationship{},
					MaxLagPartition:   cpkafkarestv3.Relationship{},
				})
				require.NoError(t, err)
			} else {
				// group not found
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this consumer group."))
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/links/{link_name}/mirrors:promote"
func handleKafkaRPMirrorsPromote(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			errorMsg := "Not authorized"
			var errorCode int32 = 401
			err := json.NewEncoder(w).Encode(cpkafkarestv3.AlterMirrorStatusResponseDataList{Data: []cpkafkarestv3.AlterMirrorStatusResponseData{
				{
					Kind:            "",
					Metadata:        cpkafkarestv3.ResourceMetadata{},
					MirrorTopicName: "dest-topic-1",
					ErrorMessage:    nil,
					ErrorCode:       nil,
					MirrorLags: []cpkafkarestv3.MirrorLag{
						{
							Partition:             0,
							Lag:                   142857,
							LastSourceFetchOffset: 1293009,
						},
						{
							Partition:             1,
							Lag:                   285714,
							LastSourceFetchOffset: 28340404,
						},
						{
							Partition:             2,
							Lag:                   571428,
							LastSourceFetchOffset: 5739304,
						},
					},
				},
				{
					Kind:            "",
					Metadata:        cpkafkarestv3.ResourceMetadata{},
					MirrorTopicName: "dest-topic-1",
					ErrorMessage:    &errorMsg,
					ErrorCode:       &errorCode,
					MirrorLags: []cpkafkarestv3.MirrorLag{
						{
							Partition:             0,
							Lag:                   142857,
							LastSourceFetchOffset: 1293009,
						},
						{
							Partition:             1,
							Lag:                   285714,
							LastSourceFetchOffset: 28340404,
						},
						{
							Partition:             2,
							Lag:                   571428,
							LastSourceFetchOffset: 5739304,
						},
					},
				},
			}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags"
func handleKafkaRPLags(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		switch r.Method {
		case http.MethodGet:
			if vars["consumer_group_id"] == "consumer-group-1" {
				instance1 := "instance-1"
				instance2 := "instance-2"
				err := json.NewEncoder(w).Encode(cpkafkarestv3.ConsumerLagDataList{
					Kind:     "",
					Metadata: cpkafkarestv3.ResourceCollectionMetadata{},
					Data: []cpkafkarestv3.ConsumerLagData{
						{
							Kind:            "",
							Metadata:        cpkafkarestv3.ResourceMetadata{},
							ClusterId:       "cluster-1",
							ConsumerGroupId: "consumer-group-1",
							TopicName:       "topic-1",
							PartitionId:     1,
							CurrentOffset:   1,
							LogEndOffset:    101,
							Lag:             100,
							ConsumerId:      "consumer-1",
							InstanceId:      &instance1,
							ClientId:        "client-1",
						},
						{
							Kind:            "",
							Metadata:        cpkafkarestv3.ResourceMetadata{},
							ClusterId:       "cluster-1",
							ConsumerGroupId: "consumer-group-1",
							TopicName:       "topic-1",
							PartitionId:     2,
							CurrentOffset:   1,
							LogEndOffset:    11,
							Lag:             10,
							ConsumerId:      "consumer-2",
							InstanceId:      &instance2,
							ClientId:        "client-2",
						},
					},
				})
				require.NoError(t, err)
			} else {
				// group not found
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this consumer group."))
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/links/{link_name}/configs"
func handleKafkaRPLinkConfigs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		link := mux.Vars(r)["link"]
		if link == "link-dne" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodGet:
			var linkMode string
			if link == "link-4" {
				linkMode = "BIDIRECTIONAL"
			} else {
				linkMode = "DESTINATION"
			}
			err := json.NewEncoder(w).Encode(cpkafkarestv3.ListLinkConfigsResponseDataList{Data: []cpkafkarestv3.ListLinkConfigsResponseData{
				{
					Kind:      "",
					Metadata:  cpkafkarestv3.ResourceMetadata{},
					ClusterId: "cluster-1",
					Name:      "replica.fetch.max.bytes",
					Value:     "1048576",
					ReadOnly:  false,
					Sensitive: false,
					Source:    "source-1",
					Synonyms:  []string{"rfmb", "bmfr"},
					LinkName:  link,
				},
				{
					Kind:      "",
					Metadata:  cpkafkarestv3.ResourceMetadata{},
					ClusterId: "cluster-1",
					Name:      "bootstrap.servers",
					Value:     "bitcoin.com:8888",
					ReadOnly:  false,
					Sensitive: false,
					Source:    "source-2",
					Synonyms:  nil,
					LinkName:  link,
				},
				{
					Kind:      "",
					Metadata:  cpkafkarestv3.ResourceMetadata{},
					ClusterId: "cluster-1",
					Name:      "link.mode",
					Value:     linkMode,
					ReadOnly:  false,
					Sensitive: false,
					Source:    "source-2",
					Synonyms:  nil,
					LinkName:  link,
				},
			}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/links/{link_name}/mirrors/{mirror_name}"
func handleKafkaRPMirror(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(cpkafkarestv3.ListMirrorTopicsResponseData{
				Kind:            "",
				Metadata:        cpkafkarestv3.ResourceMetadata{},
				LinkName:        "link-1",
				MirrorTopicName: "dest-topic-1",
				SourceTopicName: "src-topic-1",
				NumPartitions:   3,
				MirrorLags: []cpkafkarestv3.MirrorLag{
					{
						Partition:             0,
						Lag:                   142857,
						LastSourceFetchOffset: 1293009,
					},
					{
						Partition:             1,
						Lag:                   285714,
						LastSourceFetchOffset: 28340404,
					},
					{
						Partition:             2,
						Lag:                   571428,
						LastSourceFetchOffset: 5739304,
					},
				},
				MirrorStatus: "active",
				StateTimeMs:  111111111,
			})
			require.NoError(t, err)
		}
	}
}

type partitionOffsets struct {
	currentOffset int64
	logEndOffset  int64
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags/{topic_name}/partitions/{partition_id}"
func handleKafkaRPLag(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		switch r.Method {
		case http.MethodGet:
			if vars["consumer_group_id"] == "consumer-group-1" {
				partitionOffsetsMap := map[string]partitionOffsets{
					"0": {101, 101},
					"1": {1, 101},
					"2": {101, 101},
				}
				requestedPartition := vars["partition_id"]
				offsets := partitionOffsetsMap[requestedPartition]
				if vars["topic_name"] == "topic-1" && offsets != (partitionOffsets{}) {
					instance := "instance-1"
					partitionId, _ := strconv.Atoi(requestedPartition)
					err := json.NewEncoder(w).Encode(cpkafkarestv3.ConsumerLagData{
						Kind:            "",
						Metadata:        cpkafkarestv3.ResourceMetadata{},
						ClusterId:       "cluster-1",
						ConsumerGroupId: "consumer-group-1",
						TopicName:       "topic-1",
						PartitionId:     int32(partitionId),
						CurrentOffset:   offsets.currentOffset,
						LogEndOffset:    offsets.logEndOffset,
						Lag:             offsets.logEndOffset - offsets.currentOffset,
						ConsumerId:      "consumer-1",
						InstanceId:      &instance,
						ClientId:        "client-1",
					})
					require.NoError(t, err)
				} else {
					// topic and/or partition not found
					require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
				}
			} else {
				// group not found
				require.NoError(t, writeErrorResponse(w, http.StatusNotFound, 40403, "This server does not host this consumer group."))
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions"
func handleKafkaTopicPartitions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(cpkafkarestv3.PartitionDataList{
				Data: []cpkafkarestv3.PartitionData{
					{
						ClusterId:   vars["cluster_id"],
						PartitionId: 0,
						TopicName:   vars["topic_name"],
						Leader:      cpkafkarestv3.Relationship{Related: "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1/partition/2"},
					},
					{
						ClusterId:   vars["cluster_id"],
						PartitionId: 1,
						TopicName:   vars["topic_name"],
						Leader:      cpkafkarestv3.Relationship{Related: "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1/partition/1"},
					},
					{
						ClusterId:   vars["cluster_id"],
						PartitionId: 2,
						TopicName:   vars["topic_name"],
						Leader:      cpkafkarestv3.Relationship{Related: "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1/partition/0"},
					},
				},
			})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}"
func handleKafkaTopicPartitionId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		partitionIdStr := vars["partition_id"]
		partitionId, err := strconv.ParseInt(partitionIdStr, 10, 32)
		require.NoError(t, err)
		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(cpkafkarestv3.PartitionData{
				ClusterId:   vars["cluster_id"],
				PartitionId: int32(partitionId),
				TopicName:   vars["topic_name"],
				Leader:      cpkafkarestv3.Relationship{Related: "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1/partition/2"},
			})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers"
func handleKafkaBrokers(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		host1 := "kafka1"
		port1 := int32(1)
		host2 := "kafka2"
		port2 := int32(2)
		err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerDataList{
			Data: []cpkafkarestv3.BrokerData{
				{
					ClusterId: vars["cluster_id"],
					BrokerId:  1,
					Port:      &port1,
					Host:      &host1,
				},
				{
					ClusterId: vars["cluster_id"],
					BrokerId:  2,
					Port:      &port2,
					Host:      &host2,
				},
			},
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/broker-configs/{name}"
func handleKafkaBrokerConfigsName(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		configValue := "gzip"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.ClusterConfigData{
			Name:        vars["name"],
			Value:       &configValue,
			IsSensitive: false,
			IsReadOnly:  false,
			IsDefault:   false,
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/broker-configs"
func handleKafkaBrokerConfigs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		configValue1 := "gzip"
		configValue2 := "SASL/PLAIN"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.ClusterConfigDataList{
			Data: []cpkafkarestv3.ClusterConfigData{
				{
					ClusterId:   vars["cluster_id"],
					Name:        "compression.type",
					Value:       &configValue1,
					IsDefault:   true,
					IsReadOnly:  true,
					IsSensitive: true,
				},
				{
					ClusterId:   vars["cluster_id"],
					Name:        "sasl_mechanism",
					Value:       &configValue2,
					IsDefault:   false,
					IsReadOnly:  false,
					IsSensitive: false,
				},
			},
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs/{name}"
func handleKafkaBrokerIdConfigsName(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		configValue1 := "gzip"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerConfigData{
			ClusterId:   vars["cluster_id"],
			Name:        vars["name"],
			Value:       &configValue1,
			IsDefault:   true,
			IsReadOnly:  true,
			IsSensitive: true,
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs"
func handleKafkaBrokerIdConfigs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		brokerId := vars["broker_id"]
		if brokerId != "1" && brokerId != "2" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		configValue1 := "gzip"
		configValue2 := "SASL/PLAIN"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerConfigDataList{
			Data: []cpkafkarestv3.BrokerConfigData{
				{
					ClusterId:   vars["cluster_id"],
					Name:        "compression.type",
					Value:       &configValue1,
					IsDefault:   true,
					IsReadOnly:  true,
					IsSensitive: true,
				},
				{
					ClusterId:   vars["cluster_id"],
					Name:        "sasl_mechanism",
					Value:       &configValue2,
					IsDefault:   false,
					IsReadOnly:  false,
					IsSensitive: false,
				},
			},
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/broker-configs:alter"
func handleKafkaBrokerConfigsAlter(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		// var req cpkafkarestv3.UpdateKafkaClusterConfigsOpts
		// err := json.NewDecoder(r.Body).Decode(&req)
		// require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs:alter"
func handleKafkaBrokerIdConfigsAlter(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		// var req cpkafkarestv3.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts
		// err := json.NewDecoder(r.Body).Decode(&req)
		// require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}"
func handleKafkaBrokersBrokerId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		switch r.Method {
		case http.MethodDelete:
			var req cpkafkarestv3.ClustersClusterIdBrokersBrokerIdDeleteOpts
			_ = json.NewDecoder(r.Body).Decode(&req)
			err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerRemovalData{
				ClusterId:  vars["cluster_id"],
				BrokerId:   1,
				BrokerTask: cpkafkarestv3.Relationship{Related: "http://localhost:9391/kafka/v3/clusters/cluster-1/brokers/1/tasks/remove-broker"},
			})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}/reassignment"
func handleKafkaTopicPartitionIdReassignment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		partitionIdStr := vars["partition_id"]
		topicName := vars["topic_name"]
		switch r.Method {
		case http.MethodGet:
			if partitionIdStr != "-" && topicName != "-" {
				partitionId, err := strconv.ParseInt(partitionIdStr, 10, 32)
				require.NoError(t, err)
				err = json.NewEncoder(w).Encode(cpkafkarestv3.ReassignmentData{
					Kind:             "ReassignmentData",
					ClusterId:        vars["cluster_id"],
					PartitionId:      int32(partitionId),
					TopicName:        vars["topic_name"],
					AddingReplicas:   []int32{1, 2, 3},
					RemovingReplicas: []int32{4},
				})
				require.NoError(t, err)
			} else if topicName != "-" {
				err := json.NewEncoder(w).Encode(cpkafkarestv3.ReassignmentDataList{
					Data: []cpkafkarestv3.ReassignmentData{
						{
							ClusterId:        vars["cluster_id"],
							PartitionId:      0,
							TopicName:        vars["topic_name"],
							AddingReplicas:   []int32{1, 2, 3},
							RemovingReplicas: []int32{4},
						},
						{
							ClusterId:        vars["cluster_id"],
							PartitionId:      1,
							TopicName:        vars["topic_name"],
							AddingReplicas:   []int32{4},
							RemovingReplicas: []int32{1, 2, 3},
						},
					},
				})
				require.NoError(t, err)
			} else if partitionIdStr == "-" && topicName == "-" {
				err := json.NewEncoder(w).Encode(cpkafkarestv3.ReassignmentDataList{
					Data: []cpkafkarestv3.ReassignmentData{
						{
							ClusterId:        vars["cluster_id"],
							PartitionId:      0,
							TopicName:        "topic1",
							AddingReplicas:   []int32{1, 2, 3},
							RemovingReplicas: []int32{4},
						},
						{
							ClusterId:        vars["cluster_id"],
							PartitionId:      1,
							TopicName:        "topic1",
							AddingReplicas:   []int32{4},
							RemovingReplicas: []int32{1, 2, 3},
						},
						{
							ClusterId:        vars["cluster_id"],
							PartitionId:      0,
							TopicName:        "topic2",
							AddingReplicas:   []int32{1, 2, 3},
							RemovingReplicas: []int32{4},
						},
						{
							ClusterId:        vars["cluster_id"],
							PartitionId:      1,
							TopicName:        "topic2",
							AddingReplicas:   []int32{4},
							RemovingReplicas: []int32{1, 2, 3},
						},
					},
				})
				require.NoError(t, err)
			}
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/-/tasks/{task_type}"
func handleKafkaClustersClusterIdBrokersTasksTaskTypeGet(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		errorCode := int32(10014)
		errorMessage := "error message"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerTaskDataList{
			Data: []cpkafkarestv3.BrokerTaskData{
				{
					ClusterId:       vars["cluster_id"],
					BrokerId:        1,
					TaskType:        cpkafkarestv3.BrokerTaskType(vars["task_type"]),
					TaskStatus:      "SUCCESS",
					SubTaskStatuses: map[string]string{"partition_reassignment_status": "IN_PROGRESS"},
					CreatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					UpdatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
				},
				{
					ClusterId:       vars["cluster_id"],
					BrokerId:        2,
					TaskType:        cpkafkarestv3.BrokerTaskType(vars["task_type"]),
					TaskStatus:      "SUCCESS",
					SubTaskStatuses: map[string]string{"broker_shutdown_status": "COMPLETED"},
					CreatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					UpdatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					ErrorCode:       &errorCode,
					ErrorMessage:    &errorMessage,
				},
			},
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/-/tasks"
func handleKafkaClustersClusterIdBrokersTasksGet(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		errorCode := int32(10014)
		errorMessage := "error message"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerTaskDataList{
			Data: []cpkafkarestv3.BrokerTaskData{
				{
					ClusterId:       vars["cluster_id"],
					BrokerId:        1,
					TaskType:        cpkafkarestv3.BROKERTASKTYPE_REMOVE_BROKER,
					TaskStatus:      "SUCCESS",
					SubTaskStatuses: map[string]string{"partition_reassignment_status": "IN_PROGRESS"},
					CreatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					UpdatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
				},
				{
					ClusterId:       vars["cluster_id"],
					BrokerId:        2,
					TaskType:        cpkafkarestv3.BROKERTASKTYPE_ADD_BROKER,
					TaskStatus:      "SUCCESS",
					SubTaskStatuses: map[string]string{"partition_reassignment_status": "IN_PROGRESS"},
					CreatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					UpdatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					ErrorCode:       &errorCode,
					ErrorMessage:    &errorMessage,
				},
			},
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks/{task_type}"
func handleKafkaClustersClusterIdBrokersBrokerIdTasksTaskTypeGet(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		errorCode := int32(10014)
		errorMessage := "error message"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerTaskData{
			ClusterId:       vars["cluster_id"],
			BrokerId:        1,
			TaskType:        cpkafkarestv3.BrokerTaskType(vars["task_type"]),
			TaskStatus:      "SUCCESS",
			SubTaskStatuses: map[string]string{"partition_reassignment_status": "IN_PROGRESS"},
			CreatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
			UpdatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
			ErrorMessage:    &errorMessage,
			ErrorCode:       &errorCode,
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks"
func handleKafkaClustersClusterIdBrokersBrokerIdTasksGet(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		errorCode := int32(10014)
		errorMessage := "error message"
		err := json.NewEncoder(w).Encode(cpkafkarestv3.BrokerTaskDataList{
			Data: []cpkafkarestv3.BrokerTaskData{
				{
					ClusterId:       vars["cluster_id"],
					BrokerId:        1,
					TaskType:        cpkafkarestv3.BROKERTASKTYPE_REMOVE_BROKER,
					TaskStatus:      "SUCCESS",
					SubTaskStatuses: map[string]string{"partition_reassignment_status": "IN_PROGRESS"},
					CreatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					UpdatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
				},
				{
					ClusterId:       vars["cluster_id"],
					BrokerId:        1,
					TaskType:        cpkafkarestv3.BROKERTASKTYPE_ADD_BROKER,
					TaskStatus:      "SUCCESS",
					SubTaskStatuses: map[string]string{"partition_reassignment_status": "IN_PROGRESS"},
					CreatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					UpdatedAt:       time.Date(2021, 7, 1, 0, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60)),
					ErrorCode:       &errorCode,
					ErrorMessage:    &errorMessage,
				},
			},
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replica-status"
func handleClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicaStatus(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		partitionId, err := strconv.ParseInt(vars["partition"], 10, 32)
		require.NoError(t, err)
		err = json.NewEncoder(w).Encode(cpkafkarestv3.ReplicaStatusDataList{
			Data: []cpkafkarestv3.ReplicaStatusData{
				{
					ClusterId:          vars["cluster"],
					TopicName:          vars["topic"],
					PartitionId:        int32(partitionId),
					BrokerId:           1,
					IsLeader:           true,
					IsCaughtUp:         true,
					IsInIsr:            true,
					IsIsrEligible:      true,
					LastFetchTimeMs:    1640040996303,
					LastCaughtUpTimeMs: 1640040996303,
					LogStartOffset:     0,
					LogEndOffset:       1,
				},
				{
					ClusterId:          vars["cluster"],
					TopicName:          vars["topic"],
					PartitionId:        int32(partitionId),
					BrokerId:           2,
					IsLeader:           false,
					IsCaughtUp:         true,
					IsInIsr:            true,
					IsIsrEligible:      true,
					LastFetchTimeMs:    1640040996303,
					LastCaughtUpTimeMs: 1640040996303,
					LogStartOffset:     0,
					LogEndOffset:       0,
					LinkName:           "link",
				},
			},
		})
		require.NoError(t, err)
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topic/{topic}/partitions/-/replica-status"
func handleClustersClusterIdTopicsTopicsNamePartitionsReplicaStatus(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		err := json.NewEncoder(w).Encode(cpkafkarestv3.ReplicaStatusDataList{
			Data: []cpkafkarestv3.ReplicaStatusData{
				{
					ClusterId:          vars["cluster"],
					TopicName:          vars["topic"],
					PartitionId:        1,
					BrokerId:           1,
					IsLeader:           true,
					IsCaughtUp:         true,
					IsInIsr:            true,
					IsIsrEligible:      true,
					LastFetchTimeMs:    1640040996303,
					LastCaughtUpTimeMs: 1640040996303,
					LogStartOffset:     0,
					LogEndOffset:       1,
				},
				{
					ClusterId:          vars["cluster"],
					TopicName:          vars["topic"],
					PartitionId:        2,
					BrokerId:           2,
					IsLeader:           false,
					IsCaughtUp:         true,
					IsInIsr:            true,
					IsIsrEligible:      true,
					LastFetchTimeMs:    1640040996303,
					LastCaughtUpTimeMs: 1640040996303,
					LogStartOffset:     0,
					LogEndOffset:       0,
					LinkName:           "link",
				},
			},
		})
		require.NoError(t, err)
	}
}

func writeErrorResponse(responseWriter http.ResponseWriter, statusCode int, errorCode int, message string) error {
	responseWriter.WriteHeader(statusCode)
	errorResponseBody := fmt.Sprintf(`{
		"error_code": %d,
		"message": "%s"
	}`, errorCode, message)
	_, err := io.WriteString(responseWriter, errorResponseBody)
	return err
}
