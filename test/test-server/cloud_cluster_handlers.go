package test_server

import (
	"io"
	"net/http"
	"testing"

	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
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
