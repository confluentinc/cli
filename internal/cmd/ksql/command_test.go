package ksql

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"

	"github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/acl"
	"github.com/confluentinc/cli/internal/pkg/analytics"
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
	expectedACLs          = `  Principal | Permission |    Operation     |   ResourceType   |         ResourceName         | PatternType  
------------+------------+------------------+------------------+------------------------------+--------------
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
	conf            *v1.Config
	kafkaCluster    *schedv1.KafkaCluster
	ksqlCluster     *schedv1.KSQLCluster
	serviceAcct     *orgv1.User
	ksqlc           *mock.KSQL
	kafkac          *mock.Kafka
	userc           *mock.User
	analyticsClient analytics.Client
	analyticsOutput []segment.Message
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
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
		CreateACLsFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster, binding []*schedv1.ACLBinding) error {
			return nil
		},
		ListFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (clusters []*schedv1.KafkaCluster, e error) {
			return []*schedv1.KafkaCluster{suite.kafkaCluster}, nil
		},
	}
	suite.ksqlc = &mock.KSQL{
		DescribeFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error) {
			return suite.ksqlCluster, nil
		},
		CreateFunc: func(arg0 context.Context, arg1 *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error) {
			return suite.ksqlCluster, nil
		},
		ListFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error) {
			return []*schedv1.KSQLCluster{suite.ksqlCluster}, nil
		},
		DeleteFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) error {
			return nil
		},
	}
	suite.userc = &mock.User{
		GetServiceAccountsFunc: func(arg0 context.Context) (users []*orgv1.User, e error) {
			return []*orgv1.User{suite.serviceAcct}, nil
		},
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *KSQLTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		Kafka: suite.kafkac,
		User:  suite.userc,
		KSQL:  suite.ksqlc,
	}
	cmd := New(v1.AuthenticatedCloudConfigMock(), cliMock.NewPreRunnerMock(client, nil, nil, suite.conf))
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
	commandString := getCommandString(isApp)

	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "configure-acls", ksqlClusterID})

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(1, len(suite.kafkac.CreateACLsCalls()))
	bindings := suite.kafkac.CreateACLsCalls()[0].Bindings
	buf := new(bytes.Buffer)
	req.NoError(acl.PrintACLs(cmd, bindings, buf))
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestAppShouldNotConfigureAclsWhenUser() {
	suite.testShouldNotConfigureAclsWhenUser(true)
}

func (suite *KSQLTestSuite) TestClusterShouldNotConfigureAclsWhenUser() {
	suite.testShouldNotConfigureAclsWhenUser(false)
}

func (suite *KSQLTestSuite) testShouldNotConfigureAclsWhenUser(isApp bool) {
	commandString := getCommandString(isApp)

	cmd := suite.newCMD()
	suite.ksqlCluster.ServiceAccountId = 0
	cmd.SetArgs([]string{commandString, "configure-acls", ksqlClusterID})

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
	commandString := getCommandString(isApp)

	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "configure-acls", ksqlClusterID})
	suite.kafkac.DescribeFunc = func(ctx context.Context, cluster *schedv1.KafkaCluster) (cluster2 *schedv1.KafkaCluster, e error) {
		return &schedv1.KafkaCluster{Id: suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(), Enterprise: false}, nil
	}

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(1, len(suite.kafkac.CreateACLsCalls()))
	bindings := suite.kafkac.CreateACLsCalls()[0].Bindings
	buf := new(bytes.Buffer)
	req.NoError(acl.PrintACLs(cmd, bindings, buf))
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestAppShouldNotConfigureOnDryRun() {
	suite.testShouldNotConfigureOnDryRun(true)
}

func (suite *KSQLTestSuite) TestClusterShouldNotConfigureOnDryRun() {
	suite.testShouldNotConfigureOnDryRun(false)
}

func (suite *KSQLTestSuite) testShouldNotConfigureOnDryRun(isApp bool) {
	commandString := getCommandString(isApp)

	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "configure-acls", "--dry-run", ksqlClusterID})
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
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	args := []string{commandString, "create", ksqlClusterID, "--api-key", keyString, "--api-secret", keySecretString}

	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.CreateCalled())
	cfg := suite.ksqlc.CreateCalls()[0].Arg1
	req.Equal("", cfg.Image)
	req.Equal(uint32(4), cfg.TotalNumCsu)
	req.Equal(keyString, cfg.KafkaApiKey.Key)
	req.Equal(keySecretString, cfg.KafkaApiKey.Secret)
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], ksqlClusterID, req)
}

func (suite *KSQLTestSuite) TestCreateKSQLAppWithApiKeyMissingKey() {
	suite.testCreateKSQLWithApiKeyMissingKey(true)
}

func (suite *KSQLTestSuite) TestCreateKSQLClusterWithApiKeyMissingKey() {
	suite.testCreateKSQLWithApiKeyMissingKey(false)
}

func (suite *KSQLTestSuite) testCreateKSQLWithApiKeyMissingKey(isApp bool) {
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "create", ksqlClusterID, "--api-secret", keySecretString})

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
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "create", ksqlClusterID, "--api-key", keyString})

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
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "create", ksqlClusterID})

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
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	args := []string{commandString, "create", ksqlClusterID, "--api-key", keyString, "--api-secret", keySecretString, "--image", "foo"}

	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
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
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "describe", ksqlClusterID})

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
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandString, "list"})

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
	commandString := getCommandString(isApp)
	cmd := suite.newCMD()
	args := []string{commandString, "delete", ksqlClusterID}

	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DeleteCalled())
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], ksqlClusterID, req)
}

func getCommandString(isApp bool) string {
	if isApp {
		return "app"
	} else {
		return "cluster"
	}
}

func TestKsqlTestSuite(t *testing.T) {
	suite.Run(t, new(KSQLTestSuite))
}
