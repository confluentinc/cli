package apikey

import (
	"context"
	"fmt"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	cmd2 "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

const (
	kafkaClusterID = "kafka"
	srClusterID    = "sr"
)

type APITestSuite struct {
	suite.Suite
	conf         *config.Config
	kafkaCluster *kafkav1.KafkaCluster
	srCluster    *srv1.SchemaRegistryCluster
	apiMock      *mock.APIKey
	ApiKey       *authv1.ApiKey
}

func (suite *APITestSuite) SetupSuite() {
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
		Platform:      name,
		Credential:    name,
		Kafka:         kafkaClusterID,
		KafkaClusters: map[string]*config.KafkaClusterConfig{kafkaClusterID: {}},
	}

	suite.conf.CurrentContext = name

	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         kafkaClusterID,
		Enterprise: true,
	}

	suite.srCluster = &srv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
}

//Require
func (suite *APITestSuite) SetupTest() {

	suite.apiMock = &mock.APIKey{

		CreateFunc: func(ctx context.Context, apiKey *authv1.ApiKey) (*authv1.ApiKey, error) {
			return &authv1.ApiKey{
				Key:    "abrcadabra",
				Secret: "opensesame",
			}, nil
		},
		DeleteFunc: func(ctx context.Context, apiKey *authv1.ApiKey) error {
			return nil
		},
		ListFunc: func(ctx context.Context, apiKey *authv1.ApiKey) ([]*authv1.ApiKey, error) {
			var apiKeys []*authv1.ApiKey
			apiKeys = append(apiKeys, &authv1.ApiKey{
				Key:    "abrcadabra",
				Secret: "opensesame",
			})
			return apiKeys, nil
		},
	}

}

func (suite *APITestSuite) newCMD() *cobra.Command {
	//client ccloud.APIKey, ch *pcmd.ConfigHelper, keystore keystore.KeyStore
	cmd := New(&cliMock.Commander{}, suite.conf, nil, &cmd2.ConfigHelper{}, nil)
	return cmd
}

func (suite *APITestSuite) TestCreateSrApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"api-key", "create", "--resource", srClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
}

func (suite *APITestSuite) TestlistSrApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"api-key", "list", "--resource", srClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
}

func (suite *APITestSuite) TestDeleteApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"api-key", "delete", srClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.DeleteCalled())
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
