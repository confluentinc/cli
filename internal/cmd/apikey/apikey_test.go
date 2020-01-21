package apikey

import (
	"context"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	srClusterID  = "lsrc-12345"
	apiKeyVal    = "abracadabra"
	apiSecretVal = "opensesame"
)

var (
	apiValue = &authv1.ApiKey{
		Key:         apiKeyVal,
		Secret:      apiSecretVal,
		Description: "Mock Apis",
	}
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

//Require
func (suite *APITestSuite) SetupTest() {
	suite.conf = config.AuthenticatedConfigMock()
	ctx := suite.conf.Context()
	srCluster := ctx.SchemaRegistryClusters[ctx.State.Auth.Account.Id]
	srCluster.SrCredentials = &config.APIKeyPair{Key: apiKeyVal, Secret: apiSecretVal}
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
	suite.keystore = &mock.KeyStore{
		HasAPIKeyFunc: func(key, clusterId string, cmd *cobra.Command) (b bool, e error) {
			return true, nil
		},
		StoreAPIKeyFunc: func(key *authv1.ApiKey, clusterId string, cmd *cobra.Command) error {
			return nil
		},
		DeleteAPIKeyFunc: func(key string, cmd *cobra.Command) error {
			return nil
		},
	}
}

func (suite *APITestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		Auth:           &ccsdkmock.Auth{},
		Account:        &ccsdkmock.Account{},
		Kafka:          suite.kafkaMock,
		SchemaRegistry: suite.srMothershipMock,
		Connect:        &ccsdkmock.Connect{},
		User:           &ccsdkmock.User{},
		APIKey:         suite.apiMock,
		KSQL:           &ccsdkmock.MockKSQL{},
		Metrics:        &ccsdkmock.Metrics{},
	}
	cmd := New(cliMock.NewPreRunnerMock(client, nil), suite.conf, suite.keystore)
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
	cmd.SetArgs(append([]string{"create", "--resource", suite.kafkaCluster.Id}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	retValue := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, suite.kafkaCluster.Id)
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
	cmd.SetArgs(append([]string{"list", "--resource", suite.kafkaCluster.Id}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	retValue := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, suite.kafkaCluster.Id)
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
