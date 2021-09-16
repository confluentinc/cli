package ksql

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
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
+-----------+------------+------------------+------------------+------------------------------+-------------+
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
	cmd := New(suite.conf, cliMock.NewPreRunnerMock(client, nil, nil, suite.conf), &cliMock.ServerSideCompleter{}, suite.analyticsClient)
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	return cmd
}

func (suite *KSQLTestSuite) newClusterCMD() *clusterCommand {
	client := &ccloud.Client{
		Kafka: suite.kafkac,
		User:  suite.userc,
		KSQL:  suite.ksqlc,
	}
	cmd := NewClusterCommand(cliMock.NewPreRunnerMock(client, nil, nil, suite.conf), suite.analyticsClient)
	return cmd
}

func (suite *KSQLTestSuite) TestShouldConfigureACLs() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "configure-acls", ksqlClusterID})

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(1, len(suite.kafkac.CreateACLsCalls()))
	bindings := suite.kafkac.CreateACLsCalls()[0].Bindings
	buf := new(bytes.Buffer)
	req.NoError(acl.PrintACLs(cmd, bindings, buf))
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestShouldNotConfigureAclsWhenUser() {
	cmd := suite.newCMD()
	suite.ksqlCluster.ServiceAccountId = 0
	cmd.SetArgs([]string{"app", "configure-acls", ksqlClusterID})

	err := cmd.Execute()

	req := require.New(suite.T())
	req.EqualError(err, fmt.Sprintf(errors.KsqlDBNoServiceAccount, ksqlClusterID))
	req.Equal(0, len(suite.kafkac.CreateACLsCalls()))
}

func (suite *KSQLTestSuite) TestShouldAlsoConfigureForPro() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "configure-acls", ksqlClusterID})
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

func (suite *KSQLTestSuite) TestShouldNotConfigureOnDryRun() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "configure-acls", "--dry-run", ksqlClusterID})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.False(suite.kafkac.CreateACLsCalled())
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestCreateKSQL() {
	cmd := suite.newCMD()
	args := []string{"app", "create", ksqlClusterID, "--api-key", keyString, "--api-secret", keySecretString}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.CreateCalled())
	cfg := suite.ksqlc.CreateCalls()[0].Arg1
	req.Equal("", cfg.Image)
	req.Equal(uint32(4), cfg.TotalNumCsu)
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], ksqlClusterID, req)
}

func (suite *KSQLTestSuite) TestCreateKSQLWithApiKey() {
	cmd := suite.newCMD()
	args := []string{"app", "create", ksqlClusterID, "--api-key", keyString, "--api-secret", keySecretString}

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

func (suite *KSQLTestSuite) TestCreateKSQLWithApiKeyMissingKey() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "create", ksqlClusterID, "--api-secret", keySecretString})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err)
	req.False(suite.ksqlc.CreateCalled())
	req.Equal("required flag(s) \"api-key\" not set", err.Error())
}

func (suite *KSQLTestSuite) TestCreateKSQLWithApiKeyMissingSecret() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "create", ksqlClusterID, "--api-key", keyString})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err)
	req.False(suite.ksqlc.CreateCalled())
	req.Equal("required flag(s) \"api-secret\" not set", err.Error())
}

func (suite *KSQLTestSuite) TestCreateKSQLWithApiKeyMissingKeyAndSecret() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "create", ksqlClusterID})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err)
	req.False(suite.ksqlc.CreateCalled())
	req.Equal(`required flag(s) "api-key", "api-secret" not set`, err.Error())
}

func (suite *KSQLTestSuite) TestCreateKSQLWithImage() {
	cmd := suite.newCMD()
	args := []string{"app", "create", ksqlClusterID, "--api-key", keyString, "--api-secret", keySecretString, "--image", "foo"}

	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	cfg := suite.ksqlc.CreateCalls()[0].Arg1
	req.Equal("foo", cfg.Image)
}

func (suite *KSQLTestSuite) TestDescribeKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "describe", ksqlClusterID})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DescribeCalled())
}

func (suite *KSQLTestSuite) TestListKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"app", "list"})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.ListCalled())
}

func (suite *KSQLTestSuite) TestDeleteKSQL() {
	cmd := suite.newCMD()
	args := []string{"app", "delete", ksqlClusterID}

	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DeleteCalled())
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], ksqlClusterID, req)
}

func (suite *KSQLTestSuite) TestServerClusterFlagComplete() {
	flagName := "cluster"
	req := suite.Require()
	type fields struct {
		Command *clusterCommand
	}
	tests := []struct {
		name   string
		fields fields
		want   []prompt.Suggest
	}{
		{
			name: "suggest for flag",
			fields: fields{
				Command: suite.newClusterCMD(),
			},
			want: []prompt.Suggest{
				{
					Text:        suite.kafkaCluster.Id,
					Description: suite.kafkaCluster.Name,
				},
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			_ = tt.fields.Command.PersistentPreRunE(tt.fields.Command.Command, []string{})
			got := tt.fields.Command.ServerFlagComplete()[flagName]()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}

func TestKsqlTestSuite(t *testing.T) {
	suite.Run(t, new(KSQLTestSuite))
}
