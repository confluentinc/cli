package test_server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for: "/kafka/v3/clusters/{cluster}/acls"
func (r KafkaRestProxyRouter) HandleKafkaRPACLs(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			vars := mux.Vars(r)
			err := json.NewEncoder(w).Encode(kafkarestv3.AclDataList{Data: []kafkarestv3.AclData{{
				Kind:         "",
				Metadata:     kafkarestv3.ResourceMetadata{},
				ClusterId:    vars["cluster"],
				ResourceType: "TOPIC",
				ResourceName: "test-rest-proxy-topic",
				Operation:    "READ",
				Permission:   "ALLOW",
			}}})
			require.NoError(t, err)
		case "POST":
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			var req kafkarestv3.ClustersClusterIdAclsPostOpts
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(kafkarestv3.ClustersClusterIdAclsPostOpts{})
			require.NoError(t, err)
		case "DELETE":
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(kafkarestv3.InlineResponse200{Data: []kafkarestv3.AclData{{ResourceName: "topic-1"}}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics"
func (r KafkaRestProxyRouter) HandleKafkaRPTopics(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			vars := mux.Vars(r)
			err := json.NewEncoder(w).Encode(kafkarestv3.TopicDataList{Data: []kafkarestv3.TopicData{{
				Kind:              "",
				Metadata:          kafkarestv3.ResourceMetadata{},
				ClusterId:         vars["cluster"],
				TopicName:         "rest-proxy-topic",
				ReplicationFactor: int32(1),
				Partitions:        kafkarestv3.Relationship{Related: "relationship"},
			}}})
			require.NoError(t, err)
		case "POST":
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			var req kafkarestv3.ClustersClusterIdTopicsPostOpts
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions"
func (r KafkaRestProxyRouter) HandleKafkaRPPartitions(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			vars := mux.Vars(r)
			err := json.NewEncoder(w).Encode(kafkarestv3.PartitionDataList{Data: []kafkarestv3.PartitionData{{
				Kind:        "",
				Metadata:    kafkarestv3.ResourceMetadata{},
				ClusterId:   vars["cluster"],
				TopicName:   vars["topic"],
				PartitionId: int32(2),
				Leader:      kafkarestv3.Relationship{Related: "leader"},
				Replicas:    kafkarestv3.Relationship{Related: "replica"},
			}}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}/configs"
func (r KafkaRestProxyRouter) HandleKafkaRPTopicConfigs(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			vars := mux.Vars(r)
			configVal := "config-value"
			err := json.NewEncoder(w).Encode(kafkarestv3.TopicConfigDataList{Data: []kafkarestv3.TopicConfigData{{
				Kind:      "",
				Metadata:  kafkarestv3.ResourceMetadata{},
				ClusterId: vars["cluster"],
				TopicName: vars["topic"],
				Name:      "test-config",
				Value:     &configVal,
			}}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replicas"
func (r KafkaRestProxyRouter) HandleKafkaRPPartitionReplicas(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			vars := mux.Vars(r)
			err := json.NewEncoder(w).Encode(kafkarestv3.ReplicaDataList{Data: []kafkarestv3.ReplicaData{{
				Kind:      "",
				Metadata:  kafkarestv3.ResourceMetadata{},
				ClusterId: vars["cluster"],
				TopicName: vars["topic"],
				IsLeader:  true,
			}}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/configs:alter"
func (r KafkaRestProxyRouter) HandleKafkaRPConfigsAlter(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.WriteHeader(http.StatusNoContent)
			w.Header().Set("Content-Type", "application/json")
			var req kafkarestv3.ClustersClusterIdTopicsTopicNameConfigsalterPostOpts
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/kafka/v3/clusters/{cluster}/topics/{topic}"
func (r KafkaRestProxyRouter) HandlKafkaRPTopic(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
			w.Header().Set("Content-Type", "application/json")
		}
	}
}
