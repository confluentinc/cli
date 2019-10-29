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
	cmd2 "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	kafkaClusterID = "kafka"
	connectorID    = "lcc-123"
)

type ConnectTestSuite struct {
	suite.Suite
	conf         *config.Config
	kafkaCluster *kafkav1.KafkaCluster
	logger       *log.Logger
	client       ccloud.Connect
	connector    *v1.Connector
	connectMock  *ccsdkmock.Connect
	kafkaMock    *ccsdkmock.Kafka
}

func (suite *ConnectTestSuite) SetupSuite() {
	suite.conf = config.New()
	suite.conf.Logger = log.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.Auth = &config.AuthConfig{
		User:    new(orgv1.User),
		Account: &orgv1.Account{Id: "testAccount"},
	}
	user := suite.conf.Auth
	name := fmt.Sprintf("login-%s-%s", user.User.Email, suite.conf.AuthURL)

	suite.conf.Platforms[name] = &config.Platform{
		Server: suite.conf.AuthURL,
	}

	suite.conf.Credentials[name] = &config.Credential{
		Username: user.User.Email,
	}

	suite.conf.Contexts[name] = &config.Context{
		Platform:   name,
		Credential: name,
		Kafka:      kafkaClusterID,
	}

	suite.conf.CurrentContext = name

	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         kafkaClusterID,
		Enterprise: true,
	}
	suite.connector = &v1.Connector{Name: "myTestConnector", Id: connectorID, KafkaClusterId: kafkaClusterID, AccountId: "testAccount"}

}

func (suite *ConnectTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		CreateFunc: func(arg0 context.Context, arg1 *v1.ConnectorConfig) (connector *v1.Connector, e error) {
			return suite.connector, nil
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
		ListFunc: func(arg0 context.Context, arg1 *v1.Connector) (connectors []*v1.Connector, e error) {
			return []*v1.Connector{suite.connector}, nil
		},
		UpdateFunc: func(arg0 context.Context, arg1 *v1.Connector) (connector *v1.Connector, e error) {
			return suite.connector, nil
		},
		DescribeFunc: func(arg0 context.Context, arg1 *v1.Connector) (connector *v1.Connector, e error) {
			return suite.connector, nil

		},
	}

}

func (suite *ConnectTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.client, &cmd2.ConfigHelper{Config: suite.conf, Client: &ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}})
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
	req.Equal(retVal.Arg1.Id, connectorID)
}

func (suite *ConnectTestSuite) TestResumeConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"resume", connectorID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ResumeCalled())
	retVal := suite.connectMock.ResumeCalls()[0]
	req.Equal(retVal.Arg1.Id, connectorID)
}

func (suite *ConnectTestSuite) TestDeleteConnector() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"delete", connectorID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	retVal := suite.connectMock.DeleteCalls()[0]
	req.Equal(retVal.Arg1.Id, connectorID)
}

func TestConnectTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectTestSuite))
}
