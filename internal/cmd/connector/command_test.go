package connector

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	v1 "github.com/confluentinc/ccloudapis/connect/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	cliMock "github.com/confluentinc/cli/mock"

	cmd2 "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

const (
	kafkaClusterID = "kafka"
	connectorID    = "lcc-123"
	connectorName  = "myTestConnector"
)

type ConnectTestSuite struct {
	suite.Suite
	conf               *config.Config
	kafkaCluster       *kafkav1.KafkaCluster
	connector          *v1.Connector
	connectorInfo      *v1.ConnectorInfo
	connectMock        *ccsdkmock.Connect
	kafkaMock          *ccsdkmock.Kafka
	connectorExpansion *v1.ConnectorExpansion
}

func (suite *ConnectTestSuite) SetupSuite() {
	suite.conf = config.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.Auth = &config.AuthConfig{
		User:    new(orgv1.User),
		Account: &orgv1.Account{Id: "testAccount"},
	}
	suite.conf.AuthToken = "AuthToken"
	user := suite.conf.Auth
	name := fmt.Sprintf("login-%s-%s", user.User.Email, suite.conf.AuthURL)

	suite.conf.Platforms[name] = &config.Platform{
		Server: suite.conf.AuthURL,
	}

	suite.conf.Credentials[name] = &config.Credential{
		Username: user.User.Email,
	}
	suite.conf.Contexts[name] = &config.Context{
		Platform:      name,
		Credential:    name,
		Kafka:         kafkaClusterID,
		KafkaClusters: map[string]*config.KafkaClusterConfig{kafkaClusterID: {}},
	}

	suite.conf.CurrentContext = name

	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         kafkaClusterID,
		Name:       "KafkaMock",
		AccountId:  "testAccount",
		Enterprise: true,
	}
	suite.connector = &v1.Connector{
		Name:           connectorName,
		Id:             connectorID,
		KafkaClusterId: kafkaClusterID,
		AccountId:      "testAccount",
		Status:         v1.Connector_RUNNING,
		UserConfigs:    map[string]string{},
	}

	suite.connectorInfo = &v1.ConnectorInfo{
		Name: connectorName,
		Type: "source",
	}

	suite.connectorExpansion = &v1.ConnectorExpansion{
		Id: &v1.ConnectorId{Id: connectorID},
		Info: &v1.ConnectorInfo{
			Name:   connectorName,
			Type:   "Sink",
			Config: map[string]string{},
		},
		Status: &v1.ConnectorStateInfo{Name: connectorName, Connector: &v1.ConnectorState{State: "Running"},
			Tasks: []*v1.TaskState{{Id: 1, State: "Running"}},
		}}

}

func (suite *ConnectTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		CreateFunc: func(arg0 context.Context, arg1 *v1.ConnectorConfig) (connector *v1.ConnectorInfo, e error) {
			return suite.connectorInfo, nil
		},
		UpdateFunc: func(arg0 context.Context, arg1 *v1.ConnectorConfig) (info *v1.ConnectorInfo, e error) {
			return suite.connectorInfo, nil
		},
		PauseFunc: func(arg0 context.Context, arg1 *v1.Connector) error {
			return nil
		},
		ResumeFunc: func(arg0 context.Context, arg1 *v1.Connector) error {
			return nil
		},
		DeleteFunc: func(arg0 context.Context, arg1 *v1.Connector) error {
			return nil
		},
		ListWithExpansionsFunc: func(arg0 context.Context, arg1 *v1.Connector, arg2 string) (expansions map[string]*v1.ConnectorExpansion, e error) {
			return map[string]*v1.ConnectorExpansion{connectorID: suite.connectorExpansion}, nil
		},
		GetExpansionByIdFunc: func(arg0 context.Context, arg1 *v1.Connector) (expansion *v1.ConnectorExpansion, e error) {
			return suite.connectorExpansion, nil
		},
		GetExpansionByNameFunc: func(ctx context.Context, connector *v1.Connector) (expansion *v1.ConnectorExpansion, e error) {
			return suite.connectorExpansion, nil
		},
		GetFunc: func(arg0 context.Context, arg1 *v1.Connector) (connector *v1.ConnectorInfo, e error) {
			return suite.connectorInfo, nil
		},
	}

}

func (suite *ConnectTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.connectMock, &cmd2.ConfigHelper{Config: suite.conf, Client: &ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}})
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
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
}

func (suite *ConnectTestSuite) TestResumeConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"resume", connectorID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ResumeCalled())
	retVal := suite.connectMock.ResumeCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
}

func (suite *ConnectTestSuite) TestDeleteConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"delete", connectorID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	retVal := suite.connectMock.DeleteCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
}

func (suite *ConnectTestSuite) TestListConnectors() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"list"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ListWithExpansionsCalled())
	retVal := suite.connectMock.ListWithExpansionsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
}

func (suite *ConnectTestSuite) TestDescribeConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"describe", connectorID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.GetExpansionByIdCalled())
	retVal := suite.connectMock.GetExpansionByIdCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
}

func (suite *ConnectTestSuite) TestCreateConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"create", "--config", "../../../test/fixtures/input/connector-config.yaml"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.CreateCalled())
	retVal := suite.connectMock.CreateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
}

func (suite *ConnectTestSuite) TestUpdateConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"update", connectorID, "--config", "../../../test/fixtures/input/connector-config.yaml"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.UpdateCalled())
	retVal := suite.connectMock.UpdateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
}

func TestConnectTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectTestSuite))
}
