package connect

import (
	"context"
	"fmt"
	"testing"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	connectorID   = "lcc-123"
	connectorName = "myTestConnector"
	pluginType    = "DummyPlugin"
)

type ConnectTestSuite struct {
	suite.Suite
	conf               *v1.Config
	kafkaCluster       *schedv1.KafkaCluster
	connector          *schedv1.Connector
	connectorInfo      *opv1.ConnectorInfo
	connectMock        *ccsdkmock.Connect
	kafkaMock          *ccsdkmock.Kafka
	connectorExpansion *opv1.ConnectorExpansion
}

func (suite *ConnectTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
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

	suite.connectorInfo = &opv1.ConnectorInfo{
		Name: connectorName,
		Type: "source",
	}

	suite.connectorExpansion = &opv1.ConnectorExpansion{
		Id: &opv1.ConnectorId{Id: connectorID},
		Info: &opv1.ConnectorInfo{
			Name:   connectorName,
			Type:   "Sink",
			Config: map[string]string{},
		},
		Status: &opv1.ConnectorStateInfo{Name: connectorName, Connector: &opv1.ConnectorState{State: "Running"},
			Tasks: []*opv1.TaskState{{Id: 1, State: "Running"}},
		}}

}

func (suite *ConnectTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(_ context.Context, _ *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
		ListFunc: func(_ context.Context, _ *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
			return []*schedv1.KafkaCluster{suite.kafkaCluster}, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		CreateFunc: func(_ context.Context, _ *schedv1.ConnectorConfig) (*opv1.ConnectorInfo, error) {
			return suite.connectorInfo, nil
		},
		UpdateFunc: func(_ context.Context, _ *schedv1.ConnectorConfig) (*opv1.ConnectorInfo, error) {
			return suite.connectorInfo, nil
		},
		PauseFunc: func(_ context.Context, _ *schedv1.Connector) error {
			return nil
		},
		ResumeFunc: func(_ context.Context, _ *schedv1.Connector) error {
			return nil
		},
		DeleteFunc: func(_ context.Context, _ *schedv1.Connector) error {
			return nil
		},
		ListWithExpansionsFunc: func(_ context.Context, _ *schedv1.Connector, _ string) (map[string]*opv1.ConnectorExpansion, error) {
			return map[string]*opv1.ConnectorExpansion{connectorID: suite.connectorExpansion}, nil
		},
		GetExpansionByIdFunc: func(_ context.Context, _ *schedv1.Connector) (*opv1.ConnectorExpansion, error) {
			return suite.connectorExpansion, nil
		},
		GetExpansionByNameFunc: func(_ context.Context, _ *schedv1.Connector) (*opv1.ConnectorExpansion, error) {
			return suite.connectorExpansion, nil
		},
		GetFunc: func(_ context.Context, _ *schedv1.Connector) (*opv1.ConnectorInfo, error) {
			return suite.connectorInfo, nil
		},
		ValidateFunc: func(_ context.Context, _ *schedv1.ConnectorConfig) (*opv1.ConfigInfos, error) {
			return &opv1.ConfigInfos{Configs: []*opv1.Configs{{Value: &opv1.ConfigValue{Value: "abc", Errors: []string{"new error"}}}}}, errors.New("config.name")
		},
		GetPluginsFunc: func(_ context.Context, _ *schedv1.Connector, _ string) ([]*opv1.ConnectorPluginInfo, error) {
			return []*opv1.ConnectorPluginInfo{
				{
					Class: "test-plugin",
					Type:  "source",
				},
			}, nil
		},
	}
}

func (suite *ConnectTestSuite) newCmd() *cobra.Command {
	prerunner := cliMock.NewPreRunnerMock(&ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}, nil, nil, suite.conf)
	return New(prerunner)
}

func (suite *ConnectTestSuite) TestPauseConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"pause", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.PauseCalled())
	retVal := suite.connectMock.PauseCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestResumeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"resume", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ResumeCalled())
	retVal := suite.connectMock.ResumeCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestDeleteConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"delete", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	retVal := suite.connectMock.DeleteCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestListConnectors() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ListWithExpansionsCalled())
	retVal := suite.connectMock.ListWithExpansionsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestDescribeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"describe", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.GetExpansionByIdCalled())
	retVal := suite.connectMock.GetExpansionByIdCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestCreateConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--config", "../../../test/fixtures/input/connect/config.yaml"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.CreateCalled())
	retVal := suite.connectMock.CreateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestCreateConnectorNewFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--config", "../../../test/fixtures/input/connect/config-new-format.json"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.CreateCalled())
	retVal := suite.connectMock.CreateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestCreateConnectorMalformedNewFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--config", "../../../test/fixtures/input/connect/config-malformed-new.json"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
	fmt.Printf("error-- %s", err.Error())
	assert.Contains(suite.T(), err.Error(), "unable to parse config")
}

func (suite *ConnectTestSuite) TestCreateConnectorMalformedOldFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--config", "../../../test/fixtures/input/connect/config-malformed-old.json"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
	fmt.Printf("error-- %s", err.Error())
	assert.Contains(suite.T(), err.Error(), "unable to parse config")
}

func (suite *ConnectTestSuite) TestUpdateConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"update", connectorID, "--config", "../../../test/fixtures/input/connect/config.yaml"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.UpdateCalled())
	retVal := suite.connectMock.UpdateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestPluginList() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"plugin", "list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.connectMock.GetPluginsCalled())
	retVal := suite.connectMock.GetPluginsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestPluginDescribeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"plugin", "describe", pluginType})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.connectMock.ValidateCalled())
	retVal := suite.connectMock.ValidateCalls()[0]
	req.Equal(retVal.Arg1.Plugin, pluginType)
}

func TestConnectTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectTestSuite))
}
