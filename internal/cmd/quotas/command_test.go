package quotas

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/quotas/v2"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	segment "github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

type QuotasTestSuite struct {
	suite.Suite
	conf            *v1.Config
	QuotasClient    *quotasv2.APIClient
	analyticsOutput []segment.Message
	analyticsClient analytics.Client
}

func TestQuotasTestSuite(t *testing.T) {
	suite.Run(t, new(QuotasTestSuite))
}

func (suite *QuotasTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.StartBackEndServer()
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
	url := suite.StartBackEndServer()
	cfg := quotasv2.NewConfiguration()

	cfg.Servers[0].URL = url + "/api"
	suite.QuotasClient = quotasv2.NewAPIClient(cfg)
	suite.StartBackEndServer()
}

func (suite *QuotasTestSuite) newCmd() *cobra.Command {
	resolverMock := &pcmd.FlagResolverImpl{
		Out: os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver: resolverMock,
		QuotasClient: suite.QuotasClient,
		Config:       suite.conf,
	}
	return New(prerunner)
}

func (suite *QuotasTestSuite) StartBackEndServer() string {
	r := mux.NewRouter()
	r.HandleFunc("/api/quotas/v2/applied-quotas", HandleAppliedQuotas(suite.T()))
	s := httptest.NewServer(r)
	return s.URL
}

func HandleAppliedQuotas(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		quota1 := quotasv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_a"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr(""),
			Organization: quotasv2.NewObjectReference("org-123", "", ""),
			KafkaCluster: quotasv2.NewObjectReference("lkc-1", "", ""),
			Environment:  quotasv2.NewObjectReference("env-1", "", ""),
			AppliedLimit: int32ToPtr(15),
		}
		quotaList := &quotasv2.QuotasV2AppliedQuotaList{
			ApiVersion: "quotas/v2",
			Kind:       "AppliedQuotaList",
			Data:       []quotasv2.QuotasV2AppliedQuota{quota1},
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

func (suite *QuotasTestSuite) TestListQuotas() {
	cmd := suite.newCmd()
	args := []string{"list", "kafka_cluster"}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)

}

func (suite *QuotasTestSuite) TestFilterQuotasFunc() {
	t := suite.T()
	quota1 := quotasv2.QuotasV2AppliedQuota{
		Id:           stringToPtr("quota_a"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Quota A"),
		Organization: quotasv2.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv2.NewObjectReference("lkc-1", "", ""),
		Environment:  quotasv2.NewObjectReference("env-1", "", ""),
		Network:      quotasv2.NewObjectReference("n-1", "", ""),
		AppliedLimit: int32ToPtr(15),
	}

	quota2 := quotasv2.QuotasV2AppliedQuota{
		Id:           stringToPtr("quota_a"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Qutoa A"),
		Organization: quotasv2.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv2.NewObjectReference("lkc-2", "", ""),
		Environment:  quotasv2.NewObjectReference("env-2", "", ""),
		AppliedLimit: int32ToPtr(16),
	}

	quota3 := quotasv2.QuotasV2AppliedQuota{
		Id:           stringToPtr("quota_b"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Quota B"),
		Organization: quotasv2.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv2.NewObjectReference("lkc-1", "", ""),
		Environment:  quotasv2.NewObjectReference("env-1", "", ""),
		AppliedLimit: int32ToPtr(17),
	}

	quota4 := quotasv2.QuotasV2AppliedQuota{
		Id:           stringToPtr("quota_b"),
		Scope:        stringToPtr("kafka_cluster"),
		DisplayName:  stringToPtr("Quota B"),
		Organization: quotasv2.NewObjectReference("org-123", "", ""),
		KafkaCluster: quotasv2.NewObjectReference("lkc-2", "", ""),
		Environment:  quotasv2.NewObjectReference("env-2", "", ""),
		AppliedLimit: int32ToPtr(18),
	}

	quotaList := []quotasv2.QuotasV2AppliedQuota{quota1, quota2, quota3, quota4}

	type test struct {
		name               string
		filterQuotaCode    string
		filterEnvironment  string
		filterNetwork      string
		filterKafkaCluster string
		originData         []quotasv2.QuotasV2AppliedQuota
		expectedData       []quotasv2.QuotasV2AppliedQuota
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
			expectedData:    []quotasv2.QuotasV2AppliedQuota{quota1, quota2},
		},
		{
			name:              "Filter by environment",
			filterEnvironment: "env-1",
			originData:        quotaList,
			expectedData:      []quotasv2.QuotasV2AppliedQuota{quota1, quota3},
		},
		{
			name:               "Filter by kafka cluster",
			filterKafkaCluster: "lkc-1",
			originData:         quotaList,
			expectedData:       []quotasv2.QuotasV2AppliedQuota{quota1, quota3},
		},
		{
			name:          "Filter by network",
			filterNetwork: "n-1",
			originData:    quotaList,
			expectedData:  []quotasv2.QuotasV2AppliedQuota{quota1},
		},
	}

	for _, test := range tests {
		filterResult := filterQuotaResults(test.originData, test.filterQuotaCode, test.filterEnvironment, test.filterNetwork, test.filterKafkaCluster)
		require.Equal(t, test.expectedData, filterResult)
	}

}
