package test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/stretchr/testify/require"
)

/*
 * This fake server has: 3 nodes (replication-factor <= 3)
 * Only takes 2 config keys: retention.ms & compression.type
 * Contains 1 cluster: "cluster-1" and ListTopics return topic-1 and topic-2, only "topic-exist" exists for Create, Delete & Update
 */

func serveKafkaRest(t *testing.T) *httptest.Server {
	// Create a new testify/require object
	// req := require.New(t)
	router := http.NewServeMux()

	// Handler function names are based on URL (parsed)
	router.HandleFunc("/v3/clusters", getClustersHandler(t))
	router.HandleFunc("/v3/clusters/cluster-1/topics", getClustersClusterIdTopicsHandler(t))
	router.HandleFunc("/v3/clusters/cluster-1/topics/topic-exist", getClustersClusterIdTopicsTopicNameHandler(t, "topic-exist"))
	router.HandleFunc("/v3/clusters/cluster-1/topics/topic-not-exist", getClustersClusterIdTopicsTopicNameHandler(t, "topic-not-exist"))
	router.HandleFunc("/v3/clusters/cluster-1/topics/topic-exist/configs:alter", getClustersClusterIdTopicsTopicNameConfigsAlterHandler(t, "topic-exist"))
	router.HandleFunc("/v3/clusters/cluster-1/topics/topic-not-exist/configs:alter", getClustersClusterIdTopicsTopicNameConfigsAlterHandler(t, "topic-not-exist"))
	// Create a test server that routes each request to handler function described
	return httptest.NewServer(router)
}

func writeErrorResponse(responseWriter http.ResponseWriter, statusCode int, errorCode int, message string) error {
	responseWriter.WriteHeader(statusCode)
	responseBody := fmt.Sprintf(`{
		"error_code": %d,
		"message": "%s"
	}`, errorCode, message)
	_, err := io.WriteString(responseWriter, responseBody)
	return err
}

func getClustersHandler(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// List Clusters
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
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

func getClustersClusterIdTopicsHandler(t *testing.T) func(responseWriter http.ResponseWriter, r *http.Request) {
	return func(responseWriter http.ResponseWriter, r *http.Request) {
		// List Topics
		if r.Method == http.MethodGet {
			responseWriter.Header().Set("Content-Type", "application/json")
			response := `{
				"kind": "KafkaTopicList",
				"metadata": {"self": "http://localhost:9391/v3/clusters/cluster-1/topics","next": null},
				"data": [
				  {
					"kind": "KafkaTopic",
					"metadata": {"self": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-1","resource_name": "crn:///kafka=cluster-1/topic=topic-1"},
					"cluster_id": "cluster-1",
					"topic_name": "topic-1",
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
					"topic_name": "topic-2",
					"is_internal": true,
					"replication_factor": 4,
					"partitions": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-2/partitions"},
					"configs": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-2/configs"},
					"partition_reassignments": {"related": "http://localhost:9391/v3/clusters/cluster-1/topics/topic-2/partitions/-/reassignments"}
				  }
				]
			}`
			_, err := io.WriteString(responseWriter, response)
			require.NoError(t, err)
		} else if r.Method == http.MethodPost { // Create Topic
			// Parse Create Args
			reqBody, _ := ioutil.ReadAll(r.Body)
			var requestData kafkarestv3.CreateTopicRequestData
			_ = json.Unmarshal(reqBody, &requestData)

			// Process args
			if requestData.TopicName == "topic-exist" { // check topic
				require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, "Topic 'topic-exist' already exists."))
				return
			} else if requestData.PartitionsCount < -1 || requestData.PartitionsCount == 0 { // check partition
				require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, "Number of partitions must be larger than 0."))
				return
			} else if requestData.ReplicationFactor < -1 || requestData.ReplicationFactor == 0 { // check replication factor
				require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, "Replication factor must be larger than 0."))
				return
			} else if requestData.ReplicationFactor > 3 {
				require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, "Replication factor: 4 larger than available brokers: 3."))
				return
			}
			// check configs
			for _, config := range requestData.Configs {
				if config.Name != "retention.ms" && config.Name != "compression.type" {
					writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, fmt.Sprintf("Unknown topic config name: %s", config.Name))
					return
				} else if config.Name == "retention.ms" {
					if config.Value == nil { // if retention.ms but value null
						require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, "Null value not supported for topic configs : retention.ms"))
						return
					} else if _, err := strconv.Atoi(*config.Value); err != nil { // if retention.ms but value invalid
						require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, fmt.Sprintf("Invalid value %s for configuration retention.ms: Not a number of type LONG", *config.Value)))
						return
					}
				}
				// TODO: check for compression.type
			}
			// no errors = successfully created
			responseWriter.Header().Set("Content-Type", "application/json")
			responseWriter.WriteHeader(http.StatusCreated)
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
			_, err := io.WriteString(responseWriter, response)
			require.NoError(t, err)
		}
	}
}

func getClustersClusterIdTopicsTopicNameHandler(t *testing.T, topicName string) func(responseWriter http.ResponseWriter, request *http.Request) {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		// Delete topic
		// The server only houses the topic: "topic-exist"
		if request.Method == http.MethodDelete {
			if topicName == "topic-exist" {
				// successfully deleted
				responseWriter.WriteHeader(http.StatusNoContent)
			} else { // topic-not-exist
				// not found
				require.NoError(t, writeErrorResponse(responseWriter, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
			}
		}
	}
}

func getClustersClusterIdTopicsTopicNameConfigsAlterHandler(t *testing.T, topicName string) func(responseWriter http.ResponseWriter, request *http.Request) {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		// Batch alter topic configs
		if request.Method == http.MethodPost {
			if topicName == "topic-exist" { // successfully deleted
				// Parse Alter Args
				requestBody, err := ioutil.ReadAll(request.Body)
				require.NoError(t, err)
				var requestData kafkarestv3.AlterConfigBatchRequestData
				err = json.Unmarshal(requestBody, &requestData)
				require.NoError(t, err)

				// Check Alter Args if valid
				for _, config := range requestData.Data {
					if config.Name != "retention.ms" && config.Name != "compression.type" { // should be either retention.ms or compression.type
						require.NoError(t, writeErrorResponse(responseWriter, http.StatusNotFound, 404, fmt.Sprintf("Config %s cannot be found for TOPIC topic-exist in cluster cluster-1.", config.Name)))
						return
					} else if config.Name == "retention.ms" {
						if config.Value == nil { // if retention.ms but value null
							require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, "Null value not supported for : SET:retention.ms"))
							return
						} else if _, err := strconv.Atoi(*config.Value); err != nil { // if retention.ms but value invalid
							require.NoError(t, writeErrorResponse(responseWriter, http.StatusBadRequest, 40002, fmt.Sprintf("Invalid config value for resource ConfigResource(type=TOPIC, name='topic-exist'): Invalid value %s for configuration retention.ms: Not a number of type LONG", *config.Value)))
							return
						}
					}
					// TODO check for compression.type values
				}
				// No error
				responseWriter.WriteHeader(http.StatusNoContent)
			} else { // topic-not-exist
				// not found
				require.NoError(t, writeErrorResponse(responseWriter, http.StatusNotFound, 40403, "This server does not host this topic-partition."))
			}
		}
	}
}
