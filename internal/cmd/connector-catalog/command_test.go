package connector_catalog

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
	pluginType     = "DummyPlugin"
	connectorName  = "myTestConnector"
)

type CatalogTestSuite struct {
	suite.Suite
	conf         *config.Config
	kafkaCluster *kafkav1.KafkaCluster
	connector    *v1.Connector
	connectMock  *ccsdkmock.Connect
	kafkaMock    *ccsdkmock.Kafka
}

func (suite *CatalogTestSuite) SetupSuite() {
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
}

func (suite *CatalogTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		ValidateFunc: func(arg0 context.Context, arg1 *v1.ConnectorConfig, arg2 bool) (connector *v1.ConfigInfos, e error) {
			return nil, errors.New("config.name")
		},
		GetPluginsFunc: func(arg0 context.Context, arg1 *v1.Connector, arg2 string) (infos []*v1.ConnectorPluginInfo, e error) {
			return nil, nil
		},
	}
}

func (suite *CatalogTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.connectMock, &cmd2.ConfigHelper{Config: suite.conf, Client: &ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}})
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
	req.Equal(retVal.Arg1.KafkaClusterId, kafkaClusterID)
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
