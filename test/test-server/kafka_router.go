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
	rpReplicaStatus     = "/kafka/v3/clusters/{cluster_id}/topics/{topic}/partitions/-/replica-status"
	rpTopicConfigs      = "/kafka/v3/clusters/{cluster}/topics/{topic}/configs"
	rpConfigsAlter      = "/kafka/v3/clusters/{cluster}/topics/{topic}/configs:alter"
	rpTopic             = "/kafka/v3/clusters/{cluster}/topics/{topic}"
	rpLink              = "/kafka/v3/clusters/{cluster}/links/{link}"
	rpLinks             = "/kafka/v3/clusters/{cluster}/links"
	rpLinkConfigs       = "/kafka/v3/clusters/{cluster}/links/{link}/configs"
	rpMirror            = "/kafka/v3/clusters/{cluster}/links/{link}/mirrors/{mirror_topic_name}"
	rpAllMirrors        = "/kafka/v3/clusters/{cluster}/links/-/mirrors"
	rpMirrors           = "/kafka/v3/clusters/{cluster}/links/{link}/mirrors"
	rpMirrorPromote     = "/kafka/v3/clusters/{cluster}/links/{link}/mirrors:promote"
	rpClusters          = "/kafka/v3/clusters"
	rpConsumerGroups    = "/kafka/v3/clusters/{cluster_id}/consumer-groups"
	rpConsumerGroup     = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}"
	rpConsumers         = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/consumers"
	rpLagSummary        = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lag-summary"
	rpLags              = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags"
	rpLag               = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags/{topic_name}/partitions/{partition_id}"
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
	r.HandleFunc(rpReplicaStatus, r.HandleKafkaRPReplicaStatus(t))
	r.HandleFunc(rpConfigsAlter, r.HandleKafkaRPConfigsAlter(t))
	r.HandleFunc(rpTopic, r.HandleKafkaRPTopic(t))
	r.HandleFunc(rpLink, r.HandleKafkaRPLink(t))
	r.HandleFunc(rpLinks, r.HandleKafkaRPLinks(t))
	r.HandleFunc(rpLinkConfigs, r.HandleKafkaRPLinkConfigs(t))
	r.HandleFunc(rpMirrorPromote, r.HandleKafkaRPMirrorsPromote(t))
	r.HandleFunc(rpMirror, r.HandleKafkaRPMirror(t))
	r.HandleFunc(rpAllMirrors, r.HandleKafkaRPAllMirrors(t))
	r.HandleFunc(rpMirrors, r.HandleKafkaRPMirrors(t))
	r.HandleFunc(rpClusters, r.HandleKafkaRPClusters(t))
	r.HandleFunc(rpConsumerGroups, r.HandleKafkaRPConsumerGroups(t))
	r.HandleFunc(rpConsumerGroup, r.HandleKafkaRPConsumerGroup(t))
	r.HandleFunc(rpConsumers, r.HandleKafkaRPConsumers(t))
	r.HandleFunc(rpLagSummary, r.HandleKafkaRPLagSummary(t))
	r.HandleFunc(rpLags, r.HandleKafkaRPLags(t))
	r.HandleFunc(rpLag, r.HandleKafkaRPLag(t))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
}
