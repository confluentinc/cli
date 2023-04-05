package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
)

func handleAppliedQuotas(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environment := r.URL.Query().Get("environment")
		kafkaCluster := r.URL.Query().Get("kafka_cluster")
		network := r.URL.Query().Get("network")
		quotaCode := r.URL.Query().Get("id")

		quota1 := servicequotav1.ServiceQuotaV1AppliedQuota{
			Id:           ptrObj("quota_a"),
			Scope:        ptrObj("kafka_cluster"),
			DisplayName:  ptrObj("Quota A"),
			Organization: servicequotav1.NewObjectReference("org-123", "", ""),
			KafkaCluster: servicequotav1.NewObjectReference("lkc-1", "", ""),
			Environment:  servicequotav1.NewObjectReference("env-1", "", ""),
			Usage:        ptrObj(int32(10)),
			AppliedLimit: ptrObj(int32(15)),
		}

		quota2 := servicequotav1.ServiceQuotaV1AppliedQuota{
			Id:           ptrObj("quota_a"),
			Scope:        ptrObj("kafka_cluster"),
			DisplayName:  ptrObj("Qutoa A"),
			Organization: servicequotav1.NewObjectReference("org-123", "", ""),
			KafkaCluster: servicequotav1.NewObjectReference("lkc-2", "", ""),
			Environment:  servicequotav1.NewObjectReference("env-2", "", ""),
			Usage:        ptrObj(int32(11)),
			AppliedLimit: ptrObj(int32(16)),
		}

		quota3 := servicequotav1.ServiceQuotaV1AppliedQuota{
			Id:           ptrObj("quota_b"),
			Scope:        ptrObj("kafka_cluster"),
			DisplayName:  ptrObj("Quota B"),
			Organization: servicequotav1.NewObjectReference("org-123", "", ""),
			KafkaCluster: servicequotav1.NewObjectReference("lkc-1", "", ""),
			Environment:  servicequotav1.NewObjectReference("env-1", "", ""),
			AppliedLimit: ptrObj(int32(17)),
		}

		quota4 := servicequotav1.ServiceQuotaV1AppliedQuota{
			Id:           ptrObj("quota_b"),
			Scope:        ptrObj("kafka_cluster"),
			DisplayName:  ptrObj("Quota B"),
			Organization: servicequotav1.NewObjectReference("org-123", "", ""),
			KafkaCluster: servicequotav1.NewObjectReference("lkc-2", "", ""),
			Environment:  servicequotav1.NewObjectReference("env-2", "", ""),
			AppliedLimit: ptrObj(int32(18)),
		}

		filteredData := filterQuotaResults([]servicequotav1.ServiceQuotaV1AppliedQuota{quota1, quota2, quota3, quota4}, environment, network, kafkaCluster, quotaCode)
		quotaList := &servicequotav1.ServiceQuotaV1AppliedQuotaList{
			ApiVersion: "service-quota/v1",
			Kind:       "AppliedQuotaList",
			Data:       filteredData,
		}

		reply, err := json.Marshal(quotaList)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}

func filterQuotaResults(quotaList []servicequotav1.ServiceQuotaV1AppliedQuota, environment string, network string, kafkaCluster string, quotaCode string) []servicequotav1.ServiceQuotaV1AppliedQuota {
	// filter by environment id
	filtered := []servicequotav1.ServiceQuotaV1AppliedQuota{}
	if environment != "" {
		for _, quota := range quotaList {
			if quota.Environment != nil && quota.Environment.Id == environment {
				filtered = append(filtered, quota)
			}
		}
		quotaList = filtered
	}

	// filter by cluster id
	filtered = []servicequotav1.ServiceQuotaV1AppliedQuota{}
	if kafkaCluster != "" {
		for _, quota := range quotaList {
			if quota.KafkaCluster != nil && quota.KafkaCluster.Id == kafkaCluster {
				filtered = append(filtered, quota)
			}
		}
		quotaList = filtered
	}

	// filter by network id
	filtered = []servicequotav1.ServiceQuotaV1AppliedQuota{}
	if network != "" {
		for _, quota := range quotaList {
			if quota.Network != nil && quota.Network.Id == network {
				filtered = append(filtered, quota)
			}
		}
		quotaList = filtered
	}

	// filter by quota code (id)
	filtered = []servicequotav1.ServiceQuotaV1AppliedQuota{}
	if quotaCode != "" {
		for _, quota := range quotaList {
			if quota.Id != nil && *quota.Id == quotaCode {
				filtered = append(filtered, quota)
			}
		}
		quotaList = filtered
	}

	return quotaList
}
