package test_server

import (
	"io"
	"net/http"
	"strings"
	"testing"

	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for: "/api/usage_limits"
func (c *CloudRouter) HandleUsageLimits(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		usageLimitsReply := &schedv1.GetUsageLimitsReply{UsageLimits: &productv1.UsageLimits{
			TierLimits: map[string]*productv1.TierFixedLimits{
				"BASIC": {
					PartitionLimits: &productv1.KafkaPartitionLimits{},
					ClusterLimits:   &productv1.KafkaClusterLimits{},
				},
			},
			CkuLimits: map[uint32]*productv1.CKULimits{
				uint32(1): {
					NumBrokers: &productv1.IntegerUsageLimit{Limit: &productv1.IntegerUsageLimit_Value{Value: 5}},
					Storage: &productv1.IntegerUsageLimit{
						Limit: &productv1.IntegerUsageLimit_Value{Value: 500},
						Unit:  productv1.LimitUnit_GB,
					},
					NumPartitions: &productv1.IntegerUsageLimit{Limit: &productv1.IntegerUsageLimit_Value{Value: 2000}},
				},
				uint32(2): {
					NumBrokers: &productv1.IntegerUsageLimit{Limit: &productv1.IntegerUsageLimit_Value{Value: 5}},
					Storage: &productv1.IntegerUsageLimit{
						Limit: &productv1.IntegerUsageLimit_Value{Value: 1000},
						Unit:  productv1.LimitUnit_GB,
					},
					NumPartitions: &productv1.IntegerUsageLimit{Limit: &productv1.IntegerUsageLimit_Value{Value: 4000}},
				},
			},
		}}
		b, err := utilv1.MarshalJSONToBytes(usageLimitsReply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

func (c *CloudRouter) HandleCluster(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clusterId := vars["id"]
		switch clusterId {
		case "lkc-topics", "lkc-create-topic", "lkc-describe-topic", "lkc-delete-topic", "lkc-acls", "lkc-create-topic-kafka-api", "lkc-describe-topic-kafka-api", "lkc-delete-topic-kafka-api":
			c.HandleKafkaApiOrRestClusters(t)(w, r)
		case "lkc-unknown":
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		default:
			c.HandleKafkaClusterGetListDeleteDescribe(t)(w, r)
		}
	}
}

func (c *CloudRouter) HandleKafkaApiOrRestClusters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clusterId := vars["id"]
		id := r.URL.Query().Get("id")
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.ApiEndpoint = c.kafkaApiUrl
		if !strings.Contains(clusterId, "kafka-api") {
			cluster.RestEndpoint = c.kafkaRPUrl
		}
		cluster.Endpoint = "SASL_SSL://127.0.0.1:"
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

func (c *CloudRouter) HandleKafkaClusterGetListDeleteDescribe(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// this is in the body of delete requests
		require.NotEmpty(t, r.URL.Query().Get("account_id"))
		// Now return the KafkaCluster with updated ApiEndpoint
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.ApiEndpoint = c.kafkaApiUrl
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
