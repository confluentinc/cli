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
		qt1 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_a"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Quota A"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-1", "", ""),
			Environment:  qtv2.NewObjectReference("env-1", "", ""),
			AppliedLimit: int32ToPtr(15),
		}

		qt2 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_a"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Qutoa A"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-2", "", ""),
			Environment:  qtv2.NewObjectReference("env-2", "", ""),
			AppliedLimit: int32ToPtr(16),
		}

		qt3 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_b"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Quota B"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-1", "", ""),
			Environment:  qtv2.NewObjectReference("env-1", "", ""),
			AppliedLimit: int32ToPtr(17),
		}

		qt4 := qtv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_b"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr("Quota B"),
			Organization: qtv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: qtv2.NewObjectReference("lkc-2", "", ""),
			Environment:  qtv2.NewObjectReference("env-2", "", ""),
			AppliedLimit: int32ToPtr(18),
		}

		qtls := &qtv2.QuotasV2AppliedQuotaList{
			ApiVersion: "quotas/v2",
			Kind:       "AppliedQuotaList",
			Data:       []qtv2.QuotasV2AppliedQuota{qt1, qt2, qt3, qt4},
		}

		reply, err := json.Marshal(qtls)
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
