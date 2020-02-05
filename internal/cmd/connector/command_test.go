package connector

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	connectorID   = "lcc-123"
	connectorName = "myTestConnector"
)

type ConnectTestSuite struct {
	suite.Suite
	conf               *v2.Config
	kafkaCluster       *kafkav1.KafkaCluster
	connector          *connectv1.Connector
	connectorInfo      *connectv1.ConnectorInfo
	connectMock        *ccsdkmock.Connect
	kafkaMock          *ccsdkmock.Kafka
	connectorExpansion *connectv1.ConnectorExpansion
}

func (suite *ConnectTestSuite) SetupSuite() {
	suite.conf = v2.AuthenticatedConfigMock()
	ctx := suite.conf.Context()
	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         ctx.KafkaClusters[ctx.Kafka].ID,
		Name:       "KafkaMock",
		AccountId:  "testAccount",
		Enterprise: true,
	}
	suite.connector = &connectv1.Connector{
		Name:           connectorName,
		Id:             connectorID,
		KafkaClusterId: suite.kafkaCluster.Id,
		AccountId:      "testAccount",
		Status:         connectv1.Connector_RUNNING,
		UserConfigs:    map[string]string{},
	}

	suite.connectorInfo = &connectv1.ConnectorInfo{
		Name: connectorName,
		Type: "source",
	}

	suite.connectorExpansion = &connectv1.ConnectorExpansion{
		Id: &connectv1.ConnectorId{Id: connectorID},
		Info: &connectv1.ConnectorInfo{
			Name:   connectorName,
			Type:   "Sink",
			Config: map[string]string{},
		},
		Status: &connectv1.ConnectorStateInfo{Name: connectorName, Connector: &connectv1.ConnectorState{State: "Running"},
			Tasks: []*connectv1.TaskState{{Id: 1, State: "Running"}},
		}}

}

func (suite *ConnectTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		CreateFunc: func(arg0 context.Context, arg1 *connectv1.ConnectorConfig) (connector *connectv1.ConnectorInfo, e error) {
			return suite.connectorInfo, nil
		},
		UpdateFunc: func(arg0 context.Context, arg1 *connectv1.ConnectorConfig) (info *connectv1.ConnectorInfo, e error) {
			return suite.connectorInfo, nil
		},
		PauseFunc: func(arg0 context.Context, arg1 *connectv1.Connector) error {
			return nil
		},
		ResumeFunc: func(arg0 context.Context, arg1 *connectv1.Connector) error {
			return nil
		},
		DeleteFunc: func(arg0 context.Context, arg1 *connectv1.Connector) error {
			return nil
		},
		ListWithExpansionsFunc: func(arg0 context.Context, arg1 *connectv1.Connector, arg2 string) (expansions map[string]*connectv1.ConnectorExpansion, e error) {
			return map[string]*connectv1.ConnectorExpansion{connectorID: suite.connectorExpansion}, nil
		},
		GetExpansionByIdFunc: func(arg0 context.Context, arg1 *connectv1.Connector) (expansion *connectv1.ConnectorExpansion, e error) {
			return suite.connectorExpansion, nil
		},
		GetExpansionByNameFunc: func(ctx context.Context, connector *connectv1.Connector) (expansion *connectv1.ConnectorExpansion, e error) {
			return suite.connectorExpansion, nil
		},
		GetFunc: func(arg0 context.Context, arg1 *connectv1.Connector) (connector *connectv1.ConnectorInfo, e error) {
			return suite.connectorInfo, nil
		},
	}

}

func (suite *ConnectTestSuite) newCMD() *cobra.Command {
	prerunner := cliMock.NewPreRunnerMock(&ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}, nil)
	cmd := New(prerunner, suite.conf)
	return cmd
}

func (suite *ConnectTestSuite) TestPauseConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"pause", connectorID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.PauseCalled())
	retVal := suite.connectMock.PauseCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestResumeConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"resume", connectorID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ResumeCalled())
	retVal := suite.connectMock.ResumeCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestDeleteConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"delete", connectorID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	retVal := suite.connectMock.DeleteCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestListConnectors() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"list"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ListWithExpansionsCalled())
	retVal := suite.connectMock.ListWithExpansionsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestDescribeConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"describe", connectorID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.GetExpansionByIdCalled())
	retVal := suite.connectMock.GetExpansionByIdCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestCreateConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"create", "--config", "../../../test/fixtures/input/connector-config.yaml"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.CreateCalled())
	retVal := suite.connectMock.CreateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestUpdateConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"update", connectorID, "--config", "../../../test/fixtures/input/connector-config.yaml"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.UpdateCalled())
	retVal := suite.connectMock.UpdateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func TestConnectTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectTestSuite))
}
