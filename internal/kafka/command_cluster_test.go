package kafka

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	metricsmock "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2/mock"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

const (
	cloudId       = "aws"
	regionId      = "us-west-2"
	environmentId = "abc"
)

var queryTime = time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC)

var (
	shouldError  bool
	shouldPrompt bool
)

var cmkByokCluster = cmkv2.CmkV2Cluster{
	Spec: &cmkv2.CmkV2ClusterSpec{
		Environment:  &cmkv2.EnvScopedObjectReference{Id: environmentId},
		DisplayName:  cmkv2.PtrString("gcp-byok-test"),
		Cloud:        cmkv2.PtrString("gcp"),
		Region:       cmkv2.PtrString("us-central1"),
		Config:       setCmkClusterConfig("dedicated", 1, "xyz"),
		Availability: cmkv2.PtrString(lowAvailability),
	},
	Id: cmkv2.PtrString("lkc-xyz"),
	Status: &cmkv2.CmkV2ClusterStatus{
		Cku:   cmkv2.PtrInt32(1),
		Phase: ccloudv2.StatusProvisioning,
	},
}

type KafkaClusterTestSuite struct {
	suite.Suite
	conf            *config.Config
	envMetadataMock *ccloudv1mock.EnvironmentMetadata
	metricsApi      *metricsmock.Version2Api
	cmkClusterApi   *cmkmock.ClustersCmkV2Api
}

func (suite *KafkaClusterTestSuite) SetupTest() {
	suite.conf = config.AuthenticatedCloudConfigMock()
	suite.cmkClusterApi = &cmkmock.ClustersCmkV2Api{
		CreateCmkV2ClusterFunc: func(_ context.Context) cmkv2.ApiCreateCmkV2ClusterRequest {
			return cmkv2.ApiCreateCmkV2ClusterRequest{}
		},
		CreateCmkV2ClusterExecuteFunc: func(_ cmkv2.ApiCreateCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkByokCluster, nil, nil
		},
		GetCmkV2ClusterFunc: func(_ context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
			return cmkv2.ApiGetCmkV2ClusterRequest{}
		},
		GetCmkV2ClusterExecuteFunc: func(_ cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkByokCluster, nil, nil
		},
		DeleteCmkV2ClusterFunc: func(_ context.Context, _ string) cmkv2.ApiDeleteCmkV2ClusterRequest {
			return cmkv2.ApiDeleteCmkV2ClusterRequest{}
		},
		DeleteCmkV2ClusterExecuteFunc: func(_ cmkv2.ApiDeleteCmkV2ClusterRequest) (*http.Response, error) {
			return nil, nil
		},
	}
	suite.envMetadataMock = &ccloudv1mock.EnvironmentMetadata{
		GetFunc: func() ([]*ccloudv1.CloudMetadata, error) {
			cloudMeta := &ccloudv1.CloudMetadata{
				Id: cloudId,
				Regions: []*ccloudv1.Region{{
					Id:            regionId,
					IsSchedulable: true,
				}},
			}
			return []*ccloudv1.CloudMetadata{cloudMeta}, nil
		},
	}
	suite.metricsApi = &metricsmock.Version2Api{
		V2MetricsDatasetQueryPostFunc: func(_ context.Context, _ string) metricsv2.ApiV2MetricsDatasetQueryPostRequest {
			return metricsv2.ApiV2MetricsDatasetQueryPostRequest{}
		},
		V2MetricsDatasetQueryPostExecuteFunc: func(_ metricsv2.ApiV2MetricsDatasetQueryPostRequest) (*metricsv2.QueryResponse, *http.Response, error) {
			value := float32(0.1)
			if shouldPrompt {
				value = 0.8
			}
			if shouldError {
				value = 5000
			}
			resp := &metricsv2.QueryResponse{
				FlatQueryResponse: &metricsv2.FlatQueryResponse{
					Data: []metricsv2.Point{
						{Value: value, Timestamp: queryTime},
					},
				},
			}
			return resp, nil, nil
		},
	}
}

func (suite *KafkaClusterTestSuite) TestGetLkcForDescribe() {
	req := require.New(suite.T())
	cmd := new(cobra.Command)
	cfg := config.AuthenticatedCloudConfigMock()
	prerunner := &pcmd.PreRun{Config: cfg}
	c := &clusterCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.Context = dynamicconfig.NewDynamicContext(cfg.Context())
	lkc, err := c.getLkcForDescribe([]string{"lkc-123"})
	req.Equal("lkc-123", lkc)
	req.NoError(err)
	lkc, err = c.getLkcForDescribe([]string{})
	req.Equal(c.Context.KafkaClusterContext.GetActiveKafkaClusterId(), lkc)
	req.NoError(err)
	c.Context.KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster = ""
	lkc, err = c.getLkcForDescribe([]string{})
	req.Equal("", lkc)
	req.Equal(errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaForDescribeSuggestions).Error(), err.Error())
}

func TestKafkaClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClusterTestSuite))
}
