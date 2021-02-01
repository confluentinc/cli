package apikey

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/c-bata/go-prompt"
	v1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/gogo/protobuf/types"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"

	test_utils "github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	kafkaClusterID    = "lkc-12345"
	srClusterID       = "lsrc-12345"
	apiKeyVal         = "abracadabra"
	apiKeyResourceId  = int32(9999)
	anotherApiKeyVal  = "abba"
	apiSecretVal      = "opensesame"
	promptReadString  = "readstring"
	promptReadPass    = "readpassword"
	environment       = "testAccount"
	apiSecretFile     = "./api_secret_test.txt"
	apiSecretFromFile = "api_secret_test"
	apiKeyDescription = "Mock Apis"
)

var (
	apiValue = &schedv1.ApiKey{
		UserId:      123,
		Key:         apiKeyVal,
		Secret:      apiSecretVal,
		Description: apiKeyDescription,
		Created:     types.TimestampNow(),
		Id:          apiKeyResourceId,
	}
)

type APITestSuite struct {
	suite.Suite
	conf             *v3.Config
	apiMock          *ccsdkmock.APIKey
	keystore         *mock.KeyStore
	kafkaCluster     *schedv1.KafkaCluster
	srCluster        *schedv1.SchemaRegistryCluster
	srMothershipMock *ccsdkmock.SchemaRegistry
	kafkaMock        *ccsdkmock.Kafka
	isPromptPipe     bool
	userMock         *ccsdkmock.User
	analyticsOutput  []segment.Message
	analyticsClient  analytics.Client
}

//Require
func (suite *APITestSuite) SetupTest() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()
	srCluster := ctx.SchemaRegistryClusters[ctx.State.Auth.Account.Id]
	srCluster.SrCredentials = &v0.APIKeyPair{Key: apiKeyVal, Secret: apiSecretVal}
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Endpoint:   cluster.APIEndpoint,
		Enterprise: true,
		AccountId:  environment,
	}
	suite.srCluster = &schedv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.srMothershipMock = &ccsdkmock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *schedv1.SchemaRegistryClusterConfig) (*schedv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClusterFunc: func(ctx context.Context, cluster *schedv1.SchemaRegistryCluster) (*schedv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClustersFunc: func(ctx context.Context, cluster *schedv1.SchemaRegistryCluster) (clusters []*schedv1.SchemaRegistryCluster, e error) {
			return []*schedv1.SchemaRegistryCluster{suite.srCluster}, nil
		},
	}
	suite.apiMock = &ccsdkmock.APIKey{
		GetFunc: func(ctx context.Context, apiKey *schedv1.ApiKey) (key *schedv1.ApiKey, e error) {
			return apiValue, nil
		},
		UpdateFunc: func(ctx context.Context, apiKey *schedv1.ApiKey) error {
			return nil
		},
		CreateFunc: func(ctx context.Context, apiKey *schedv1.ApiKey) (*schedv1.ApiKey, error) {
			return apiValue, nil
		},
		DeleteFunc: func(ctx context.Context, apiKey *schedv1.ApiKey) error {
			return nil
		},
		ListFunc: func(ctx context.Context, apiKey *schedv1.ApiKey) ([]*schedv1.ApiKey, error) {
			return []*schedv1.ApiKey{apiValue}, nil
		},
	}
	suite.keystore = &mock.KeyStore{
		HasAPIKeyFunc: func(key, clusterId string, cmd *cobra.Command) (b bool, e error) {
			return key == apiKeyVal, nil
		},
		StoreAPIKeyFunc: func(key *schedv1.ApiKey, clusterId string, cmd *cobra.Command) error {
			return nil
		},
		DeleteAPIKeyFunc: func(key string, cmd *cobra.Command) error {
			return nil
		},
	}
	suite.userMock = &ccsdkmock.User{
		DescribeFunc: func(arg0 context.Context, arg1 *v1.User) (user *v1.User, e error) {
			return &v1.User{
				Email: "csreesangkom@confluent.io",
			}, nil
		},
		GetServiceAccountsFunc: func(arg0 context.Context) (users []*v1.User, e error) {
			return []*v1.User{}, nil
		},
		CheckEmailFunc: nil,
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = test_utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *APITestSuite) newCmd() *command {
	client := &ccloud.Client{
		Auth:           &ccsdkmock.Auth{},
		Account:        &ccsdkmock.Account{},
		Kafka:          suite.kafkaMock,
		SchemaRegistry: suite.srMothershipMock,
		Connect:        &ccsdkmock.Connect{},
		User:           suite.userMock,
		APIKey:         suite.apiMock,
		KSQL:           &ccsdkmock.KSQL{},
		Metrics:        &ccsdkmock.Metrics{},
	}
	prompt := &mock.Prompt{
		ReadLineFunc: func() (string, error) {
			return promptReadString + "\n", nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			return promptReadPass + "\n", nil
		},
		IsPipeFunc: func() (b bool, e error) {
			return suite.isPromptPipe, nil
		},
	}
	resolverMock := &pcmd.FlagResolverImpl{
		Prompt: prompt,
		Out:    os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver: resolverMock,
		Client:       client,
		MDSClient:    nil,
		Config:       suite.conf,
	}
	return New(prerunner, suite.keystore, resolverMock, suite.analyticsClient)
}

func (suite *APITestSuite) TestCreateSrApiKey() {
	cmd := suite.newCmd()
	args := append([]string{"create", "--resource", srClusterID})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	inputKey := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, srClusterID)
	checkTrackedResourceAndKey(suite.analyticsOutput[0], req)
}

func checkTrackedResourceAndKey(segmentMsg segment.Message, req *require.Assertions) {
	test_utils.CheckTrackedResourceIDInt32(segmentMsg, apiKeyResourceId, req)

	key, err := test_utils.GetPagePropertyValue(segmentMsg, analytics.ApiKeyPropertiesKey)
	req.NoError(err)
	req.Equal(apiKeyVal, key.(string))
}

func (suite *APITestSuite) TestCreateKafkaApiKey() {
	cmd := suite.newCmd()
	args := append([]string{"create", "--resource", suite.kafkaCluster.Id})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	inputKey := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, suite.kafkaCluster.Id)
	checkTrackedResourceAndKey(suite.analyticsOutput[0], req)
}

func (suite *APITestSuite) TestCreateCloudAPIKey() {
	cmd := suite.newCmd()
	args := append([]string{"create", "--resource", "cloud"})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	inputKey := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(0, len(inputKey.LogicalClusters))
	checkTrackedResourceAndKey(suite.analyticsOutput[0], req)
}

func (suite *APITestSuite) TestDeleteApiKey() {
	cmd := suite.newCmd()
	args := append([]string{"delete", apiKeyVal})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.DeleteCalled())
	inputKey := suite.apiMock.DeleteCalls()[0].Arg1
	req.Equal(inputKey.Key, apiKeyVal)
	checkTrackedResourceAndKey(suite.analyticsOutput[0], req)
}

func (suite *APITestSuite) TestListSrApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"list", "--resource", srClusterID}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	inputKey := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, srClusterID)
}

func (suite *APITestSuite) TestListKafkaApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"list", "--resource", suite.kafkaCluster.Id}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	inputKey := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, suite.kafkaCluster.Id)
}

func (suite *APITestSuite) TestListCloudAPIKey() {
	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"list", "--resource", "cloud"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	inputKey := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(0, len(inputKey.LogicalClusters))
}

func (suite *APITestSuite) TestStoreApiKeyForce() {
	req := require.New(suite.T())
	suite.isPromptPipe = false
	cmd := suite.newCmd()
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
	cmd := suite.newCmd()
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
	cmd := suite.newCmd()
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
	cmd := suite.newCmd()
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
	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"store", "--resource", kafkaClusterID}))
	err := cmd.Execute()
	req.NoError(err)
	req.True(suite.keystore.StoreAPIKeyCalled())
	args := suite.keystore.StoreAPIKeyCalls()[0]
	req.Equal(promptReadString, args.Key.Key)
	req.Equal(promptReadPass, args.Key.Secret)
}

func (suite *APITestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	suite.isPromptPipe = false
	cmd := suite.newCmd()
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"api-key update", "api-key delete", "api-key store", "api-key use"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}

func (suite *APITestSuite) TestServerComplete() {
	req := suite.Require()
	type fields struct {
		Command *command
	}
	tests := []struct {
		name   string
		fields fields
		want   []prompt.Suggest
	}{
		{
			name: "suggest for authenticated user",
			fields: fields{
				Command: suite.newCmd(),
			},
			want: []prompt.Suggest{
				{
					Text:        apiKeyVal,
					Description: apiKeyDescription,
				},
			},
		},
		{
			name: "don't suggest for unauthenticated user",
			fields: fields{
				Command: func() *command {
					oldConf := suite.conf
					suite.conf = v3.UnauthenticatedCloudConfigMock()
					c := suite.newCmd()
					suite.conf = oldConf
					return c
				}(),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := tt.fields.Command.ServerComplete()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}
func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
