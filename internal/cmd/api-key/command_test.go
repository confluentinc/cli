package apikey

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	apikeysmock "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2/mock"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	iammock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"
	ksqlmock "github.com/confluentinc/ccloud-sdk-go-v2/ksql/mock"
	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/mock"
	climock "github.com/confluentinc/cli/mock"
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
	userResourceId     = "sa-12345"
	serviceAccountName = "service-account"

	auditLogApiKeyVal         = "auditlog-apikey"
	auditLogApiKeySecretVal   = "opensesameforauditlogs"
	auditLogApiKeyDescription = "Mock Apis for Audit Logs"
	auditLogServiceAccountId  = int32(748)
	auditLogUserResourceId    = "sa-55555"

	myApiKeyVal         = "user-apikey"
	myApiKeySecretVal   = "icreatedthis"
	myApiKeyDescription = "Mock Apis for User"
	myServiceAccountId  = int32(987)
	myUserResourceId    = "u-123"
	myAccountName       = "My Account"
)

var (
	apiValue = apikeysv2.IamV2ApiKey{
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Description: apikeysv2.PtrString(apiKeyDescription),
			Resource: &apikeysv2.ObjectReference{
				Id:   kafkaClusterID,
				Kind: apikeysv2.PtrString("Cluster"),
			},
			Owner: &apikeysv2.ObjectReference{
				Id: userResourceId,
			},
			Secret: apikeysv2.PtrString(apiSecretVal),
		},
		Id: apikeysv2.PtrString(apiKeyVal),
		Metadata: &apikeysv2.ObjectMeta{
			CreatedAt: &time.Time{},
		},
	}
	auditLogApiValue = apikeysv2.IamV2ApiKey{
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Description: apikeysv2.PtrString(auditLogApiKeyDescription),
			Resource: &apikeysv2.ObjectReference{
				Id:   kafkaClusterID,
				Kind: apikeysv2.PtrString("Cluster"),
			},
			Owner: &apikeysv2.ObjectReference{
				Id: auditLogUserResourceId,
			},
			Secret: apikeysv2.PtrString(auditLogApiKeySecretVal),
		},
		Id: apikeysv2.PtrString(auditLogApiKeyVal),
		Metadata: &apikeysv2.ObjectMeta{
			CreatedAt: &time.Time{},
		},
	}
	myApiValue = apikeysv2.IamV2ApiKey{
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Description: apikeysv2.PtrString(myApiKeyDescription),
			Resource: &apikeysv2.ObjectReference{
				Id:   kafkaClusterID,
				Kind: apikeysv2.PtrString("Cluster"),
			},
			Owner: &apikeysv2.ObjectReference{
				Id: myUserResourceId,
			},
			Secret: apikeysv2.PtrString(myApiKeySecretVal),
		},
		Id:       apikeysv2.PtrString(myApiKeyVal),
		Metadata: &apikeysv2.ObjectMeta{CreatedAt: &time.Time{}},
	}
)

type APITestSuite struct {
	suite.Suite
	conf                  *v1.Config
	apiKeysMock           *apikeysmock.APIKeysIamV2Api
	iamServiceAccountMock *iammock.ServiceAccountsIamV2Api
	keystore              *mock.KeyStore
	kafkaCluster          *cmkv2.CmkV2Cluster
	srCluster             *ccloudv1.SchemaRegistryCluster
	srMothershipMock      *ccloudv1mock.SchemaRegistry
	kafkaMock             *cmkmock.ClustersCmkV2Api
	ksqlmock              *ksqlmock.ClustersKsqldbcmV2Api
	isPromptPipe          bool
	userMock              *ccloudv1mock.UserInterface
}

// Require
func (suite *APITestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()

	srCluster := ctx.SchemaRegistryClusters[ctx.GetEnvironment().GetId()]
	srCluster.SrCredentials = &v1.APIKeyPair{Key: apiKeyVal, Secret: apiSecretVal}
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	// Set up audit logs
	ctx.State.Auth.Organization.AuditLog = &ccloudv1.AuditLog{
		ClusterId:        cluster.ID,
		AccountId:        "env-zy987",
		ServiceAccountId: auditLogServiceAccountId,
		TopicName:        "confluent-audit-log-events",
	}
	suite.kafkaCluster = &cmkv2.CmkV2Cluster{
		Id:   cmkv2.PtrString(cluster.ID),
		Spec: &cmkv2.CmkV2ClusterSpec{DisplayName: cmkv2.PtrString(cluster.Name)},
	}
	suite.srCluster = &ccloudv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
	suite.kafkaMock = &cmkmock.ClustersCmkV2Api{
		GetCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
			return cmkv2.ApiGetCmkV2ClusterRequest{}
		},
		GetCmkV2ClusterExecuteFunc: func(req cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return *suite.kafkaCluster, nil, nil
		},
	}
	suite.ksqlmock = &ksqlmock.ClustersKsqldbcmV2Api{
		GetKsqldbcmV2ClusterFunc: func(context.Context, string) ksqlv2.ApiGetKsqldbcmV2ClusterRequest {
			return ksqlv2.ApiGetKsqldbcmV2ClusterRequest{}
		},
		GetKsqldbcmV2ClusterExecuteFunc: func(ksqlv2.ApiGetKsqldbcmV2ClusterRequest) (ksqlv2.KsqldbcmV2Cluster, *http.Response, error) {
			cluster := ksqlv2.KsqldbcmV2Cluster{
				Id: ksqlv2.PtrString("ksql-123"),
				Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
					DisplayName: ksqlv2.PtrString("ksql"),
				},
			}
			return cluster, nil, nil
		},
	}
	suite.srMothershipMock = &ccloudv1mock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *ccloudv1.SchemaRegistryClusterConfig) (*ccloudv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClusterFunc: func(ctx context.Context, cluster *ccloudv1.SchemaRegistryCluster) (*ccloudv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClustersFunc: func(ctx context.Context, cluster *ccloudv1.SchemaRegistryCluster) (clusters []*ccloudv1.SchemaRegistryCluster, e error) {
			return []*ccloudv1.SchemaRegistryCluster{suite.srCluster}, nil
		},
	}
	suite.apiKeysMock = &apikeysmock.APIKeysIamV2Api{
		GetIamV2ApiKeyFunc: func(_ context.Context, _ string) apikeysv2.ApiGetIamV2ApiKeyRequest {
			return apikeysv2.ApiGetIamV2ApiKeyRequest{}
		},
		GetIamV2ApiKeyExecuteFunc: func(_ apikeysv2.ApiGetIamV2ApiKeyRequest) (apikeysv2.IamV2ApiKey, *http.Response, error) {
			return apiValue, nil, nil
		},
		CreateIamV2ApiKeyFunc: func(_ context.Context) apikeysv2.ApiCreateIamV2ApiKeyRequest {
			return apikeysv2.ApiCreateIamV2ApiKeyRequest{}
		},
		CreateIamV2ApiKeyExecuteFunc: func(_ apikeysv2.ApiCreateIamV2ApiKeyRequest) (apikeysv2.IamV2ApiKey, *http.Response, error) {
			return apiValue, nil, nil
		},
		DeleteIamV2ApiKeyFunc: func(_ context.Context, _ string) apikeysv2.ApiDeleteIamV2ApiKeyRequest {
			return apikeysv2.ApiDeleteIamV2ApiKeyRequest{}
		},
		DeleteIamV2ApiKeyExecuteFunc: func(_ apikeysv2.ApiDeleteIamV2ApiKeyRequest) (*http.Response, error) {
			return nil, nil
		},
		ListIamV2ApiKeysFunc: func(_ context.Context) apikeysv2.ApiListIamV2ApiKeysRequest {
			return apikeysv2.ApiListIamV2ApiKeysRequest{}
		},
		ListIamV2ApiKeysExecuteFunc: func(_ apikeysv2.ApiListIamV2ApiKeysRequest) (apikeysv2.IamV2ApiKeyList, *http.Response, error) {
			list := apikeysv2.IamV2ApiKeyList{
				Data: []apikeysv2.IamV2ApiKey{apiValue, auditLogApiValue, myApiValue},
			}
			return list, nil, nil
		},
	}
	suite.keystore = &mock.KeyStore{
		HasAPIKeyFunc: func(key, clusterId string) (b bool, e error) {
			return key == apiKeyVal, nil
		},
		StoreAPIKeyFunc: func(key *v1.APIKeyPair, clusterId string) error {
			return nil
		},
		DeleteAPIKeyFunc: func(key string) error {
			return nil
		},
	}
	suite.iamServiceAccountMock = &iammock.ServiceAccountsIamV2Api{
		ListIamV2ServiceAccountsFunc: func(_ context.Context) iamv2.ApiListIamV2ServiceAccountsRequest {
			return iamv2.ApiListIamV2ServiceAccountsRequest{}
		},
		ListIamV2ServiceAccountsExecuteFunc: func(_ iamv2.ApiListIamV2ServiceAccountsRequest) (iamv2.IamV2ServiceAccountList, *http.Response, error) {
			list := iamv2.IamV2ServiceAccountList{
				Data: []iamv2.IamV2ServiceAccount{iamv2.IamV2ServiceAccount{
					Id: iamv2.PtrString(userResourceId),
				}},
			}
			return list, nil, nil
		},
	}
	suite.userMock = &ccloudv1mock.UserInterface{
		GetServiceAccountsFunc: func(arg0 context.Context) (users []*ccloudv1.User, e error) {
			return []*ccloudv1.User{
				{
					Id:          serviceAccountId,
					ResourceId:  userResourceId,
					ServiceName: serviceAccountName,
				},
			}, nil
		},
		GetServiceAccountFunc: func(_ context.Context, _ int32) (*ccloudv1.User, error) {
			return &ccloudv1.User{
				Id:          serviceAccountId,
				ResourceId:  userResourceId,
				ServiceName: serviceAccountName,
			}, nil
		},
		ListFunc: func(_ context.Context) ([]*ccloudv1.User, error) {
			return []*ccloudv1.User{
				{
					Id:          serviceAccountId,
					ResourceId:  userResourceId,
					ServiceName: serviceAccountName,
				},
				{
					Id:          myServiceAccountId,
					ResourceId:  myUserResourceId,
					ServiceName: myAccountName,
					Email:       "cli-mock-email@confluent.io",
				},
				{
					Id:         auditLogServiceAccountId,
					ResourceId: auditLogUserResourceId,
				},
			}, nil
		},
	}
}

func (suite *APITestSuite) newCmd() *cobra.Command {
	client := &ccloudv1.Client{
		User:           suite.userMock,
		SchemaRegistry: suite.srMothershipMock,
	}
	v2Client := &ccloudv2.Client{
		ApiKeysClient: &apikeysv2.APIClient{APIKeysIamV2Api: suite.apiKeysMock},
		CmkClient:     &cmkv2.APIClient{ClustersCmkV2Api: suite.kafkaMock},
		IamClient:     &iamv2.APIClient{ServiceAccountsIamV2Api: suite.iamServiceAccountMock},
		KsqlClient:    &ksqlv2.APIClient{ClustersKsqldbcmV2Api: suite.ksqlmock},
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
	prerunner := &climock.Commander{
		FlagResolver: resolverMock,
		Client:       client,
		V2Client:     v2Client,
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
	req.True(suite.apiKeysMock.CreateIamV2ApiKeyExecuteCalled())
}

func (suite *APITestSuite) TestCreateKafkaApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--resource", suite.kafkaCluster.GetId()})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.CreateIamV2ApiKeyExecuteCalled())
}

func (suite *APITestSuite) TestCreateCloudAPIKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--resource", "cloud"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.CreateIamV2ApiKeyExecuteCalled())
}

func (suite *APITestSuite) TestDeleteApiKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"delete", apiKeyVal, "--force"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.DeleteIamV2ApiKeyExecuteCalled())
}

func (suite *APITestSuite) TestListSrApiKey() {
	cmd := suite.newCmd()
	_, err := pcmd.ExecuteCommand(cmd, "list", "--resource", srClusterID)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.ListIamV2ApiKeysExecuteCalled())
}

func (suite *APITestSuite) TestListKafkaApiKey() {
	cmd := suite.newCmd()
	_, err := pcmd.ExecuteCommand(cmd, "list", "--resource", suite.kafkaCluster.GetId())
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.ListIamV2ApiKeysExecuteCalled())
}

// Audit Log Destination Clusters are kafka clusters, however their API keys are created by internal service accounts
func (suite *APITestSuite) TestListAuditLogDestinationClusterApiKey() {
	cmd := suite.newCmd()
	out, err := pcmd.ExecuteCommand(cmd, "list", "--resource", suite.kafkaCluster.GetId())

	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.ListIamV2ApiKeysExecuteCalled())
	req.Contains(out, "<auditlog service account>")
}

func (suite *APITestSuite) TestListCloudAPIKey() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"list", "--resource", "cloud"})
	_, err := pcmd.ExecuteCommand(cmd, "list", "--resource", "cloud")
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.ListIamV2ApiKeysExecuteCalled())
}

func (suite *APITestSuite) TestListEmails() {
	cmd := suite.newCmd()
	out, err := pcmd.ExecuteCommand(cmd, "list")
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.apiKeysMock.ListIamV2ApiKeysExecuteCalled())
	req.Contains(out, "<auditlog service account>")
	req.Contains(out, "<service account>")
	req.Contains(out, "cli-mock-email@confluent.io")
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
