package connector_catalog

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	connectorID   = "lcc-123"
	pluginType    = "DummyPlugin"
	connectorName = "myTestConnector"
)

type CatalogTestSuite struct {
	suite.Suite
	conf         *v3.Config
	kafkaCluster *schedv1.KafkaCluster
	connector    *schedv1.Connector
	connectMock  *ccsdkmock.Connect
	kafkaMock    *ccsdkmock.Kafka
}

func (suite *CatalogTestSuite) SetupSuite() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:         ctx.KafkaClusterContext.GetActiveKafkaClusterId(),
		Name:       "KafkaMock",
		AccountId:  "testAccount",
		Enterprise: true,
	}
	suite.connector = &schedv1.Connector{
		Name:           connectorName,
		Id:             connectorID,
		KafkaClusterId: suite.kafkaCluster.Id,
		AccountId:      "testAccount",
		Status:         schedv1.Connector_RUNNING,
		UserConfigs:    map[string]string{},
	}
}

func (suite *CatalogTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		ValidateFunc: func(arg0 context.Context, arg1 *schedv1.ConnectorConfig) (connector *opv1.ConfigInfos, e error) {
			return &opv1.ConfigInfos{Configs: []*opv1.Configs{{Value: &opv1.ConfigValue{Value: "abc", Errors: []string{"new error"}}}}}, errors.New("config.name")
		},
		GetPluginsFunc: func(arg0 context.Context, arg1 *schedv1.Connector, arg2 string) (infos []*opv1.ConnectorPluginInfo, e error) {
			return nil, nil
		},
	}
}

func (suite *CatalogTestSuite) newCMD() *cobra.Command {
	prerunner := cliMock.NewPreRunnerMock(&ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}, nil)
	cmd := New(prerunner, suite.conf)
	return cmd
}

func (suite *CatalogTestSuite) TestCatalogList() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"list"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.GetPluginsCalled())
	retVal := suite.connectMock.GetPluginsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *CatalogTestSuite) TestCatalogDescribeConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"describe", pluginType}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ValidateCalled())
	retVal := suite.connectMock.ValidateCalls()[0]
	req.Equal(retVal.Arg1.Plugin, pluginType)
}

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, new(CatalogTestSuite))
}
