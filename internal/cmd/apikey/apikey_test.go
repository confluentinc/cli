package apikey

import (
	"context"
	"fmt"
	"github.com/confluentinc/ccloud-sdk-go"
	"testing"

	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	kafkaClusterID = "kafka"
	srClusterID    = "lsrc-sr"
	apiKey         = "abracadabra"
)

type APITestSuite struct {
	suite.Suite
	conf             *config.Config
	apiMock          *ccsdkmock.APIKey
	keystore         keystore.KeyStore
	kafkaCluster     *kafkav1.KafkaCluster
	srCluster        *srv1.SchemaRegistryCluster
	srMothershipMock *ccsdkmock.SchemaRegistry
	kafkaMock        *ccsdkmock.Kafka
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

	srCluster, _ := suite.conf.SchemaRegistryCluster()
	srCluster.SrCredentials = &config.APIKeyPair{Key: "abracadabra", Secret: "opensesame"}

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
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.srMothershipMock = &ccsdkmock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryClusterConfig) (*srv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryCluster) (*srv1.SchemaRegistryCluster, error) {
			return nil, nil
		},
		GetSchemaRegistryClustersFunc: func(arg0 context.Context, arg1 *srv1.SchemaRegistryCluster) (clusters []*srv1.SchemaRegistryCluster, e error) {
			return []*srv1.SchemaRegistryCluster{suite.srCluster}, nil
		},
	}
	suite.keystore = &mock.KeyStore{
		HasAPIKeyFunc: func(key, clusterID, environment string) (b bool, e error) {
			return true, nil
		},
		StoreAPIKeyFunc: func(key *authv1.ApiKey, clusterID, environment string) error {
			return nil
		},
		DeleteAPIKeyFunc: func(key string) error {
			return nil
		},
	}
	suite.apiMock = &ccsdkmock.APIKey{
		GetFunc: func(ctx context.Context, apiKey *authv1.ApiKey) (key *authv1.ApiKey, e error) {
			return &authv1.ApiKey{
				Key:    "abrcadabra",
				Secret: "opensesame",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, apiKey *authv1.ApiKey) error {
			return nil
		},
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
	cmd := New(&cliMock.Commander{}, suite.conf, suite.apiMock, &pcmd.ConfigHelper{Config: suite.conf, Client: &ccloud.Client{Kafka: suite.kafkaMock, SchemaRegistry: suite.srMothershipMock}}, suite.keystore)
	return cmd
}

func (suite *APITestSuite) TestCreateSrApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"create", "--resource", srClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
}

func (suite *APITestSuite) TestListSrApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"list", "--resource", srClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
}

func (suite *APITestSuite) TestCreateKafkaApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"create", "--resource", kafkaClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
}

func (suite *APITestSuite) TestListKafkaApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"list", "--resource", kafkaClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
}

func (suite *APITestSuite) TestDeleteApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"delete", apiKey}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.DeleteCalled())
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
