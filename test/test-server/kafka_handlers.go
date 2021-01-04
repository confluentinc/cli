package test_server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

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
