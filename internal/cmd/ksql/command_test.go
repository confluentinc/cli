package ksql

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	kafkarestmock "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3/mock"
	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	ksqlClusterID         = "lksqlc-12345"
	physicalClusterID     = "pksqlc-zxcvb"
	outputTopicPrefix     = "pksqlc-abcde"
	keyString             = "key"
	keySecretString       = "secret"
	serviceAcctID         = int32(123)
	serviceAcctResourceID = "sa-12345"
	expectedACLs          = `  Principal | Permission |    Operation     |  Resource Type   |        Resource Name         | Pattern Type  
------------+------------+------------------+------------------+------------------------------+---------------
  User:123  | ALLOW      | DESCRIBE         | CLUSTER          | kafka-cluster                | LITERAL       
  User:123  | ALLOW      | DESCRIBE_CONFIGS | CLUSTER          | kafka-cluster                | LITERAL       
  User:123  | ALLOW      | CREATE           | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | CREATE           | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | CREATE           | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | DESCRIBE         | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | DESCRIBE         | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | DESCRIBE         | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | ALTER            | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | ALTER            | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | ALTER            | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | DESCRIBE_CONFIGS | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | DESCRIBE_CONFIGS | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | DESCRIBE_CONFIGS | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | ALTER_CONFIGS    | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | ALTER_CONFIGS    | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | ALTER_CONFIGS    | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | READ             | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | READ             | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | READ             | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | WRITE            | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | WRITE            | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | WRITE            | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | DELETE           | TOPIC            | pksqlc-abcde                 | PREFIXED      
  User:123  | ALLOW      | DELETE           | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | DELETE           | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED      
  User:123  | ALLOW      | DESCRIBE         | TOPIC            | *                            | LITERAL       
  User:123  | ALLOW      | DESCRIBE         | GROUP            | *                            | LITERAL       
  User:123  | ALLOW      | DESCRIBE_CONFIGS | TOPIC            | *                            | LITERAL       
  User:123  | ALLOW      | DESCRIBE_CONFIGS | GROUP            | *                            | LITERAL       
  User:123  | ALLOW      | DESCRIBE         | TRANSACTIONAL_ID | pksqlc-zxcvb                 | LITERAL       
  User:123  | ALLOW      | WRITE            | TRANSACTIONAL_ID | pksqlc-zxcvb                 | LITERAL       
`
)

type KSQLTestSuite struct {
	suite.Suite
	conf         *v1.Config
	kafkaCluster *schedv1.KafkaCluster
	ksqlCluster  *schedv1.KSQLCluster
	serviceAcct  *orgv1.User
	ksqlc        *mock.KSQL
	kafkac       *mock.Kafka
	userc        *mock.User
	aclc         *kafkarestmock.ACLV3Api
}

func (suite *KSQLTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:   "lkc-123",
		Name: "kafka",
	}
	suite.serviceAcct = &orgv1.User{
		ServiceAccount: true,
		ServiceName:    "KSQL." + ksqlClusterID,
		Id:             serviceAcctID,
		ResourceId:     serviceAcctResourceID,
	}
}

func (suite *KSQLTestSuite) SetupTest() {
	suite.ksqlCluster = &schedv1.KSQLCluster{
		Id:                ksqlClusterID,
		KafkaClusterId:    suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(),
		PhysicalClusterId: physicalClusterID,
		OutputTopicPrefix: outputTopicPrefix,
		ServiceAccountId:  serviceAcctID,
	}
	suite.kafkac = &mock.Kafka{
		DescribeFunc: func(_ context.Context, _ *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
		CreateACLsFunc: func(_ context.Context, _ *schedv1.KafkaCluster, binding []*schedv1.ACLBinding) error {
			return nil
		},
		ListFunc: func(_ context.Context, _ *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
			return []*schedv1.KafkaCluster{suite.kafkaCluster}, nil
		},
	}
	suite.ksqlc = &mock.KSQL{
		DescribeFunc: func(_ context.Context, _ *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error) {
			return suite.ksqlCluster, nil
		},
		CreateFunc: func(_ context.Context, _ *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error) {
			return suite.ksqlCluster, nil
		},
		ListFunc: func(_ context.Context, _ *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error) {
			return []*schedv1.KSQLCluster{suite.ksqlCluster}, nil
		},
		DeleteFunc: func(_ context.Context, _ *schedv1.KSQLCluster) error {
			return nil
		},
	}
	suite.userc = &mock.User{
		GetServiceAccountsFunc: func(_ context.Context) ([]*orgv1.User, error) {
			return []*orgv1.User{suite.serviceAcct}, nil
		},
	}
	suite.aclc = &kafkarestmock.ACLV3Api{
		CreateKafkaAclsFunc: func(_ context.Context, _ string) kafkarestv3.ApiCreateKafkaAclsRequest {
			return kafkarestv3.ApiCreateKafkaAclsRequest{}
		},
		CreateKafkaAclsExecuteFunc: func(_ kafkarestv3.ApiCreateKafkaAclsRequest) (*http.Response, error) {
			return nil, nil
		},
	}
}

func (suite *KSQLTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		Kafka: suite.kafkac,
		User:  suite.userc,
		KSQL:  suite.ksqlc,
	}

	provider := (pcmd.KafkaRESTProvider)(func() (*pcmd.KafkaREST, error) {
		ctx := context.WithValue(context.Background(), krsdk.ContextAccessToken, "dummy-bearer-token")
		client := &ccloudv2.Client{KafkaRestClient: &kafkarestv3.APIClient{ACLV3Api: suite.aclc}}
		restMock := krsdk.NewAPIClient(&krsdk.Configuration{BasePath: "/dummy-base-path"})
		restMock.ACLV3Api = cliMock.NewACLMock()

		return pcmd.NewKafkaREST(ctx, client, restMock), nil
	})

	cmd := New(v1.AuthenticatedCloudConfigMock(), cliMock.NewPreRunnerMock(client, nil, nil, &provider, suite.conf))
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	return cmd
}

func (suite *KSQLTestSuite) TestAppShouldConfigureACLs() {
	suite.testShouldConfigureACLs(true)
}

func (suite *KSQLTestSuite) TestClusterShouldConfigureACLs() {
	suite.testShouldConfigureACLs(false)
}

func (suite *KSQLTestSuite) testShouldConfigureACLs(isApp bool) {
	commandName := getCommandName(isApp)

	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "configure-acls", ksqlClusterID})

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(32, len(suite.aclc.CreateKafkaAclsCalls()))
}

func (suite *KSQLTestSuite) TestAppShouldNotConfigureAclsWhenUser() {
	suite.testShouldNotConfigureAclsWhenUser(true)
}

func (suite *KSQLTestSuite) TestClusterShouldNotConfigureAclsWhenUser() {
	suite.testShouldNotConfigureAclsWhenUser(false)
}

func (suite *KSQLTestSuite) testShouldNotConfigureAclsWhenUser(isApp bool) {
	commandName := getCommandName(isApp)

	cmd := suite.newCMD()
	suite.ksqlCluster.ServiceAccountId = 0
	cmd.SetArgs([]string{commandName, "configure-acls", ksqlClusterID})

	err := cmd.Execute()

	req := require.New(suite.T())
	req.EqualError(err, fmt.Sprintf(errors.KsqlDBNoServiceAccountErrorMsg, ksqlClusterID))
	req.Equal(0, len(suite.kafkac.CreateACLsCalls()))
}

func (suite *KSQLTestSuite) TestAppShouldAlsoConfigureForPro() {
	suite.testShouldAlsoConfigureForPro(true)
}

func (suite *KSQLTestSuite) TestClusterShouldAlsoConfigureForPro() {
	suite.testShouldAlsoConfigureForPro(false)
}

func (suite *KSQLTestSuite) testShouldAlsoConfigureForPro(isApp bool) {
	commandName := getCommandName(isApp)

	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "configure-acls", ksqlClusterID})
	suite.kafkac.DescribeFunc = func(_ context.Context, _ *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
		return &schedv1.KafkaCluster{Id: suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId()}, nil
	}

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(32, len(suite.aclc.CreateKafkaAclsCalls()))
}

func (suite *KSQLTestSuite) TestAppShouldNotConfigureOnDryRun() {
	suite.testShouldNotConfigureOnDryRun(true)
}

func (suite *KSQLTestSuite) TestClusterShouldNotConfigureOnDryRun() {
	suite.testShouldNotConfigureOnDryRun(false)
}

func (suite *KSQLTestSuite) testShouldNotConfigureOnDryRun(isApp bool) {
	commandName := getCommandName(isApp)

	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "configure-acls", "--dry-run", ksqlClusterID})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.False(suite.kafkac.CreateACLsCalled())
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestCreateKSQLAppWithApiKey() {
	suite.testCreateKSQLWithApiKey(true)
}

func (suite *KSQLTestSuite) TestCreateKSQLClusterWithApiKey() {
	suite.testCreateKSQLWithApiKey(false)
}

func (suite *KSQLTestSuite) testCreateKSQLWithApiKey(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "create", ksqlClusterID, "--api-key", keyString, "--api-secret", keySecretString})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.CreateCalled())
	cfg := suite.ksqlc.CreateCalls()[0].Arg1
	req.Equal("", cfg.Image)
	req.Equal(uint32(4), cfg.TotalNumCsu)
	req.Equal(keyString, cfg.KafkaApiKey.Key)
	req.Equal(keySecretString, cfg.KafkaApiKey.Secret)
}

func (suite *KSQLTestSuite) TestCreateKSQLAppWithApiKeyMissingKey() {
	suite.testCreateKSQLWithApiKeyMissingKey(true)
}

func (suite *KSQLTestSuite) TestCreateKSQLClusterWithApiKeyMissingKey() {
	suite.testCreateKSQLWithApiKeyMissingKey(false)
}

func (suite *KSQLTestSuite) testCreateKSQLWithApiKeyMissingKey(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "create", ksqlClusterID, "--api-secret", keySecretString})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err)
	req.False(suite.ksqlc.CreateCalled())
	req.Equal("required flag(s) \"api-key\" not set", err.Error())
}

func (suite *KSQLTestSuite) TestCreateKSQLAppWithApiKeyMissingSecret() {
	suite.testCreateKSQLWithApiKeyMissingSecret(true)
}

func (suite *KSQLTestSuite) TestCreateKSQLClusterWithApiKeyMissingSecret() {
	suite.testCreateKSQLWithApiKeyMissingSecret(false)
}

func (suite *KSQLTestSuite) testCreateKSQLWithApiKeyMissingSecret(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "create", ksqlClusterID, "--api-key", keyString})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err)
	req.False(suite.ksqlc.CreateCalled())
	req.Equal("required flag(s) \"api-secret\" not set", err.Error())
}

func (suite *KSQLTestSuite) TestCreateKSQLAppWithApiKeyMissingKeyAndSecret() {
	suite.testCreateKSQLWithApiKeyMissingKeyAndSecret(true)
}

func (suite *KSQLTestSuite) TestCreateKSQLClusterWithApiKeyMissingKeyAndSecret() {
	suite.testCreateKSQLWithApiKeyMissingKeyAndSecret(false)
}

func (suite *KSQLTestSuite) testCreateKSQLWithApiKeyMissingKeyAndSecret(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "create", ksqlClusterID})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err)
	req.False(suite.ksqlc.CreateCalled())
	req.Equal(`required flag(s) "api-key", "api-secret" not set`, err.Error())
}

func (suite *KSQLTestSuite) TestCreateKSQLAppWithImage() {
	suite.testCreateKSQLWithImage(true)
}

func (suite *KSQLTestSuite) TestCreateKSQLClusterWithImage() {
	suite.testCreateKSQLWithImage(false)
}

func (suite *KSQLTestSuite) testCreateKSQLWithImage(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "create", ksqlClusterID, "--api-key", keyString, "--api-secret", keySecretString, "--image", "foo"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	cfg := suite.ksqlc.CreateCalls()[0].Arg1
	req.Equal("foo", cfg.Image)
}

func (suite *KSQLTestSuite) TestDescribeKSQLApp() {
	suite.testDescribeKSQL(true)
}

func (suite *KSQLTestSuite) TestDescribeKSQLCluster() {
	suite.testDescribeKSQL(false)
}

func (suite *KSQLTestSuite) testDescribeKSQL(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "describe", ksqlClusterID})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DescribeCalled())
}

func (suite *KSQLTestSuite) TestListKSQLApp() {
	suite.testListKSQL(true)
}

func (suite *KSQLTestSuite) TestListKSQLCluster() {
	suite.testListKSQL(false)
}

func (suite *KSQLTestSuite) testListKSQL(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "list"})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.ListCalled())
}

func (suite *KSQLTestSuite) TestDeleteKSQLApp() {
	suite.testDeleteKSQL(true)
}

func (suite *KSQLTestSuite) TestDeleteKSQLCluster() {
	suite.testDeleteKSQL(false)
}

func (suite *KSQLTestSuite) testDeleteKSQL(isApp bool) {
	commandName := getCommandName(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "delete", ksqlClusterID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DeleteCalled())
}

func getCommandName(isApp bool) string {
	if isApp {
		return "app"
	} else {
		return "cluster"
	}
}

func TestKsqlTestSuite(t *testing.T) {
	suite.Run(t, new(KSQLTestSuite))
}
