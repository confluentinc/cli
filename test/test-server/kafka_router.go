package test_server

import (
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// kafka urls
const (
	// kafka api urls
	aclsCreate      = "/2.0/kafka/{cluster}/acls"
	aclsList        = "/2.0/kafka/{cluster}/acls:search"
	aclsDelete      = "/2.0/kafka/{cluster}/acls/delete"
	link            = "/2.0/kafka/{cluster}/links/{link}"
	links           = "/2.0/kafka/{cluster}/links"
	topicMirrorStop = "/2.0/kafka/{cluster}/topics/{topic}/mirror:stop"
	topics          = "/2.0/kafka/{cluster}/topics"
	topic           = "/2.0/kafka/{cluster}/topics/{topic}"
	topicConfig     = "/2.0/kafka/{cluster}/topics/{topic}/config"

	//kafka rest urls
	rpAcls              = "/kafka/v3/clusters/{cluster}/acls"
	rpTopics            = "/kafka/v3/clusters/{cluster}/topics"
	rpPartitions        = "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions"
	rpPartitionReplicas = "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replicas"
	rpTopicConfigs      = "/kafka/v3/clusters/{cluster}/topics/{topic}/configs"
	rpConfigsAlter      = "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/configs:alter"
	rpTopic             = "/kafka/v3/clusters/{cluster}/topics/{topic}"
)

type KafkaRouter struct {
	KafkaApi KafkaApiRouter
	KafkaRP  KafkaRestProxyRouter
}

type KafkaApiRouter struct {
	*mux.Router
}

type KafkaRestProxyRouter struct {
	*mux.Router
}

func NewKafkaRouter(t *testing.T) *KafkaRouter {
	router := NewEmptyKafkaRouter()
	router.KafkaApi.buildKafkaApiHandler(t)
	router.KafkaRP.buildKafkaRPHandler(t)
	return router
}

func NewEmptyKafkaRouter() *KafkaRouter {
	return &KafkaRouter{
		KafkaApi: KafkaApiRouter{mux.NewRouter()},
		KafkaRP:  KafkaRestProxyRouter{mux.NewRouter()},
	}
}

func (k *KafkaApiRouter) buildKafkaApiHandler(t *testing.T) {
	k.HandleFunc(aclsCreate, k.HandleKafkaACLsCreate(t))
	k.HandleFunc(aclsList, k.HandleKafkaACLsList(t))
	k.HandleFunc(aclsDelete, k.HandleKafkaACLsDelete(t))
	k.HandleFunc(link, k.HandleKafkaLink(t))
	k.HandleFunc(links, k.HandleKafkaLinks(t))
	k.HandleFunc(topicMirrorStop, k.HandleKafkaTopicMirrorStop(t))
	k.HandleFunc(topics, k.HandleKafkaListCreateTopic(t))
	k.HandleFunc(topic, k.HandleKafkaDescribeDeleteTopic(t))
	k.HandleFunc(topicConfig, k.HandleKafkaTopicListConfig(t))
	k.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
}

func (r KafkaRestProxyRouter) buildKafkaRPHandler(t *testing.T) {
	r.HandleFunc(rpAcls, r.HandleKafkaRPACLs(t))
	r.HandleFunc(rpTopics, r.HandleKafkaRPTopics(t))
	r.HandleFunc(rpPartitions, r.HandleKafkaRPPartitions(t))
	r.HandleFunc(rpTopicConfigs, r.HandleKafkaRPTopicConfigs(t))
	r.HandleFunc(rpPartitionReplicas, r.HandleKafkaRPPartitionReplicas(t))
	r.HandleFunc(rpConfigsAlter, r.HandleKafkaRPConfigsAlter(t))
	r.HandleFunc(rpTopic, r.HandlKafkaRPTopic(t))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
}
