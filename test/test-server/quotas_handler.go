package test_server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	qtv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/quotas/v2"
	"github.com/stretchr/testify/require"
)

func (c *CloudRouter) HandleAppliedQuotas(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		environment := r.URL.Query().Get("environment")
		kafkaCluster := r.URL.Query().Get("kafka_cluster")
		network := r.URL.Query().Get("network")

		quota1 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_a"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Quota A"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-1", "", ""),
			Environment:  qtv2.NewObjectReference("env-1", "", ""),
			AppliedLimit: int32ToPtr(15),
		}

		quota2 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_a"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Qutoa A"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-2", "", ""),
			Environment:  qtv2.NewObjectReference("env-2", "", ""),
			AppliedLimit: int32ToPtr(16),
		}

		quota3 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_b"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Quota B"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-1", "", ""),
			Environment:  qtv2.NewObjectReference("env-1", "", ""),
			AppliedLimit: int32ToPtr(17),
		}

		quota4 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_b"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Quota B"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-2", "", ""),
			Environment:  qtv2.NewObjectReference("env-2", "", ""),
			AppliedLimit: int32ToPtr(18),
		}

		filteredData := filterQuotaResults([]qtv2.QuotasV2AppliedQuota{quota1, quota2, quota3, quota4}, environment, network, kafkaCluster)
		quotaList := &qtv2.QuotasV2AppliedQuotaList{
			ApiVersion: "quotas/v2",
			Kind:       "AppliedQuotaList",
			Data:       filteredData,
		}

		reply, err := json.Marshal(quotaList)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}

func stringToPtr(s string) *string {
	return &s
}

func int32ToPtr(i int32) *int32 {
	return &i
}

func filterQuotaResults(quotaList []qtv2.QuotasV2AppliedQuota, environment string, network string, kafkaCluster string) []qtv2.QuotasV2AppliedQuota {

	//filter by environment id
	filtered := []qtv2.QuotasV2AppliedQuota{}
	if environment != "" {
		for _, qt := range quotaList {
			if qt.Environment != nil && qt.Environment.Id == environment {
				filtered = append(filtered, qt)
			}
		}
		quotaList = filtered
	}

	//filter by cluster id
	filtered = []qtv2.QuotasV2AppliedQuota{}
	if kafkaCluster != "" {
		for _, qt := range quotaList {
			if qt.KafkaCluster != nil && qt.KafkaCluster.Id == kafkaCluster {
				filtered = append(filtered, qt)
			}
		}
		quotaList = filtered
	}

	//filter by network id
	filtered = []qtv2.QuotasV2AppliedQuota{}
	if network != "" {
		for _, qt := range quotaList {
			if qt.Network != nil && qt.Network.Id == network {
				filtered = append(filtered, qt)
			}
		}
		quotaList = filtered
	}

	return quotaList
}
