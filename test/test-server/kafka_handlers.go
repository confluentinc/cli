package test_server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"

	linkv1 "github.com/confluentinc/cc-structs/kafka/clusterlink/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for: "/2.0/kafka/{cluster}/acls:search"
func (k *KafkaRouter) HandleKafkaACLsList(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		results := []*schedv1.ACLBinding{
			{
				Pattern: &schedv1.ResourcePatternConfig{
					ResourceType: schedv1.ResourceTypes_TOPIC,
					Name:         "test-topic",
					PatternType:  schedv1.PatternTypes_LITERAL,
				},
				Entry: &schedv1.AccessControlEntryConfig{
					Operation:      schedv1.ACLOperations_READ,
					PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				},
			},
		}
		reply, err := json.Marshal(results)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}

// Handler for: "/2.0/kafka/{cluster}/acls"
func (k *KafkaRouter) HandleKafkaACLsCreate(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var bindings []*schedv1.ACLBinding
			err := json.NewDecoder(r.Body).Decode(&bindings)
			require.NoError(t, err)
			require.NotEmpty(t, bindings)
			for _, binding := range bindings {
				require.NotEmpty(t, binding.GetPattern())
				require.NotEmpty(t, binding.GetEntry())
			}
		}
	}
}

// Handler for: "/2.0/kafka/{cluster}/acls/delete"
func (k *KafkaRouter) HandleKafkaACLsDelete(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var filters []*schedv1.ACLFilter
		err := json.NewDecoder(r.Body).Decode(&filters)
		require.NoError(t, err)
		require.NotEmpty(t, filters)
		for _, filter := range filters {
			require.NotEmpty(t, filter.GetEntryFilter())
			require.NotEmpty(t, filter.GetPatternFilter())
		}
	}
}

// Handler for: "/2.0/kafka/{cluster}/links"
func (k *KafkaRouter) HandleKafkaLinks(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		listResponsePayload := []*linkv1.ListLinksResponseItem{
			&linkv1.ListLinksResponseItem{LinkName: "link-1", LinkId: "1234", ClusterId: "Blah"},
			&linkv1.ListLinksResponseItem{LinkName: "link-2", LinkId: "4567", ClusterId: "blah"},
		}

		listReply, err := json.Marshal(listResponsePayload)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(listReply))
		require.NoError(t, err)
	}
}

// Handler for: "/2.0/kafka/{cluster}/links/{link}"
func (k *KafkaRouter) HandleKafkaLink(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		link := vars["link"]
		var describeResponsePayload linkv1.DescribeLinkResponse
		switch link {
		default:
			describeResponsePayload = linkv1.DescribeLinkResponse{
				Entries: []*linkv1.DescribeLinkResponseEntry{
					{
						Name:  "replica.fetch.max.bytes",
						Value: "1048576",
					},
				},
			}
		}
		// Return properties for the selected link.
		describeReply, err := json.Marshal(describeResponsePayload)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(describeReply))
		require.NoError(t, err)
	}
}

// Handler for: "/2.0/kafka/{cluster}/topics/{topic}/mirror:stop"
func (k *KafkaRouter) HandleKafkaTopicMirrorStop(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topic := vars["topic"]
		switch topic {
		case "not-found":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			if r.Method == "POST" {
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
}

// Handler for: "/2.0/kafka/{cluster}/topics"
func (k *KafkaRouter) HandleKafkaListCreateTopic(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET": //list call
			k.HandleKafkaListTopic(t)(w, r)
		case "PUT": //create call
			k.HandleKafkaCreateTopic(t)(w, r)
		}
	}
}

// Handler for: GET "/2.0/kafka/{cluster}/topics"
func (k *KafkaRouter) HandleKafkaListTopic(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		cluster := vars["cluster"]
		var listTopicReply schedv1.ListTopicReply
		switch cluster {
		case "lkc-topics":
			listTopicReply = schedv1.ListTopicReply{
				Topics: []*schedv1.TopicDescription{
					{
						Name: "topic1",
					},
					{
						Name: "topic2",
					},
				},
			}
		case "lkc-no-topics":
			listTopicReply = schedv1.ListTopicReply{Topics: []*schedv1.TopicDescription{}}
		default: //cluster not ready
			w.WriteHeader(http.StatusInternalServerError)
			listTopicReply = schedv1.ListTopicReply{Error: &schedv1.KafkaAPIError{Message: "Authentication failed: 1 extensions are invalid! They are: logicalCluster: Authentication failed"}}
			topicReply, err := json.Marshal(listTopicReply.Error)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(topicReply))
			require.NoError(t, err)
			return
		}
		topicReply, err := json.Marshal(listTopicReply.Topics)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(topicReply))
		require.NoError(t, err)
	}
}

// Handler for: POST "/2.0/kafka/{cluster}/topics"
func (k *KafkaRouter) HandleKafkaCreateTopic(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		cluster := vars["cluster"]
		spec := &schedv1.TopicSpecification{} //CreateTopicRequest{}
		err := utilv1.UnmarshalJSON(r.Body, spec)
		require.NoError(t, err)
		topicName := spec.Name
		var createTopicReply *schedv1.CreateTopicReply
		switch {
		case cluster == "lkc-create-topic" && topicName == "dupTopic":
			w.WriteHeader(http.StatusBadRequest)
			createTopicReply = &schedv1.CreateTopicReply{Error: &schedv1.KafkaAPIError{Message: "topic already exists"}}
		default:
			createTopicReply = &schedv1.CreateTopicReply{}
		}
		topicReply, err := json.Marshal(createTopicReply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(topicReply))
		require.NoError(t, err)
	}
}

// Handler for: "/2.0/kafka/{cluster}/topics/{topic}"
func (k *KafkaRouter) HandleKafkaDescribeDeleteTopic(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET": //describe call
			k.HandleKafkaDescribeTopic(t)(w, r)
		case "DELETE": //delete call
			k.HandleKafkaDeleteTopic(t)(w, r)
		}
	}
}

// Handler for: GET "/2.0/kafka/{cluster}/topics/{topic}"
func (k *KafkaRouter) HandleKafkaDescribeTopic(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		cluster := vars["cluster"]
		topic := vars["topic"]
		var describeTopicReply *schedv1.DescribeTopicReply
		switch {
		case cluster == "lkc-describe-topic" && topic == "topic1":
			describeTopicReply = &schedv1.DescribeTopicReply{Topic: &schedv1.TopicDescription{
				Name:       "topic1",
				Partitions: []*schedv1.TopicPartitionInfo{{Partition: 1, Leader: &schedv1.KafkaNode{Id: 1}, Replicas: []*schedv1.KafkaNode{{Id: 1}}}},
			}}
		default:
			w.WriteHeader(http.StatusNotFound)
			describeTopicReply = &schedv1.DescribeTopicReply{Error: &schedv1.KafkaAPIError{Message: "topic not found"}}
			var err error
			topicReply, err := json.Marshal(describeTopicReply.Error)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(topicReply))
			require.NoError(t, err)
			return
		}
		topicReply, err := json.Marshal(describeTopicReply.Topic)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(topicReply))
		require.NoError(t, err)
	}
}

// Handler for: DELETE "/2.0/kafka/{cluster}/topics/{topic}"
func (k *KafkaRouter) HandleKafkaDeleteTopic(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		cluster := vars["cluster"]
		topic := vars["topic"]
		var deleteTopicReply *schedv1.DeleteTopicReply
		switch {
		case cluster == "lkc-delete-topic" && topic == "topic1":
			deleteTopicReply = &schedv1.DeleteTopicReply{}
		default:
			w.WriteHeader(http.StatusNotFound)
			deleteTopicReply = &schedv1.DeleteTopicReply{Error: &schedv1.KafkaAPIError{Message: "topic not found"}}
		}
		topicReply, err := json.Marshal(deleteTopicReply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(topicReply))
		require.NoError(t, err)
	}
}

//Handler for: "/2.0/kafka/{cluster}/topics/{topic}/config"
func (k *KafkaRouter) HandleKafkaTopicListConfig(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var listTopicConfigReply *schedv1.ListTopicConfigReply
		if r.Method == "GET" { //part of describe call
			listTopicConfigReply = &schedv1.ListTopicConfigReply{TopicConfig: &schedv1.TopicConfig{Entries: []*schedv1.TopicConfigEntry{{Name: "testConfig", Value: "testValue"}}}}
			topicReply, err := json.Marshal(listTopicConfigReply.TopicConfig)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(topicReply))
			require.NoError(t, err)
		}
		// PUTs are part of update calls
		// nothing to validate
	}
}
