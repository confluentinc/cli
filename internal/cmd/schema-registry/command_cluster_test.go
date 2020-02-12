package schema_registry

import (
	"context"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	metricsv1 "github.com/confluentinc/ccloudapis/metrics/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	srClusterID = "sr"
)

type ClusterTestSuite struct {
	suite.Suite
	conf         *v2.Config
	kafkaCluster *kafkav1.KafkaCluster
	srCluster    *srv1.SchemaRegistryCluster
	srMock       *mock.SchemaRegistry
	srClientMock *srsdk.APIClient
	metrics      *ccsdkmock.Metrics
	logger       *log.Logger
}

func (suite *ClusterTestSuite) SetupSuite() {
	suite.conf = v2.AuthenticatedConfigMock()
	ctx := suite.conf.Context()
	cluster := ctx.KafkaClusters[ctx.Kafka]
	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Endpoint:   cluster.APIEndpoint,
		Enterprise: true,
	}
	suite.srCluster = &srv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			GetTopLevelConfigFunc: func(ctx context.Context) (srsdk.Config, *http.Response, error) {
				return srsdk.Config{CompatibilityLevel: "FULL"}, nil, nil
			},
			GetTopLevelModeFunc: func(ctx context.Context) (srsdk.ModeGetResponse, *http.Response, error) {
				return srsdk.ModeGetResponse{}, nil, nil
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
	suite.srMock = &mock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryClusterConfig) (*srv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClustersFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryCluster) ([]*srv1.SchemaRegistryCluster, error) {
			return []*srv1.SchemaRegistryCluster{suite.srCluster}, nil
		},
	}
	suite.metrics = &ccsdkmock.Metrics{
		SchemaRegistryMetricsFunc: func(arg0 context.Context, arg1 string) (*metricsv1.SchemaRegistryMetric, error) {
			return &metricsv1.SchemaRegistryMetric{
				NumSchemas: 8,
			}, nil
		},
	}
}

func (suite *ClusterTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		SchemaRegistry: suite.srMock,
		Metrics:        suite.metrics,
	}
	cmd := New(cliMock.NewPreRunnerMock(client, nil), suite.conf, suite.srClientMock, suite.logger)
	return cmd
}

func (suite *ClusterTestSuite) TestCreateSR() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"cluster", "enable", "--cloud", "aws", "--geo", "us"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.srMock.CreateSchemaRegistryClusterCalled())
}

func (suite *ClusterTestSuite) TestDescribeSR() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"cluster", "describe"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.srMock.GetSchemaRegistryClustersCalled())
	req.True(suite.metrics.SchemaRegistryMetricsCalled())
}

func (suite *ClusterTestSuite) TestUpdateCompatibility() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"cluster", "update", "--compatibility", "BACKWARD"}))
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
	cmd.SetArgs(append([]string{"cluster", "update", "--mode", "READWRITE"}))
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
	cmd.SetArgs(append([]string{"cluster", "update"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err, "flag string not set")
}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}
