package testserver

import (
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

const (
	rpAcls                     = "/kafka/v3/clusters/{cluster}/acls"
	rpAclsBatch                = "/kafka/v3/clusters/{cluster_id}/acls:batch"
	rpTopics                   = "/kafka/v3/clusters/{cluster}/topics"
	rpReplicaStatus            = "/kafka/v3/clusters/{cluster_id}/topics/{topic}/partitions/-/replica-status"
	rpTopicConfigs             = "/kafka/v3/clusters/{cluster}/topics/{topic}/configs"
	rpConfigsAlter             = "/kafka/v3/clusters/{cluster}/topics/{topic}/configs:alter"
	rpTopic                    = "/kafka/v3/clusters/{cluster}/topics/{topic}"
	rpLink                     = "/kafka/v3/clusters/{cluster}/links/{link}"
	rpLinks                    = "/kafka/v3/clusters/{cluster}/links"
	rpLinkConfigs              = "/kafka/v3/clusters/{cluster}/links/{link}/configs"
	rpMirror                   = "/kafka/v3/clusters/{cluster}/links/{link}/mirrors/{mirror_topic_name}"
	rpAllMirrors               = "/kafka/v3/clusters/{cluster}/links/-/mirrors"
	rpMirrors                  = "/kafka/v3/clusters/{cluster}/links/{link}/mirrors"
	rpMirrorPromote            = "/kafka/v3/clusters/{cluster}/links/{link}/mirrors:promote"
	rpClusters                 = "/kafka/v3/clusters"
	rpConsumerGroups           = "/kafka/v3/clusters/{cluster_id}/consumer-groups"
	rpConsumerGroup            = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}"
	rpConsumers                = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/consumers"
	rpLagSummary               = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lag-summary"
	rpLags                     = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags"
	rpLag                      = "/kafka/v3/clusters/{cluster_id}/consumer-groups/{consumer_group_id}/lags/{topic_name}/partitions/{partition_id}"
	rpTopicNamePartitions      = "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions"
	rpPartitionIdPartitions    = "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}"
	rpPartitionIdReassignments = "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/partitions/{partition_id}/reassignment"
	rpBrokers                  = "/kafka/v3/clusters/{cluster_id}/brokers"
	rpBrokerConfigsName        = "/kafka/v3/clusters/{cluster_id}/broker-configs/{name}"
	rpBrokerConfigs            = "/kafka/v3/clusters/{cluster_id}/broker-configs"
	rpBrokerIdConfigName       = "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs/{name}"
	rpBrokerIdConfigs          = "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs"
	rpBrokerConfigsAlter       = "/kafka/v3/clusters/{cluster_id}/broker-configs:alter"
	rpBrokerIdConfigsAlter     = "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/configs:alter"
	rpBrokerId                 = "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}"
	rpBrokersTaskType          = "/kafka/v3/clusters/{cluster_id}/brokers/-/tasks/{task_type}"
	rpBrokersTasks             = "/kafka/v3/clusters/{cluster_id}/brokers/-/tasks"
	rpBrokerIdTaskType         = "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks/{task_type}"
	rpBrokerIdTasks            = "/kafka/v3/clusters/{cluster_id}/brokers/{broker_id}/tasks"
	rpPartitionReplicas        = "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replicas"
	rpTopicReplicaStatuses     = "/kafka/v3/clusters/{cluster}/topic/{topic}/partitions/-/replica-status"
	rpPartitionReplicaStatuses = "/kafka/v3/clusters/{cluster}/topics/{topic}/partitions/{partition}/replica-status"
)

type KafkaRouter struct {
	KafkaRP KafkaRestProxyRouter
}

type KafkaApiRouter struct {
	*mux.Router
}

type KafkaRestProxyRouter struct {
	*mux.Router
}

func NewKafkaRouter(t *testing.T) *KafkaRouter {
	router := NewEmptyKafkaRouter()
	router.KafkaRP.buildKafkaRPHandler(t)
	return router
}

func NewEmptyKafkaRouter() *KafkaRouter {
	return &KafkaRouter{
		KafkaRP: KafkaRestProxyRouter{mux.NewRouter()},
	}
}

func (r KafkaRestProxyRouter) buildKafkaRPHandler(t *testing.T) {
	r.HandleFunc(rpAcls, r.HandleKafkaRPACLs(t))
	r.HandleFunc(rpAclsBatch, r.HandleKafkaRPACLsBatch(t))
	r.HandleFunc(rpTopics, r.HandleKafkaRPTopics(t))
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
	r.HandleFunc(rpTopicNamePartitions, r.HandleKafkaTopicPartitions(t))
	r.HandleFunc(rpPartitionIdPartitions, r.HandleKafkaTopicPartitionId(t))
	r.HandleFunc(rpPartitionIdReassignments, r.HandleKafkaTopicPartitionIdReassignment(t))
	r.HandleFunc(rpBrokers, r.HandleKafkaBrokers(t))
	r.HandleFunc(rpBrokerConfigsName, r.HandleKafkaBrokerConfigsName(t))
	r.HandleFunc(rpBrokerConfigs, r.HandleKafkaBrokerConfigs(t))
	r.HandleFunc(rpBrokerIdConfigName, r.HandleKafkaBrokerIdConfigsName(t))
	r.HandleFunc(rpBrokerIdConfigs, r.HandleKafkaBrokerIdConfigs(t))
	r.HandleFunc(rpBrokerConfigsAlter, r.HandleKafkaBrokerConfigsAlter(t))
	r.HandleFunc(rpBrokerIdConfigsAlter, r.HandleKafkaBrokerIdConfigsAlter(t))
	r.HandleFunc(rpBrokerId, r.HandleKafkaBrokersBrokerId(t))
	r.HandleFunc(rpBrokersTaskType, r.HandleKafkaClustersClusterIdBrokersTasksTaskTypeGet(t))
	r.HandleFunc(rpBrokersTasks, r.HandleKafkaClustersClusterIdBrokersTasksGet(t))
	r.HandleFunc(rpBrokerIdTaskType, r.HandleKafkaClustersClusterIdBrokersBrokerIdTasksTaskTypeGet(t))
	r.HandleFunc(rpBrokerIdTasks, r.HandleKafkaClustersClusterIdBrokersBrokerIdTasksGet(t))
	r.HandleFunc(rpTopicReplicaStatuses, r.HandleClustersClusterIdTopicsTopicsNamePartitionsReplicaStatus(t))
	r.HandleFunc(rpPartitionReplicaStatuses, r.HandleClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicaStatus(t))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
}
