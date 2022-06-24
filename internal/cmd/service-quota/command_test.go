package servicequota

import (
	"testing"

	quotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type QuotasTestSuite struct {
	suite.Suite
	QuotasClient *quotasv1.APIClient
}

func TestQuotasTestSuite(t *testing.T) {
	suite.Run(t, new(QuotasTestSuite))
}

func stringToPtr(s string) *string {
	return &s
}

func int32ToPtr(i int32) *int32 {
	return &i
}

func (suite *QuotasTestSuite) TestFilterQuotasFunc() {
	t := suite.T()
	quota1 := quotasv1.ServiceQuotaV1AppliedQuota{
		Id:           stringToPtr("quota_a"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Quota A"),
		Organization: quotasv1.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv1.NewObjectReference("lkc-1", "", ""),
		Environment:  quotasv1.NewObjectReference("env-1", "", ""),
		Network:      quotasv1.NewObjectReference("n-1", "", ""),
		AppliedLimit: int32ToPtr(15),
	}

	quota2 := quotasv1.ServiceQuotaV1AppliedQuota{
		Id:           stringToPtr("quota_a"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Qutoa A"),
		Organization: quotasv1.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv1.NewObjectReference("lkc-2", "", ""),
		Environment:  quotasv1.NewObjectReference("env-2", "", ""),
		AppliedLimit: int32ToPtr(16),
	}

	quota3 := quotasv1.ServiceQuotaV1AppliedQuota{
		Id:           stringToPtr("quota_b"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Quota B"),
		Organization: quotasv1.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv1.NewObjectReference("lkc-1", "", ""),
		Environment:  quotasv1.NewObjectReference("env-1", "", ""),
		AppliedLimit: int32ToPtr(17),
	}

	quota4 := quotasv1.ServiceQuotaV1AppliedQuota{
		Id:           stringToPtr("quota_b"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Quota B"),
		Organization: quotasv1.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv1.NewObjectReference("lkc-2", "", ""),
		Environment:  quotasv1.NewObjectReference("env-2", "", ""),
		AppliedLimit: int32ToPtr(18),
	}

	quotaList := []quotasv1.ServiceQuotaV1AppliedQuota{quota1, quota2, quota3, quota4}

	type test struct {
		name            string
		filterQuotaCode string
		originData      []quotasv1.ServiceQuotaV1AppliedQuota
		expectedData    []quotasv1.ServiceQuotaV1AppliedQuota
	}

	tests := []*test{
		{
			name:         "No filter",
			originData:   quotaList,
			expectedData: quotaList,
		},
		{
			name:            "Filter by quota code",
			filterQuotaCode: "quota_a",
			originData:      quotaList,
			expectedData:    []quotasv1.ServiceQuotaV1AppliedQuota{quota1, quota2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filterResult := filterQuotaResults(test.originData, test.filterQuotaCode)
			require.Equal(t, test.expectedData, filterResult)
		})
	}

}
