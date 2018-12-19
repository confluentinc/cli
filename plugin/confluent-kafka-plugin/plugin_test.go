package main

import (
	"fmt"
	"testing"
	"net/http"
	"encoding/json"
	"net/http/httptest"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	chttp "github.com/confluentinc/ccloud-sdk-go"
	log "github.com/confluentinc/cli/log"

	"context"
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
		fmt.Println(json.NewEncoder(w).Encode(&kafkav1.TopicDescription{}))
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
	json.NewEncoder(w).Encode([]kafkav1.ACLBinding{})
}

func TestMockClient(t *testing.T) {
	logger := log.New()
	done := make(chan struct{})


	client := NewMockClient(logger, done)
	t.Log(client.Kafka.ListTopics(context.Background(), &kafkav1.Cluster{}))
	t.Log(client.Kafka.CreateTopic(context.Background(), &kafkav1.Cluster{}, &kafkav1.Topic{}))
	t.Log(client.Kafka.DeleteTopic(context.Background(), &kafkav1.Cluster{}, &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: "topic_test"}}))
	t.Log(client.Kafka.UpdateTopic(context.Background(), &kafkav1.Cluster{}, &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: "topic_test"}}))
	t.Log(client.Kafka.DescribeTopic(context.Background(), &kafkav1.Cluster{}, &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: "topic_test"}}))
	t.Log(client.Kafka.CreateACL(context.Background(), &kafkav1.Cluster{}, &kafkav1.ACLBinding{}))
	t.Log(client.Kafka.DeleteACL(context.Background(), &kafkav1.Cluster{}, &kafkav1.ACLFilter{}))
	t.Log(client.Kafka.ListACL(context.Background(), &kafkav1.Cluster{}, &kafkav1.ACLFilter{}))


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
	mux.HandleFunc(fmt.Sprintf(string(chttp.TOPICS), clusterID), handleTopics)
	mux.HandleFunc(fmt.Sprintf(string(chttp.TOPIC), clusterID, "topic_test"), handleTopic)
	mux.HandleFunc(fmt.Sprintf(string(chttp.TOPICCONFIG), clusterID, "topic_test"), handleTopicConfig)
	mux.HandleFunc(fmt.Sprintf(string(chttp.ACL), clusterID), handleACL)
	mux.HandleFunc(fmt.Sprintf(string(chttp.ACLSEARCH), clusterID), handleACLSearch)

	api = httptest.NewServer(mux)
	fmt.Println(api.URL)
	logger.Log("msg", api.URL)
	client := chttp.NewClient(api.URL, api.Client(), logger)

	return client
}
