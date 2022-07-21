package schemaregistry

import (
	"context"
	"net/http"
	"testing"
	"time"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	metricsmock "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	srClusterID = "sr"
)

var queryTime = time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC)

type ClusterTestSuite struct {
	suite.Suite
	conf         *v1.Config
	kafkaCluster *schedv1.KafkaCluster
	srCluster    *schedv1.SchemaRegistryCluster
	srMock       *ccsdkmock.SchemaRegistry
	srClientMock *srsdk.APIClient
	metricsApi   *metricsmock.Version2Api
}

func (suite *ClusterTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Endpoint:   cluster.APIEndpoint,
		Enterprise: true,
	}
	suite.srCluster = &schedv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			GetTopLevelConfigFunc: func(ctx context.Context) (srsdk.Config, *http.Response, error) {
				return srsdk.Config{CompatibilityLevel: "FULL"}, nil, nil
			},
			GetTopLevelModeFunc: func(ctx context.Context) (srsdk.Mode, *http.Response, error) {
				return srsdk.Mode{}, nil, nil
			},
			UpdateTopLevelModeFunc: func(ctx context.Context, body srsdk.ModeUpdateRequest) (request srsdk.ModeUpdateRequest, response *http.Response, e error) {
				return srsdk.ModeUpdateRequest{Mode: body.Mode}, nil, nil
			},
			UpdateTopLevelConfigFunc: func(ctx context.Context, body srsdk.ConfigUpdateRequest) (request srsdk.ConfigUpdateRequest, response *http.Response, e error) {
				return srsdk.ConfigUpdateRequest{Compatibility: body.Compatibility}, nil, nil
			},
		},
	}
}

func (suite *ClusterTestSuite) SetupTest() {
	suite.srMock = &ccsdkmock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *schedv1.SchemaRegistryClusterConfig) (*schedv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClustersFunc: func(ctx context.Context, clusterConfig *schedv1.SchemaRegistryCluster) ([]*schedv1.SchemaRegistryCluster, error) {
			return []*schedv1.SchemaRegistryCluster{suite.srCluster}, nil
		},
	}
	suite.metricsApi = &metricsmock.Version2Api{
		V2MetricsDatasetQueryPostFunc: func(_ context.Context, _ string) metricsv2.ApiV2MetricsDatasetQueryPostRequest {
			return metricsv2.ApiV2MetricsDatasetQueryPostRequest{}
		},
		V2MetricsDatasetQueryPostExecuteFunc: func(_ metricsv2.ApiV2MetricsDatasetQueryPostRequest) (*metricsv2.QueryResponse, *http.Response, error) {
			resp := &metricsv2.QueryResponse{
				FlatQueryResponse: &metricsv2.FlatQueryResponse{
					Data: []metricsv2.Point{
						{Value: 0.0, Timestamp: queryTime},
					},
				},
			}
			return resp, nil, nil
		},
	}
}

func (suite *ClusterTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		SchemaRegistry: suite.srMock,
	}
	v2Client := &ccloudv2.Client{
		AuthToken:     "auth-token",
		MetricsClient: &metricsv2.APIClient{Version2Api: suite.metricsApi},
	}
	return New(suite.conf, cliMock.NewPreRunnerMock(client, v2Client, nil, nil, suite.conf), suite.srClientMock)
}

func (suite *ClusterTestSuite) TestCreateSR() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "enable", "--cloud", "aws", "--geo", "us"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.srMock.CreateSchemaRegistryClusterCalled())
}

func (suite *ClusterTestSuite) TestDescribeSR() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "describe"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.srMock.GetSchemaRegistryClustersCalled())
	req.True(suite.metricsApi.V2MetricsDatasetQueryPostCalled())
	req.True(suite.metricsApi.V2MetricsDatasetQueryPostExecuteCalled())
}

func (suite *ClusterTestSuite) TestUpdateCompatibility() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "update", "--compatibility", "BACKWARD"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.UpdateTopLevelConfigCalled())
	retVal := apiMock.UpdateTopLevelConfigCalls()[0]
	req.Equal(retVal.Body.Compatibility, "BACKWARD")
}

func (suite *ClusterTestSuite) TestUpdateMode() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "update", "--mode", "READWRITE"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.UpdateTopLevelModeCalled())
	retVal := apiMock.UpdateTopLevelModeCalls()[0]
	req.Equal(retVal.Body.Mode, "READWRITE")
}

func (suite *ClusterTestSuite) TestUpdateNoArgs() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "update"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err, "flag string not set")
}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}
