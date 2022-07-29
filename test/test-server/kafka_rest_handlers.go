package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for: "/kafka/v3/clusters/{cluster}/acls"
func handleKafkaRestAcls(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		data := kafkarestv3.AclData{
			ClusterId:    mux.Vars(r)["cluster"],
			ResourceType: kafkarestv3.TOPIC,
			ResourceName: "test-topic",
			Operation:    "READ",
			Permission:   "ALLOW",
			Host:         "*",
			Principal:    "User:12345",
			PatternType:  "LITERAL",
		}

		var res interface{}
	
		switch r.Method {
		case http.MethodGet:
			res = &kafkarestv3.AclDataList{Data: []kafkarestv3.AclData{data}}
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			res = &kafkarestv3.AclData{}
		case http.MethodDelete:
			res = &kafkarestv3.InlineResponse200{Data: []kafkarestv3.AclData{data}}
		}

		err := json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}
