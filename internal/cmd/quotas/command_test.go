package quotas

import (
	"encoding/json"
	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/quotas/v2"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	segment "github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"


	"github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

type QuotasTestSuite struct {
	suite.Suite
	conf              *v3.Config
	QuotasClient     *quotasv2.APIClient
	analyticsOutput   []segment.Message
	analyticsClient   analytics.Client
}

func TestQuotasTestSuite(t *testing.T) {
	suite.Run(t, new(QuotasTestSuite))
}

func (suite *QuotasTestSuite) SetupTest() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	suite.StartBackEndServer()
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
	url := suite.StartBackEndServer()
	cfg := quotasv2.NewConfiguration()

	cfg.Servers[0].URL = url+"/api"
	suite.QuotasClient = quotasv2.NewAPIClient(cfg)
	suite.StartBackEndServer()
}



func (suite *QuotasTestSuite) newCmd() *command {
	resolverMock := &pcmd.FlagResolverImpl{
		Out: os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver: resolverMock,
		QuotasClient: suite.QuotasClient,
		Config:       suite.conf,
	}
	return New("ccloud", prerunner, suite.analyticsClient)
}

func (suite *QuotasTestSuite) StartBackEndServer() string {
	r := mux.NewRouter()
	r.HandleFunc("/api/quotas/v2/applied-quotas", HandleAppliedQuotas())
	s := httptest.NewServer(r)
	return s.URL
}

func HandleAppliedQuotas() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		qt1 := quotasv2.QuotasV2AppliedQuota{
			Id:           stringToPtr("quota_a"),
			Scope:        stringToPtr("kafka_cluster"),
			DisplayName:  stringToPtr(""),
			Organization: quotasv2.NewObjectReference("org-123", "",""),
			KafkaCluster: quotasv2.NewObjectReference("lkc-1","",""),
			Environment: quotasv2.NewObjectReference("env-1","",""),
			AppliedLimit: int32ToPtr(15),
		}
		qtls :=  &quotasv2.QuotasV2AppliedQuotaList{
			ApiVersion: "quotas/v2",
			Kind: "AppliedQuotaList",
			Data: []quotasv2.QuotasV2AppliedQuota{qt1},
		}
		reply, _ := json.Marshal(qtls)
		io.WriteString(w, string(reply))
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
	err := utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)

}
