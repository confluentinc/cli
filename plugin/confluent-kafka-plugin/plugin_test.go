package main

import (
	"fmt"
	"net/http/httptest"
	"net/http"
	"encoding/json"

	chttp "github.com/confluentinc/cli/http"
	log "github.com/confluentinc/cli/log"
	"testing"
	"github.com/confluentinc/cli/shared/kafka"
)

// @Path /topics
// get,put: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L66-L86
func handleTopics(w http.ResponseWriter, req *http.Request){
	switch req.Method{
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Path /topics/{topic}
// get: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L161-L170
// delete: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L88-L97
func handleTopic(w http.ResponseWriter, req *http.Request){
	switch req.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		fmt.Println(json.NewEncoder(w).Encode(&kafka.KafkaTopicDescription{}))
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	}

}

// @Path /topics/{topic}/config
// get/put: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L172-L207
func handleTopicConfig(w http.ResponseWriter, req *http.Request) {
	switch req.Method{
	case http.MethodGet:
		fmt.Println(json.NewEncoder(w).Encode([]string{"test"}))
	case http.MethodPut:
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Path /acls
// post/delete: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L209-L231
func handleACL(w http.ResponseWriter, req *http.Request) {
	switch req.Method{
	case http.MethodPost:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Path /acls:search
//https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L303-L322
func handleACLSearch(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode([]kafka.KafkaAPIACLRequest{})
}

func TestMockClient(t *testing.T) {
	logger := log.New()
	done := make(chan struct{})

	client := NewMockClient(logger, done)
	client.Kafka.ListTopics()
	client.Kafka.CreateTopic(&kafka.KafkaAPITopicRequest{})
	client.Kafka.DeleteTopic(&kafka.KafkaAPITopicRequest{Spec: &kafka.KafkaTopicSpecification{Name: "topic_test"},})
	client.Kafka.UpdateTopic(&kafka.KafkaAPITopicRequest{Spec: &kafka.KafkaTopicSpecification{Name: "topic_test"},})
	client.Kafka.DescribeTopic(&kafka.KafkaAPITopicRequest{Spec: &kafka.KafkaTopicSpecification{Name: "topic_test"},})
	client.Kafka.CreateACL(&kafka.KafkaAPIACLRequest{})
	client.Kafka.DeleteACL(&kafka.KafkaAPIACLFilterRequest{})
	client.Kafka.ListACL(&kafka.KafkaAPIACLFilterRequest{})

	done<-struct{}{}
}

func NewMockClient(logger *log.Logger, done chan struct{}) *chttp.Client {
	var api *httptest.Server
	const clusterID = "test"
	go func(){
		<-done
		api.Close()
	}()

	mux := 	http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf(chttp.TOPICS, clusterID), handleTopics)
	mux.HandleFunc(fmt.Sprintf(chttp.TOPIC, clusterID, "topic_test"), handleTopic)
	mux.HandleFunc(fmt.Sprintf(chttp.TOPICCONFIG, clusterID, "topic_test"), handleTopicConfig)
	mux.HandleFunc(fmt.Sprintf(chttp.ACL, clusterID), handleACL)
	mux.HandleFunc(fmt.Sprintf(chttp.ACLSEARCH, clusterID), handleACLSearch)

	api = httptest.NewServer(mux)

	client := chttp.NewClient(api.Client(), api.URL, logger)
	client.Kafka.ConfigureKafkaAPI(clusterID, api.URL)

	return client
}
