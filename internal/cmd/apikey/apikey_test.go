package apikey

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	kafkaClusterID    = "lkc-12345"
	srClusterID       = "lsrc-12345"
	apiKeyVal         = "abracadabra"
	anotherApiKeyVal  = "abba"
	apiSecretVal      = "opensesame"
	promptReadString  = "readstring"
	promptReadPass    = "readpassword"
	environment       = "testAccount"
	apiSecretFile     = "./api_secret_test.txt"
	apiSecretFromFile = "api_secret_test"
)

var (
	apiValue = &authv1.ApiKey{
		Key:         apiKeyVal,
		Secret:      apiSecretVal,
		Description: "Mock Api's",
	}
)

type APITestSuite struct {
	suite.Suite
	conf             *config.Config
	apiMock          *ccsdkmock.APIKey
	keystore         *mock.KeyStore
	kafkaCluster     *kafkav1.KafkaCluster
	srCluster        *srv1.SchemaRegistryCluster
	srMothershipMock *ccsdkmock.SchemaRegistry
	kafkaMock        *ccsdkmock.Kafka
	isPromptPipe     bool
}

func (suite *APITestSuite) SetupSuite() {
	suite.conf = config.New()
	suite.conf.Logger = log.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.Auth = &config.AuthConfig{
		User:    new(orgv1.User),
		Account: &orgv1.Account{Id: environment},
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

	srCluster, _ := suite.conf.SchemaRegistryCluster()
	srCluster.SrCredentials = &config.APIKeyPair{Key: apiKeyVal, Secret: apiSecretVal}

	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         kafkaClusterID,
		Enterprise: true,
		AccountId:  environment,
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
		GetSchemaRegistryClusterFunc: func(ctx context.Context, cluster *srv1.SchemaRegistryCluster) (*srv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClustersFunc: func(ctx context.Context, cluster *srv1.SchemaRegistryCluster) (clusters []*srv1.SchemaRegistryCluster, e error) {
			return []*srv1.SchemaRegistryCluster{suite.srCluster}, nil
		},
	}
	suite.keystore = &mock.KeyStore{
		HasAPIKeyFunc: func(key, clusterID, environment string) (b bool, e error) {
			return key == apiKeyVal, nil
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
			return apiValue, nil
		},
		UpdateFunc: func(ctx context.Context, apiKey *authv1.ApiKey) error {
			return nil
		},
		CreateFunc: func(ctx context.Context, apiKey *authv1.ApiKey) (*authv1.ApiKey, error) {
			return apiValue, nil
		},
		DeleteFunc: func(ctx context.Context, apiKey *authv1.ApiKey) error {
			return nil
		},
		ListFunc: func(ctx context.Context, apiKey *authv1.ApiKey) ([]*authv1.ApiKey, error) {
			return []*authv1.ApiKey{apiValue}, nil
		},
	}
}

func (suite *APITestSuite) newCMD() *cobra.Command {
	prompt := &cliMock.Prompt{
		ReadStringFunc: func(delim byte) (s string, e error) {
			return promptReadString + "\n", nil
		},
		ReadPasswordFunc: func() (bytes []byte, e error) {
			return []byte(promptReadPass + "\n"), nil
		},
		IsPipeFunc: func() (b bool, e error) {
			return suite.isPromptPipe, nil
		},
	}
	resolver := &pcmd.FlagResolverImpl{Prompt: prompt, Out: os.Stdout}
	cmd := New(&cliMock.Commander{}, suite.conf, suite.apiMock, &pcmd.ConfigHelper{Config: suite.conf, Client: &ccloud.Client{Kafka: suite.kafkaMock, SchemaRegistry: suite.srMothershipMock, APIKey: suite.apiMock}}, suite.keystore, resolver)
	return cmd
}

func (suite *APITestSuite) TestCreateSrApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"create", "--resource", srClusterID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	retValue := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, srClusterID)
}

func (suite *APITestSuite) TestCreateKafkaApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"create", "--resource", kafkaClusterID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	retValue := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, kafkaClusterID)
}

func (suite *APITestSuite) TestDeleteApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"delete", apiKeyVal}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.DeleteCalled())
	retValue := suite.apiMock.DeleteCalls()[0].Arg1
	req.Equal(retValue.Key, apiKeyVal)
}

func (suite *APITestSuite) TestListSrApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"list", "--resource", srClusterID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	retValue := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, srClusterID)
}

func (suite *APITestSuite) TestListKafkaApiKey() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"list", "--resource", kafkaClusterID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	retValue := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, kafkaClusterID)
}

func (suite *APITestSuite) TestStoreApiKeyForce() {
	req := require.New(suite.T())
	suite.isPromptPipe = false
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"store", apiKeyVal, apiSecretVal, "--resource", kafkaClusterID}))
	err := cmd.Execute()
	// refusing to overwrite existing secret
	req.Error(err)
	req.False(suite.keystore.StoreAPIKeyCalled())

	cmd.SetArgs(append([]string{"store", apiKeyVal, apiSecretVal, "-f", "--resource", kafkaClusterID}))
	err = cmd.Execute()
	req.NoError(err)
	req.True(suite.keystore.StoreAPIKeyCalled())
	args := suite.keystore.StoreAPIKeyCalls()[0]
	req.Equal(apiKeyVal, args.Key.Key)
	req.Equal(apiSecretVal, args.Key.Secret)
}

func (suite *APITestSuite) TestStoreApiKeyPipe() {
	req := require.New(suite.T())
	suite.isPromptPipe = true
	cmd := suite.newCMD()
	// no need to force for new api keys
	cmd.SetArgs(append([]string{"store", anotherApiKeyVal, "-", "--resource", kafkaClusterID}))
	err := cmd.Execute()
	req.NoError(err)
	req.True(suite.keystore.StoreAPIKeyCalled())
	args := suite.keystore.StoreAPIKeyCalls()[0]
	req.Equal(anotherApiKeyVal, args.Key.Key)
	req.Equal(promptReadString, args.Key.Secret)
}

func (suite *APITestSuite) TestStoreApiKeyPromptUserForSecret() {
	req := require.New(suite.T())
	suite.isPromptPipe = false
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"store", anotherApiKeyVal, "--resource", kafkaClusterID}))
	err := cmd.Execute()
	req.NoError(err)
	req.True(suite.keystore.StoreAPIKeyCalled())
	args := suite.keystore.StoreAPIKeyCalls()[0]
	req.Equal(anotherApiKeyVal, args.Key.Key)
	req.Equal(promptReadPass, args.Key.Secret)
}

func (suite *APITestSuite) TestStoreApiKeyPassSecretByFile() {
	req := require.New(suite.T())
	suite.isPromptPipe = false
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"store", anotherApiKeyVal, "@" + apiSecretFile, "--resource", kafkaClusterID}))
	err := cmd.Execute()
	req.NoError(err)
	req.True(suite.keystore.StoreAPIKeyCalled())
	args := suite.keystore.StoreAPIKeyCalls()[0]
	req.Equal(anotherApiKeyVal, args.Key.Key)
	req.Equal(apiSecretFromFile, args.Key.Secret)
}

func (suite *APITestSuite) TestStoreApiKeyPromptUserForKeyAndSecret() {
	req := require.New(suite.T())
	suite.isPromptPipe = false
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"store", "--resource", kafkaClusterID}))
	err := cmd.Execute()
	req.NoError(err)
	req.True(suite.keystore.StoreAPIKeyCalled())
	args := suite.keystore.StoreAPIKeyCalls()[0]
	req.Equal(promptReadString, args.Key.Key)
	req.Equal(promptReadPass, args.Key.Secret)
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
