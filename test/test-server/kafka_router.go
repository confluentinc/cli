package testserver

import (
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type KafkaRestProxyRouter struct {
	*mux.Router
}

func NewKafkaRestProxyRouter(t *testing.T) *KafkaRestProxyRouter {
	router := &KafkaRestProxyRouter{mux.NewRouter()}

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "{}")
		require.NoError(t, err)
	})

	routes := map[string]http.HandlerFunc{
		"/kafka/v3/clusters":                                                                                              router.HandleKafkaRPClusters(t),
		"/kafka/v3/clusters/{cluster_id}/acls:batch":                                                                      router.HandleKafkaRPACLsBatch(t),
		"/kafka/v3/clusters/{cluster_id}/broker-configs":                                                                  router.HandleKafkaBrokerConfigs(t),
		"/kafka/v3/clusters/{cluster_id}/broker-configs/{name}":                                                           router.HandleKafkaBrokerConfigsName(t),
		"/kafka/v3/clusters/{cluster_id}/broker-configs:alter":                                                            router.HandleKafkaBrokerConfigsAlter(t),
		"/kafka/v3/clusters/{cluster_id}/brokers":                                                                         router.HandleKafkaBrokers(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/-/tasks":                                                                 router.HandleKafkaClustersClusterIdBrokersTasksGet(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/-/tasks/{task_type}":                                                     router.HandleKafkaClustersClusterIdBrokersTasksTaskTypeGet(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}":                                                             router.HandleKafkaBrokersBrokerId(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs":                                                     router.HandleKafkaBrokerIdConfigs(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs/{name}":                                              router.HandleKafkaBrokerIdConfigsName(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs:alter":                                               router.HandleKafkaBrokerIdConfigsAlter(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks":                                                       router.HandleKafkaClustersClusterIdBrokersBrokerIdTasksGet(t),
		"/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks/{task_type}":                                           router.HandleKafkaClustersClusterIdBrokersBrokerIdTasksTaskTypeGet(t),
		"/kafka/v3/clusters/{cluster_id}/consumer-groups":                                                                 router.HandleKafkaRPConsumerGroups(t),
		"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}":                                             router.HandleKafkaRPConsumerGroup(t),
		"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/consumers":                                   router.HandleKafkaRPConsumers(t),
		"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lag-summary":                                 router.HandleKafkaRPLagSummary(t),
		"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags":                                        router.HandleKafkaRPLags(t),
		"/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags/{topic_name}/partitions/{partition_id}": router.HandleKafkaRPLag(t),
		"/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions":                                                  router.HandleKafkaTopicPartitions(t),
		"/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}":                                   router.HandleKafkaTopicPartitionId(t),
		"/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}/reassignment":                      router.HandleKafkaTopicPartitionIdReassignment(t),
		"/kafka/v3/clusters/{cluster_id}/topics/{topic}/partitions/-/replica-status":                                      router.HandleKafkaRPReplicaStatus(t),
		"/kafka/v3/clusters/{cluster}/acls":                                                                               router.HandleKafkaRPACLs(t),
		"/kafka/v3/clusters/{cluster}/links":                                                                              router.HandleKafkaRPLinks(t),
		"/kafka/v3/clusters/{cluster}/links/-/mirrors":                                                                    router.HandleKafkaRPAllMirrors(t),
		"/kafka/v3/clusters/{cluster}/links/{link}":                                                                       router.HandleKafkaRPLink(t),
		"/kafka/v3/clusters/{cluster}/links/{link}/configs":                                                               router.HandleKafkaRPLinkConfigs(t),
		"/kafka/v3/clusters/{cluster}/links/{link}/mirrors":                                                               router.HandleKafkaRPMirrors(t),
		"/kafka/v3/clusters/{cluster}/links/{link}/mirrors/{mirror_topic_name}":                                           router.HandleKafkaRPMirror(t),
		"/kafka/v3/clusters/{cluster}/links/{link}/mirrors:promote":                                                       router.HandleKafkaRPMirrorsPromote(t),
		"/kafka/v3/clusters/{cluster}/topic/{topic}/partitions/-/replica-status":                                          router.HandleClustersClusterIdTopicsTopicsNamePartitionsReplicaStatus(t),
		"/kafka/v3/clusters/{cluster}/topics":                                                                             router.HandleKafkaRPTopics(t),
		"/kafka/v3/clusters/{cluster}/topics/{topic}":                                                                     router.HandleKafkaRPTopic(t),
		"/kafka/v3/clusters/{cluster}/topics/{topic}/configs":                                                             router.HandleKafkaRPTopicConfigs(t),
		"/kafka/v3/clusters/{cluster}/topics/{topic}/configs:alter":                                                       router.HandleKafkaRPConfigsAlter(t),
		"/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replica-status":                               router.HandleClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicaStatus(t),
		"/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replicas":                                     router.HandleKafkaRPPartitionReplicas(t),
	}

	for path, f := range routes {
		router.HandleFunc(path, f)
	}

	return router
}
