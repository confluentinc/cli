package apikey

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	kafkaClusterID     = "lkc-12345"
	srClusterID        = "lsrc-12345"
	apiKeyVal          = "abracadabra"
	apiKeyResourceId   = int32(9999)
	anotherApiKeyVal   = "abba"
	apiSecretVal       = "opensesame"
	promptReadString   = "readstring"
	promptReadPass     = "readpassword"
	environment        = "testAccount"
	apiSecretFile      = "./api_secret_test.txt"
	apiSecretFromFile  = "api_secret_test"
	apiKeyDescription  = "Mock Apis"
	serviceAccountId   = int32(123)
	userResourceId     = "sa-55555"
	serviceAccountName = "service-account"

	auditLogApiKeyResourceId  = int32(7753)
	auditLogApiKeyVal         = "auditlog-apikey"
	auditLogApiKeySecretVal   = "opensesameforauditlogs"
	auditLogApiKeyDescription = "Mock Apis for Audit Logs"
	auditLogServiceAccountId  = int32(748)
	auditLogUserResourceId    = "sa-55555"

	myApiKeyResourceId  = int32(1234)
	myApiKeyVal         = "user-apikey"
	myApiKeySecretVal   = "icreatedthis"
	myApiKeyDescription = "Mock Apis for User"
	myServiceAccountId  = int32(987)
	myUserResourceId    = "u-123"
	myAccountName       = "My Account"
)

var (
	// apiValue = &schedv1.ApiKey{
	// 	LogicalClusters: []*schedv1.ApiKey_Cluster{{Id: kafkaClusterID, Type: "kafka"}},
	// 	UserId:          serviceAccountId,
	// 	UserResourceId:  userResourceId,
	// 	Key:             apiKeyVal,
	// 	Secret:          apiSecretVal,
	// 	Description:     apiKeyDescription,
	// 	Created:         types.TimestampNow(),
	// 	Id:              apiKeyResourceId,
	// }
	apiValue = &apikeysv2.IamV2ApiKey{
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Description: apikeysv2.PtrString(apiKeyDescription),
			Resource:    &apikeysv2.ObjectReference{},
			Owner: &apikeysv2.ObjectReference{
				Id: *apikeysv2.PtrString(apiValue.UserResourceId),
			},
			Secret: apikeysv2.PtrString(apiSecretVal),
		},
		Id: apikeysv2.PtrString(apiKeyVal),
		Metadata: &apikeysv2.ObjectMeta{
			CreatedAt: &time.Time{},
		},
	}
	auditLogApiValue = &schedv1.ApiKey{
		LogicalClusters: []*schedv1.ApiKey_Cluster{{Id: kafkaClusterID, Type: "kafka"}},
		UserId:          auditLogServiceAccountId,
		UserResourceId:  auditLogUserResourceId,
		Key:             auditLogApiKeyVal,
		Secret:          auditLogApiKeySecretVal,
		Description:     auditLogApiKeyDescription,
		Created:         types.TimestampNow(),
		Id:              auditLogApiKeyResourceId,
	}
	myApiValue = &schedv1.ApiKey{
		LogicalClusters: []*schedv1.ApiKey_Cluster{{Id: kafkaClusterID, Type: "kafka"}},
		UserId:          myServiceAccountId,
		UserResourceId:  myUserResourceId,
		Key:             myApiKeyVal,
		Secret:          myApiKeySecretVal,
		Description:     myApiKeyDescription,
		Created:         types.TimestampNow(),
		Id:              myApiKeyResourceId,
	}
)

type APITestSuite struct {
	suite.Suite
	conf             *v1.Config
	apiMock          *ccsdkmock.APIKey
	keystore         *mock.KeyStore
	kafkaCluster     *schedv1.KafkaCluster
	ksqlCluster      *schedv1.KSQLCluster
	srCluster        *schedv1.SchemaRegistryCluster
	srMothershipMock *ccsdkmock.SchemaRegistry
	kafkaMock        *ccsdkmock.Kafka
	ksqlMock         *ccsdkmock.KSQL
	isPromptPipe     bool
	userMock         *ccsdkmock.User
}

//Require
func (suite *APITestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()

	srCluster := ctx.SchemaRegistryClusters[ctx.State.Auth.Account.Id]
	srCluster.SrCredentials = &v1.APIKeyPair{Key: apiKeyVal, Secret: apiSecretVal}
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	// Set up audit logs
	ctx.State.Auth.Organization.AuditLog = &orgv1.AuditLog{
		ClusterId:        cluster.ID,
		AccountId:        "env-zy987",
		ServiceAccountId: auditLogServiceAccountId,
		TopicName:        "confluent-audit-log-events",
	}
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Endpoint:   cluster.APIEndpoint,
		Enterprise: true,
		AccountId:  environment,
	}
	suite.ksqlCluster = &schedv1.KSQLCluster{
		Id:   "ksql-123",
		Name: "ksql",
	}
	suite.srCluster = &schedv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
		ListFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (clusters []*schedv1.KafkaCluster, e error) {
			return []*schedv1.KafkaCluster{suite.kafkaCluster}, nil
		},
	}
	suite.ksqlMock = &ccsdkmock.KSQL{
		ListFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) (clusters []*schedv1.KSQLCluster, e error) {
			return []*schedv1.KSQLCluster{suite.ksqlCluster}, nil
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
			switch apiKey.Key {
			case auditLogApiValue.Key:
				return auditLogApiValue, nil
			default:
				return apiValue, nil
			}
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
			return []*schedv1.ApiKey{apiValue, auditLogApiValue, myApiValue}, nil
		},
	}
	suite.keystore = &mock.KeyStore{
		HasAPIKeyFunc: func(key, clusterId string) (b bool, e error) {
			return key == apiKeyVal, nil
		},
		StoreAPIKeyFunc: func(key *schedv1.ApiKey, clusterId string) error {
			return nil
		},
		DeleteAPIKeyFunc: func(key string) error {
			return nil
		},
	}
	suite.userMock = &ccsdkmock.User{
		DescribeFunc: func(arg0 context.Context, arg1 *orgv1.User) (user *orgv1.User, e error) {
			return &orgv1.User{
				Email: "csreesangkom@confluent.io",
			}, nil
		},
		GetServiceAccountsFunc: func(arg0 context.Context) (users []*orgv1.User, e error) {
			return []*orgv1.User{
				{
					Id:          serviceAccountId,
					ResourceId:  userResourceId,
					ServiceName: serviceAccountName,
				},
			}, nil
		},
		GetServiceAccountFunc: func(_ context.Context, _ int32) (*orgv1.User, error) {
			return &orgv1.User{
				Id:          serviceAccountId,
				ResourceId:  userResourceId,
				ServiceName: serviceAccountName,
			}, nil
		},
		ListFunc: func(_ context.Context) ([]*orgv1.User, error) {
			return []*orgv1.User{
				{
					Id:          serviceAccountId,
					ResourceId:  userResourceId,
					ServiceName: serviceAccountName,
				},
				{
					Id:          myServiceAccountId,
					ResourceId:  myUserResourceId,
					ServiceName: myAccountName,
					Email:       "csreesangkom@confluent.io",
				},
			}, nil
		},
	}
}

func (suite *APITestSuite) newCmd() *cobra.Command {
	client := &ccloud.Client{
		Auth:           &ccsdkmock.Auth{},
		Account:        &ccsdkmock.Account{},
		Kafka:          suite.kafkaMock,
		SchemaRegistry: suite.srMothershipMock,
		Connect:        &ccsdkmock.Connect{},
		User:           suite.userMock,
		APIKey:         suite.apiMock,
		KSQL:           suite.ksqlMock,
		Metrics:        &ccsdkmock.Metrics{},
	}
	resolverMock := &pcmd.FlagResolverImpl{
		Prompt: &mock.Prompt{
			ReadLineFunc: func() (string, error) {
				return promptReadString + "\n", nil
			},
			ReadLineMaskedFunc: func() (string, error) {
				return promptReadPass + "\n", nil
			},
			IsPipeFunc: func() (bool, error) {
				return suite.isPromptPipe, nil
			},
		},
		Out: os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver: resolverMock,
		Client:       client,
		MDSClient:    nil,
		Config:       suite.conf,
	}
	return New(prerunner, suite.keystore, resolverMock)
}

func (suite *APITestSuite) TestCreateSrApiKey() {
	cmd := suite.newCmd()
	args := []string{"create", "--resource", srClusterID}
	cmd.SetArgs(args)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	inputKey := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, srClusterID)
}

func (suite *APITestSuite) TestCreateKafkaApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--resource", suite.kafkaCluster.Id})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	inputKey := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, suite.kafkaCluster.Id)
}

func (suite *APITestSuite) TestCreateCloudAPIKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--resource", "cloud"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.CreateCalled())
	inputKey := suite.apiMock.CreateCalls()[0].Arg1
	req.Equal(0, len(inputKey.LogicalClusters))
}

func (suite *APITestSuite) TestDeleteApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"delete", apiKeyVal})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.DeleteCalled())
	inputKey := suite.apiMock.DeleteCalls()[0].Arg1
	req.Equal(inputKey.Key, apiKeyVal)
}

func (suite *APITestSuite) TestListSrApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"list", "--resource", srClusterID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	inputKey := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, srClusterID)
}

func (suite *APITestSuite) TestListKafkaApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"list", "--resource", suite.kafkaCluster.Id})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	inputKey := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, suite.kafkaCluster.Id)
}

// Audit Log Destination Clusters are kafka clusters, however their API keys are created by internal service accounts
func (suite *APITestSuite) TestListAuditLogDestinationClusterApiKey() {
	cmd := suite.newCmd()
	buf := new(bytes.Buffer)
	cmd.SetArgs([]string{"list", "--resource", suite.kafkaCluster.Id})
	cmd.SetOut(buf)

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	inputKey := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(inputKey.LogicalClusters[0].Id, suite.kafkaCluster.Id)
	req.Equal(inputKey.LogicalClusters[0].Id, suite.kafkaCluster.Id)
	req.Contains(buf.String(), "auditlog service account")
}

func (suite *APITestSuite) TestListCloudAPIKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"list", "--resource", "cloud"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	inputKey := suite.apiMock.ListCalls()[0].Arg1
	req.Equal(0, len(inputKey.LogicalClusters))
}

func (suite *APITestSuite) TestListEmails() {
	cmd := suite.newCmd()
	buf := new(bytes.Buffer)
	cmd.SetArgs([]string{"list"})
	cmd.SetOut(buf)

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiMock.ListCalled())
	req.Contains(buf.String(), "<auditlog service account>")
	req.Contains(buf.String(), "<service account>")
	req.Contains(buf.String(), "csreesangkom@confluent.io")
}

func (suite *APITestSuite) TestStoreApiKeyForce() {
	req := require.New(suite.T())
	suite.isPromptPipe = false
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"store", apiKeyVal, apiSecretVal, "--resource", kafkaClusterID})
	err := cmd.Execute()
	// refusing to overwrite existing secret
	req.Error(err)
	req.False(suite.keystore.StoreAPIKeyCalled())

	cmd.SetArgs([]string{"store", apiKeyVal, apiSecretVal, "-f", "--resource", kafkaClusterID})
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
	cmd.SetArgs([]string{"store", anotherApiKeyVal, "-", "--resource", kafkaClusterID})
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
	cmd.SetArgs([]string{"store", anotherApiKeyVal, "--resource", kafkaClusterID})
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
	cmd.SetArgs([]string{"store", anotherApiKeyVal, "@" + apiSecretFile, "--resource", kafkaClusterID})
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
	cmd.SetArgs([]string{"store", "--resource", kafkaClusterID})
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
