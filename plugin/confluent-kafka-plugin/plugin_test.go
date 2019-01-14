package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	log "github.com/confluentinc/cli/log"

	"context"
)

var (
	api     *httptest.Server
	cluster = &kafkav1.Cluster{
		Id: "cluster_test",
	}

	topic = &kafkav1.Topic{
		Spec: &kafkav1.TopicSpecification{
			Name: "topic_test",
		},
	}

	tConfig = &kafkav1.TopicConfig{
		Entries: []*kafkav1.TopicConfigEntry{{Name: "min.insync.replicas", Value: "1"}},
	}

	token = &kafkav1.KafkaAPI{
		Token: "test-token",
	}
)

func handleAuthorization(w http.ResponseWriter, req *http.Request) bool {
	if "Bearer "+token.Token != req.Header.Get("Authorization") {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}

// @Path /access_tokens
// post: returns short-lived token used by KafkaAPI to authenticate requests
func handleToken(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		_ = json.NewEncoder(w).Encode(token)
	}
}

// @Path /topics
// get,put: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L66-L86
func handleTopics(w http.ResponseWriter, req *http.Request) {
	if !handleAuthorization(w, req) {
		return
	}
	switch req.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Path /topics/{topic}
// get: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L161-L170
// delete: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L88-L97
func handleTopic(w http.ResponseWriter, req *http.Request) {
	if !handleAuthorization(w, req) {
		return
	}
	var topic *kafkav1.Topic
	_ = json.NewDecoder(req.Body).Decode(&topic)

	switch req.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(topic)
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	}

}

// @Path /topics/{topic}/config
// get/put: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L172-L207
func handleTopicConfig(w http.ResponseWriter, req *http.Request) {
	if !handleAuthorization(w, req) {
		return
	}
	switch req.Method {
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(tConfig)
	case http.MethodPut:
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Path /acls
// post/delete: https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L209-L231
func handleACL(w http.ResponseWriter, req *http.Request) {
	if !handleAuthorization(w, req) {
		return
	}
	switch req.Method {
	case http.MethodPost:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Path /acls:search
//https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java#L303-L322
func handleACLSearch(w http.ResponseWriter, req *http.Request) {
	if !handleAuthorization(w, req) {
		return
	}
	_ = json.NewEncoder(w).Encode([]kafkav1.ACLBinding{})
}

// TestMockClient ensures the plugin properly routes requests to the KafkaAPI given a particular set of arguments
func TestMockClient(t *testing.T) {
	logger := log.New()
	done := make(chan struct{})

	client := NewMockClient(logger, done, t)
	t.Log(client.Kafka.ListTopics(context.Background(), cluster))
	t.Log(client.Kafka.CreateTopic(context.Background(), cluster, topic))
	t.Log(client.Kafka.DeleteTopic(context.Background(), cluster, topic))
	t.Log(client.Kafka.UpdateTopic(context.Background(), cluster, topic))
	t.Log(client.Kafka.DescribeTopic(context.Background(), cluster, topic))
	t.Log(client.Kafka.CreateACL(context.Background(), cluster, []*kafkav1.ACLBinding{&kafkav1.ACLBinding{}}))
	t.Log(client.Kafka.DeleteACL(context.Background(), cluster, &kafkav1.ACLFilter{}))
	t.Log(client.Kafka.ListACL(context.Background(), cluster, &kafkav1.ACLFilter{}))

	done <- struct{}{}
}

func NewMockClient(logger *log.Logger, done chan struct{}, t *testing.T) *chttp.Client {
	go func() {
		<-done
		api.Close()
	}()

	mux := http.NewServeMux()
	mux.HandleFunc(chttp.ACCESS_TOKENS, handleToken)
	mux.HandleFunc(fmt.Sprintf(chttp.TOPICS, cluster.Id), handleTopics)
	mux.HandleFunc(fmt.Sprintf(chttp.TOPIC, cluster.Id, topic.Spec.Name), handleTopic)
	mux.HandleFunc(fmt.Sprintf(chttp.TOPICCONFIG, cluster.Id, topic.Spec.Name), handleTopicConfig)
	mux.HandleFunc(fmt.Sprintf(chttp.ACL, cluster.Id), handleACL)
	mux.HandleFunc(fmt.Sprintf(chttp.ACLSEARCH, cluster.Id), handleACLSearch)

	api = httptest.NewServer(mux)

	cluster.ApiEndpoint = api.URL

	client := chttp.NewClient(api.URL, api.Client(), logger)

	return client
}
