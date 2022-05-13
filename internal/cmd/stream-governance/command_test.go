package streamgovernance

import (
	"context"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	sgmock "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2/mock"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

const (
	id                = "lsrc-1234"
	httpEndpoint      = "https://sr-endpoint"
	status            = "PROVISIONED"
	packageType       = "advanced"
	regionId          = "sgreg=1"
	cloud             = "aws"
	regionName        = "us-east-2"
	regionDisplayName = "Ohio (us-east-2)"
)

type StreamGovernanceTestSuite struct {
	suite.Suite
	conf            *v1.Config
	kafkaCluster    *schedv1.KafkaCluster
	envMetadataMock *ccsdkmock.EnvironmentMetadata
	sgClusterApi    *sgmock.MockClustersStreamGovernanceV2Api
	sgRegionApi     *sgmock.MockRegionsStreamGovernanceV2Api
	prompt          *mock.Prompt
}

func (suite *StreamGovernanceTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:       cluster.ID,
		Name:     cluster.Name,
		Endpoint: cluster.APIEndpoint,
	}

	suite.sgClusterApi = suite.getMockClustersStreamGovernanceV2Api()
	suite.sgRegionApi = suite.getMockRegionsStreamGovernanceV2Api()

	suite.envMetadataMock = &ccsdkmock.EnvironmentMetadata{
		GetFunc: func(arg0 context.Context) (metadata []*schedv1.CloudMetadata, e error) {
			cloudMeta := &schedv1.CloudMetadata{
				Id: cloud,
				Regions: []*schedv1.Region{
					{
						Id:            regionName,
						IsSchedulable: true,
					},
				},
			}
			return []*schedv1.CloudMetadata{
				cloudMeta,
			}, nil
		},
	}
}

func (suite *StreamGovernanceTestSuite) getMockClustersStreamGovernanceV2Api() *sgmock.MockClustersStreamGovernanceV2Api {
	return &sgmock.MockClustersStreamGovernanceV2Api{
		CreateStreamGovernanceV2ClusterFunc: func(ctx context.Context) sgsdk.ApiCreateStreamGovernanceV2ClusterRequest {
			return sgsdk.ApiCreateStreamGovernanceV2ClusterRequest{
				ApiService: suite.sgClusterApi,
			}
		},
		CreateStreamGovernanceV2ClusterExecuteFunc: func(req sgsdk.ApiCreateStreamGovernanceV2ClusterRequest) (sgsdk.StreamGovernanceV2Cluster, *http.Response, error) {
			return getStreamGovernanceCluster(id, packageType, httpEndpoint, v1.MockEnvironmentId, regionId, status), nil, nil
		},
		GetStreamGovernanceV2ClusterFunc: func(ctx context.Context, _ string) sgsdk.ApiGetStreamGovernanceV2ClusterRequest {
			return sgsdk.ApiGetStreamGovernanceV2ClusterRequest{
				ApiService: suite.sgClusterApi,
			}
		},
		GetStreamGovernanceV2ClusterExecuteFunc: func(req sgsdk.ApiGetStreamGovernanceV2ClusterRequest) (sgsdk.StreamGovernanceV2Cluster, *http.Response, error) {
			return getStreamGovernanceCluster(id, packageType, httpEndpoint, v1.MockEnvironmentId, regionId, status), nil, nil

		},
		ListStreamGovernanceV2ClustersFunc: func(ctx context.Context) sgsdk.ApiListStreamGovernanceV2ClustersRequest {
			return sgsdk.ApiListStreamGovernanceV2ClustersRequest{
				ApiService: suite.sgClusterApi,
			}
		},
		ListStreamGovernanceV2ClustersExecuteFunc: func(req sgsdk.ApiListStreamGovernanceV2ClustersRequest) (sgsdk.StreamGovernanceV2ClusterList, *http.Response, error) {
			return sgsdk.StreamGovernanceV2ClusterList{
				Data: []sgsdk.StreamGovernanceV2Cluster{
					getStreamGovernanceCluster(id, packageType, httpEndpoint, v1.MockEnvironmentId, regionId, status)},
			}, nil, nil
		},
		UpdateStreamGovernanceV2ClusterFunc: func(ctx context.Context, _ string) sgsdk.ApiUpdateStreamGovernanceV2ClusterRequest {
			return sgsdk.ApiUpdateStreamGovernanceV2ClusterRequest{
				ApiService: suite.sgClusterApi,
			}
		},
		UpdateStreamGovernanceV2ClusterExecuteFunc: func(req sgsdk.ApiUpdateStreamGovernanceV2ClusterRequest) (sgsdk.StreamGovernanceV2Cluster, *http.Response, error) {
			return getStreamGovernanceCluster(id, packageType, httpEndpoint, v1.MockEnvironmentId, regionId, status), nil, nil
		},
		DeleteStreamGovernanceV2ClusterFunc: func(ctx context.Context, _ string) sgsdk.ApiDeleteStreamGovernanceV2ClusterRequest {
			return sgsdk.ApiDeleteStreamGovernanceV2ClusterRequest{}
		},
		DeleteStreamGovernanceV2ClusterExecuteFunc: func(req sgsdk.ApiDeleteStreamGovernanceV2ClusterRequest) (*http.Response, error) {
			return nil, nil
		},
	}
}

func (suite *StreamGovernanceTestSuite) getMockRegionsStreamGovernanceV2Api() *sgmock.MockRegionsStreamGovernanceV2Api {
	return &sgmock.MockRegionsStreamGovernanceV2Api{
		GetStreamGovernanceV2RegionFunc: func(ctx context.Context, _ string) sgsdk.ApiGetStreamGovernanceV2RegionRequest {
			return sgsdk.ApiGetStreamGovernanceV2RegionRequest{
				ApiService: suite.sgRegionApi,
			}
		},
		GetStreamGovernanceV2RegionExecuteFunc: func(req sgsdk.ApiGetStreamGovernanceV2RegionRequest) (sgsdk.StreamGovernanceV2Region, *http.Response, error) {
			return getStreamGovernanceRegion(regionId, regionName, cloud, packageType, regionDisplayName), nil, nil
		},
		ListStreamGovernanceV2RegionsFunc: func(ctx context.Context) sgsdk.ApiListStreamGovernanceV2RegionsRequest {
			return sgsdk.ApiListStreamGovernanceV2RegionsRequest{
				ApiService: suite.sgRegionApi,
			}
		},
		ListStreamGovernanceV2RegionsExecuteFunc: func(r sgsdk.ApiListStreamGovernanceV2RegionsRequest) (sgsdk.StreamGovernanceV2RegionList, *http.Response, error) {
			return sgsdk.StreamGovernanceV2RegionList{
				Data: []sgsdk.StreamGovernanceV2Region{
					getStreamGovernanceRegion(regionId, regionName, cloud, packageType, regionDisplayName)},
			}, nil, nil
		},
	}
}

func getStreamGovernanceCluster(id, packageType, endpoint, envId, regionId, status string) sgsdk.StreamGovernanceV2Cluster {
	return sgsdk.StreamGovernanceV2Cluster{
		Id: &id,
		Spec: &sgsdk.StreamGovernanceV2ClusterSpec{
			DisplayName:  &id,
			Package:      &packageType,
			HttpEndpoint: &endpoint,
			Environment: &sgsdk.ObjectReference{
				Id: envId,
			},
			Region: &sgsdk.ObjectReference{
				Id: regionId,
			},
		},
		Status: &sgsdk.StreamGovernanceV2ClusterStatus{
			Phase: status,
		},
	}
}

func getStreamGovernanceRegion(id, region, cloud, packageType, displayName string) sgsdk.StreamGovernanceV2Region {
	return sgsdk.StreamGovernanceV2Region{
		Id: &id,
		Spec: &sgsdk.StreamGovernanceV2RegionSpec{
			RegionName:  &region,
			Cloud:       &cloud,
			Packages:    &[]string{packageType},
			DisplayName: &displayName,
		},
	}
}

func (suite *StreamGovernanceTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	client := &ccloud.Client{
		EnvironmentMetadata: suite.envMetadataMock,
	}
	sgClient := &sgsdk.APIClient{
		ClustersStreamGovernanceV2Api: suite.sgClusterApi,
		RegionsStreamGovernanceV2Api:  suite.sgRegionApi,
	}

	prerunner := cliMock.NewPreRunnerMock(client, &ccloudv2.Client{StreamGovernanceClient: sgClient, AuthToken: "auth-token"}, nil, nil, conf)
	return New(conf, prerunner)
}

func (suite *StreamGovernanceTestSuite) TestDescribeStreamGovernanceCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"describe"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.sgClusterApi.GetStreamGovernanceV2ClusterCalled())
}

func (suite *StreamGovernanceTestSuite) TestEnableStreamGovernanceCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"enable", "--cloud", cloud, "--package", packageType, "--region", regionName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.envMetadataMock.GetCalled())
	req.True(suite.sgClusterApi.CreateStreamGovernanceV2ClusterCalled())
}

func (suite *StreamGovernanceTestSuite) TestUpdateStreamGovernanceCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"upgrade", "--package", packageType})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.sgClusterApi.UpdateStreamGovernanceV2ClusterCalled())
}

func (suite *StreamGovernanceTestSuite) TestDeleteStreamGovernanceClusterShouldPrompt() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"delete"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Contains(err.Error(), errors.SGFailedToReadDeletionConfirmationErrorMsg)
}

func TestStreamGovernanceTestSuite(t *testing.T) {
	suite.Run(t, new(StreamGovernanceTestSuite))
}
