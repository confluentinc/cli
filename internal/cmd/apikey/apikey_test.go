package apikey

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	srClusterID    = "lsrc-12345"
	apiKeyVal      = "abracadabra"
	apiSecretVal   = "opensesame"
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
	keystore         keystore.KeyStore
	kafkaCluster     *kafkav1.KafkaCluster
	srCluster        *srv1.SchemaRegistryCluster
	srMothershipMock *ccsdkmock.SchemaRegistry
	kafkaMock        *ccsdkmock.Kafka
}

func (suite *APITestSuite) SetupSuite() {
	suite.conf = config.New()
	suite.conf.Logger = log.New()
	suite.conf = config.AuthenticatedConfigMock()
	srCluster, _ := suite.conf.SchemaRegistryCluster()
	srCluster.SrCredentials = &config.APIKeyPair{Key: apiKeyVal, Secret: apiSecretVal}
	cluster := suite.conf.Context().ActiveKafkaCluster()
	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Endpoint:   cluster.APIEndpoint,
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
		GetSchemaRegistryClusterFunc: func(ctx context.Context, cluster *srv1.SchemaRegistryCluster) (*srv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClustersFunc: func(ctx context.Context, cluster *srv1.SchemaRegistryCluster) (clusters []*srv1.SchemaRegistryCluster, e error) {
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
	cmd := New(&cliMock.Commander{}, suite.conf, suite.apiMock, &pcmd.ConfigHelper{Config: suite.conf, Client: &ccloud.Client{Kafka: suite.kafkaMock, SchemaRegistry: suite.srMothershipMock, APIKey: suite.apiMock}}, suite.keystore)
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
	cluster := suite.conf.Context().ActiveKafkaCluster()
	cmd.SetArgs(append([]string{"create", "--resource", cluster.ID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	retValue := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, cluster.ID)
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
	cluster := suite.conf.Context().ActiveKafkaCluster()
	cmd.SetArgs(append([]string{"list", "--resource", cluster.ID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	retValue := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(retValue.LogicalClusters[0].Id, cluster.ID)
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
