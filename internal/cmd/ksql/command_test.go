package ksql

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	_nethttp "net/http"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	ksqlMock "github.com/confluentinc/ccloud-sdk-go-v2-internal/ksql/mock"
	ksql "github.com/confluentinc/ccloud-sdk-go-v2-internal/ksql/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/acl"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	ksqlClusterID         = "lksqlc-12345"
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
  User:123  | ALLOW      | DESCRIBE         | TRANSACTIONAL_ID | pksqlc-abcde                 | LITERAL       
  User:123  | ALLOW      | WRITE            | TRANSACTIONAL_ID | pksqlc-abcde                 | LITERAL       
`
)

type KSQLTestSuite struct {
	suite.Suite
	conf          *v1.Config
	kafkaCluster  *schedv1.KafkaCluster
	v1ksqlCluster *schedv1.KSQLCluster
	ksqlCluster   *ksql.KsqldbcmV2Cluster
	serviceAcct   *orgv1.User
	v1ksqlc       *mock.KSQL
	ksqlc         *ksqlMock.ClustersKsqldbcmV2Api
	kafkac        *mock.Kafka
	userc         *mock.User
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
	suite.v1ksqlCluster = &schedv1.KSQLCluster{
		Id:                ksqlClusterID,
		KafkaClusterId:    suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(),
		PhysicalClusterId: outputTopicPrefix,
		OutputTopicPrefix: outputTopicPrefix,
		ServiceAccountId:  serviceAcctID,
	}
	ksqlClusterId := ksqlClusterID
	outputTopicPrefix := "pksqlc-abcde"
	useDetailedProcessingLog := true
	suite.ksqlCluster = &ksql.KsqldbcmV2Cluster{
		Id: &ksqlClusterId,
		Spec: &ksql.KsqldbcmV2ClusterSpec{
			KafkaCluster: &ksql.ObjectReference{
				Id: suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(),
			},
			CredentialIdentity: &ksql.ObjectReference{
				Id: serviceAcctResourceID,
			},
			UseDetailedProcessingLog: &useDetailedProcessingLog,
		},
		Status: &ksql.KsqldbcmV2ClusterStatus{
			TopicPrefix: &outputTopicPrefix,
		},
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

	suite.v1ksqlc = &mock.KSQL{
		DescribeFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error) {
			return suite.v1ksqlCluster, nil
		},
		CreateFunc: func(arg0 context.Context, arg1 *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error) {
			return suite.v1ksqlCluster, nil
		},
		ListFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error) {
			return []*schedv1.KSQLCluster{suite.v1ksqlCluster}, nil
		},
		DeleteFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) error {
			return nil
		},
	}

	suite.ksqlc = &ksqlMock.ClustersKsqldbcmV2Api{
		GetKsqldbcmV2ClusterFunc: func(_ context.Context, _ string) ksql.ApiGetKsqldbcmV2ClusterRequest {
			return ksql.ApiGetKsqldbcmV2ClusterRequest{
				ApiService: suite.ksqlc,
			}
		},
		CreateKsqldbcmV2ClusterFunc: func(_ context.Context) ksql.ApiCreateKsqldbcmV2ClusterRequest {
			return ksql.ApiCreateKsqldbcmV2ClusterRequest{
				ApiService: suite.ksqlc,
			}
		},
		ListKsqldbcmV2ClustersFunc: func(_ context.Context) ksql.ApiListKsqldbcmV2ClustersRequest {
			return ksql.ApiListKsqldbcmV2ClustersRequest{
				ApiService: suite.ksqlc,
			}
		},
		DeleteKsqldbcmV2ClusterFunc: func(_ context.Context, _ string) ksql.ApiDeleteKsqldbcmV2ClusterRequest {
			return ksql.ApiDeleteKsqldbcmV2ClusterRequest{
				ApiService: suite.ksqlc,
			}
		},
		GetKsqldbcmV2ClusterExecuteFunc: func(_ ksql.ApiGetKsqldbcmV2ClusterRequest) (ksql.KsqldbcmV2Cluster, *_nethttp.Response, error) {
			return *suite.ksqlCluster, nil, nil
		},
		CreateKsqldbcmV2ClusterExecuteFunc: func(_ ksql.ApiCreateKsqldbcmV2ClusterRequest) (ksql.KsqldbcmV2Cluster, *_nethttp.Response, error) {
			return *suite.ksqlCluster, nil, nil
		},
		ListKsqldbcmV2ClustersExecuteFunc: func(_ ksql.ApiListKsqldbcmV2ClustersRequest) (ksql.KsqldbcmV2ClusterList, *_nethttp.Response, error) {
			return ksql.KsqldbcmV2ClusterList{
				Data: []ksql.KsqldbcmV2Cluster{*suite.ksqlCluster}}, nil, nil
		},
		DeleteKsqldbcmV2ClusterExecuteFunc: func(_ ksql.ApiDeleteKsqldbcmV2ClusterRequest) (*_nethttp.Response, error) {
			return nil, nil
		},
	}
	suite.userc = &mock.User{
		GetServiceAccountsFunc: func(arg0 context.Context) (users []*orgv1.User, e error) {
			return []*orgv1.User{suite.serviceAcct}, nil
		},
	}
}

func (suite *KSQLTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		Kafka: suite.kafkac,
		User:  suite.userc,
		KSQL:  suite.v1ksqlc,
	}
	v2Client := &ccloudv2.Client{
		KsqlClient: &ksql.APIClient{
			ClustersKsqldbcmV2Api: suite.ksqlc,
		},
	}
	cmd := New(v1.AuthenticatedCloudConfigMock(), cliMock.NewPreRunnerMock(client, v2Client, nil, nil, suite.conf))
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
	commandName := getCommandName(isApp)

	cmd := suite.newCMD()
	suite.ksqlCluster.Spec.CredentialIdentity.Id = "u-123"
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
	req.True(suite.v1ksqlc.CreateCalled())
	cfg := suite.v1ksqlc.CreateCalls()[0].Arg1
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
	req.False(suite.v1ksqlc.CreateCalled())
	req.False(suite.ksqlc.CreateKsqldbcmV2ClusterCalled())
	req.False(suite.ksqlc.CreateKsqldbcmV2ClusterExecuteCalled())
	req.Equal("if any flags in the group [api-key api-secret] are set they must all be set; missing [api-key]", err.Error())
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
	req.False(suite.v1ksqlc.CreateCalled())
	req.False(suite.ksqlc.CreateKsqldbcmV2ClusterCalled())
	req.False(suite.ksqlc.CreateKsqldbcmV2ClusterExecuteCalled())
	req.Equal("if any flags in the group [api-key api-secret] are set they must all be set; missing [api-secret]", err.Error())
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
	req.False(suite.v1ksqlc.CreateCalled())
	req.False(suite.ksqlc.CreateKsqldbcmV2ClusterCalled())
	req.False(suite.ksqlc.CreateKsqldbcmV2ClusterExecuteCalled())

	req.EqualError(err, errors.KsqlCreateRequiresCredentials)
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
	cfg := suite.v1ksqlc.CreateCalls()[0].Arg1
	req.Equal("foo", cfg.Image)
}

func (suite *KSQLTestSuite) TestCreateKSQLClusterWithCredentialIdentity() {
	commandName := getCommandName(false)
	cmd := suite.newCMD()
	cmd.SetArgs([]string{commandName, "create", ksqlClusterID, "--credential-identity", serviceAcctResourceID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.CreateKsqldbcmV2ClusterCalled())
	req.True(suite.ksqlc.CreateKsqldbcmV2ClusterExecuteCalled())

	// call.R.ksqldbcmV2Cluster.Spec.CredentialIdentity.Id
	credentialIdentityOnRequest := reflect.Indirect(reflect.Indirect(reflect.Indirect(
		reflect.ValueOf(suite.ksqlc.CreateKsqldbcmV2ClusterExecuteCalls()[0].R,
		).FieldByName("ksqldbcmV2Cluster"),
	).FieldByName("Spec"),
	).FieldByName("CredentialIdentity"),
	).FieldByName("Id").String()

	req.Equal(serviceAcctResourceID, credentialIdentityOnRequest)
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
	req.True(suite.ksqlc.GetKsqldbcmV2ClusterCalled())
	req.True(suite.ksqlc.GetKsqldbcmV2ClusterExecuteCalled())
	req.Equal(ksqlClusterID, suite.ksqlc.GetKsqldbcmV2ClusterCalls()[0].Id)
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
	req.True(suite.ksqlc.ListKsqldbcmV2ClustersCalled())
	req.True(suite.ksqlc.ListKsqldbcmV2ClustersExecuteCalled())
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
	req.True(suite.ksqlc.DeleteKsqldbcmV2ClusterCalled())
	req.True(suite.ksqlc.DeleteKsqldbcmV2ClusterExecuteCalled())
	req.Equal(ksqlClusterID, suite.ksqlc.DeleteKsqldbcmV2ClusterCalls()[0].Id)
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
