package kafka

import (
	"context"
	"net/http"
	"testing"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	metricsmock "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	climock "github.com/confluentinc/cli/mock"
)

const (
	clusterId     = "lkc-0000"
	clusterName   = "testCluster"
	cloudId       = "aws"
	regionId      = "us-west-2"
	environmentId = "abc"
)

var queryTime = time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC)

var shouldError bool
var shouldPrompt bool

var cmkByokCluster = cmkv2.CmkV2Cluster{
	Spec: &cmkv2.CmkV2ClusterSpec{
		Environment: &cmkv2.EnvScopedObjectReference{
			Id: environmentId,
		},
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

var cmkExpandCluster = cmkv2.CmkV2Cluster{
	Spec: &cmkv2.CmkV2ClusterSpec{
		Environment: &cmkv2.EnvScopedObjectReference{
			Id: environmentId,
		},
		DisplayName:  cmkv2.PtrString("gcp-shrink-test"),
		Cloud:        cmkv2.PtrString("gcp"),
		Region:       cmkv2.PtrString("us-central1"),
		Config:       setCmkClusterConfig("dedicated", 3, ""),
		Availability: cmkv2.PtrString(lowAvailability),
	},
	Id: cmkv2.PtrString("lkc-xyz"),
	Status: &cmkv2.CmkV2ClusterStatus{
		Cku:   cmkv2.PtrInt32(3),
		Phase: "PROVISIONED",
	},
}

type KafkaClusterTestSuite struct {
	suite.Suite
	conf            *v1.Config
	envMetadataMock *ccloudv1mock.EnvironmentMetadata
	metricsApi      *metricsmock.Version2Api
	cmkClusterApi   *cmkmock.ClustersCmkV2Api
}

func (suite *KafkaClusterTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.cmkClusterApi = &cmkmock.ClustersCmkV2Api{
		CreateCmkV2ClusterFunc: func(ctx context.Context) cmkv2.ApiCreateCmkV2ClusterRequest {
			return cmkv2.ApiCreateCmkV2ClusterRequest{}
		},
		CreateCmkV2ClusterExecuteFunc: func(req cmkv2.ApiCreateCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkByokCluster, nil, nil
		},
		GetCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
			return cmkv2.ApiGetCmkV2ClusterRequest{}
		},
		GetCmkV2ClusterExecuteFunc: func(req cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkByokCluster, nil, nil
		},
		DeleteCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiDeleteCmkV2ClusterRequest {
			return cmkv2.ApiDeleteCmkV2ClusterRequest{}
		},
		DeleteCmkV2ClusterExecuteFunc: func(req cmkv2.ApiDeleteCmkV2ClusterRequest) (*http.Response, error) {
			return nil, nil
		},
	}
	suite.envMetadataMock = &ccloudv1mock.EnvironmentMetadata{
		GetFunc: func(arg0 context.Context) (metadata []*ccloudv1.CloudMetadata, e error) {
			cloudMeta := &ccloudv1.CloudMetadata{
				Id: cloudId,
				Regions: []*ccloudv1.Region{
					{
						Id:            regionId,
						IsSchedulable: true,
					},
				},
			}
			return []*ccloudv1.CloudMetadata{
				cloudMeta,
			}, nil
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

func (suite *KafkaClusterTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	client := &ccloudv1.Client{
		EnvironmentMetadata: suite.envMetadataMock,
	}
	v2Client := &ccloudv2.Client{
		AuthToken:     "auth-token",
		CmkClient:     &cmkv2.APIClient{ClustersCmkV2Api: suite.cmkClusterApi},
		MetricsClient: &metricsv2.APIClient{Version2Api: suite.metricsApi},
	}
	prerunner := climock.NewPreRunnerMock(client, v2Client, nil, nil, conf)
	return newClusterCommand(conf, prerunner)
}

func (suite *KafkaClusterTestSuite) TestClusterShrinkShouldPrompt() {
	req := require.New(suite.T())
	suite.cmkClusterApi = &cmkmock.ClustersCmkV2Api{
		GetCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
			return cmkv2.ApiGetCmkV2ClusterRequest{}
		},
		GetCmkV2ClusterExecuteFunc: func(req cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkExpandCluster, nil, nil
		},
	}
	// Set variable for Metrics API mock
	shouldError = false
	shouldPrompt = true
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"update", clusterName, "--cku", "2"})
	err := cmd.Execute()
	req.Contains(err.Error(), "cluster resize error: failed to read your confirmation")
	req.True(suite.metricsApi.V2MetricsDatasetQueryPostCalled())
	req.True(suite.metricsApi.V2MetricsDatasetQueryPostExecuteCalled())
}

func (suite *KafkaClusterTestSuite) TestCreateKafkaCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"create", clusterName, "--cloud", cloudId, "--region", regionId})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.envMetadataMock.GetCalled())
	req.True(suite.cmkClusterApi.CreateCmkV2ClusterCalled())
}

func (suite *KafkaClusterTestSuite) TestDeleteKafkaCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"delete", clusterId, "--force"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.cmkClusterApi.DeleteCmkV2ClusterCalled())
}

func (suite *KafkaClusterTestSuite) TestGetLkcForDescribe() {
	req := require.New(suite.T())
	cmd := new(cobra.Command)
	cfg := v1.AuthenticatedCloudConfigMock()
	prerunner := &pcmd.PreRun{Config: cfg}
	c := &clusterCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.Config = dynamicconfig.New(cfg, nil, nil)
	lkc, err := c.getLkcForDescribe([]string{"lkc-123"})
	req.Equal("lkc-123", lkc)
	req.NoError(err)
	lkc, err = c.getLkcForDescribe([]string{})
	req.Equal(c.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId(), lkc)
	req.NoError(err)
	c.Config.Context().KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster = ""
	lkc, err = c.getLkcForDescribe([]string{})
	req.Equal("", lkc)
	req.Equal(errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaForDescribeSuggestions).Error(), err.Error())
}

func TestKafkaClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClusterTestSuite))
}
