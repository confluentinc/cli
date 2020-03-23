package ksql

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/acl"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	ksqlClusterID     = "lksqlc-12345"
	physicalClusterID = "pksqlc-zxcvb"
	outputTopicPrefix = "pksqlc-abcde"
	serviceAcctID     = int32(123)
	expectedACLs      = `  ServiceAccountId | Permission |    Operation     |     Resource     |             Name             |   Type    
+------------------+------------+------------------+------------------+------------------------------+----------+
  User:123         | ALLOW      | DESCRIBE         | CLUSTER          | kafka-cluster                | LITERAL   
  User:123         | ALLOW      | DESCRIBE_CONFIGS | CLUSTER          | kafka-cluster                | LITERAL   
  User:123         | ALLOW      | CREATE           | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | CREATE           | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | CREATE           | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER            | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | ALTER            | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER            | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE_CONFIGS | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | DESCRIBE_CONFIGS | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE_CONFIGS | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER_CONFIGS    | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | ALTER_CONFIGS    | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER_CONFIGS    | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | READ             | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | READ             | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | READ             | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | WRITE            | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | WRITE            | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | WRITE            | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DELETE           | TOPIC            | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | DELETE           | TOPIC            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DELETE           | GROUP            | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | TOPIC            | *                            | LITERAL   
  User:123         | ALLOW      | DESCRIBE         | GROUP            | *                            | LITERAL   
  User:123         | ALLOW      | DESCRIBE_CONFIGS | TOPIC            | *                            | LITERAL   
  User:123         | ALLOW      | DESCRIBE_CONFIGS | GROUP            | *                            | LITERAL   
  User:123         | ALLOW      | DESCRIBE         | TRANSACTIONAL_ID | pksqlc-zxcvb                 | LITERAL   
  User:123         | ALLOW      | WRITE            | TRANSACTIONAL_ID | pksqlc-zxcvb                 | LITERAL   
`
)

type KSQLTestSuite struct {
	suite.Suite
	conf         *v3.Config
	kafkaCluster *kafkav1.KafkaCluster
	ksqlCluster  *ksqlv1.KSQLCluster
	serviceAcct  *orgv1.User
	ksqlc        *mock.MockKSQL
	kafkac       *mock.Kafka
	userc        *mock.User
}

func (suite *KSQLTestSuite) SetupSuite() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	suite.ksqlCluster = &ksqlv1.KSQLCluster{
		Id:                ksqlClusterID,
		KafkaClusterId:    suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(),
		PhysicalClusterId: physicalClusterID,
		OutputTopicPrefix: outputTopicPrefix,
	}
	suite.serviceAcct = &orgv1.User{
		ServiceAccount: true,
		ServiceName:    "KSQL." + ksqlClusterID,
		Id:             serviceAcctID,
	}
}

func (suite *KSQLTestSuite) SetupTest() {
	suite.kafkac = &mock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
		CreateACLsFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster, binding []*kafkav1.ACLBinding) error {
			return nil
		},
	}
	suite.ksqlc = &mock.MockKSQL{
		DescribeFunc: func(arg0 context.Context, arg1 *ksqlv1.KSQLCluster) (*ksqlv1.KSQLCluster, error) {
			return suite.ksqlCluster, nil
		},
		CreateFunc: func(arg0 context.Context, arg1 *ksqlv1.KSQLClusterConfig) (*ksqlv1.KSQLCluster, error) {
			return suite.ksqlCluster, nil
		},
		ListFunc: func(arg0 context.Context, arg1 *ksqlv1.KSQLCluster) ([]*ksqlv1.KSQLCluster, error) {
			return []*ksqlv1.KSQLCluster{suite.ksqlCluster}, nil
		},
		DeleteFunc: func(arg0 context.Context, arg1 *ksqlv1.KSQLCluster) error {
			return nil
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
		KSQL:  suite.ksqlc,
	}
	cmd := New(cliMock.NewPreRunnerMock(client, nil), suite.conf)
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	return cmd
}

func (suite *KSQLTestSuite) TestShouldConfigureACLs() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "configure-acls", ksqlClusterID}))

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(1, len(suite.kafkac.CreateACLsCalls()))
	bindings := suite.kafkac.CreateACLsCalls()[0].Bindings
	buf := new(bytes.Buffer)
	req.NoError(acl.PrintAcls(cmd, bindings, buf))
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestShouldAlsoConfigureForPro() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "configure-acls", ksqlClusterID}))
	suite.kafkac.DescribeFunc = func(ctx context.Context, cluster *kafkav1.KafkaCluster) (cluster2 *kafkav1.KafkaCluster, e error) {
		return &kafkav1.KafkaCluster{Id: suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(), Enterprise: false}, nil
	}

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(1, len(suite.kafkac.CreateACLsCalls()))
	bindings := suite.kafkac.CreateACLsCalls()[0].Bindings
	buf := new(bytes.Buffer)
	req.NoError(acl.PrintAcls(cmd, bindings, buf))
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestShouldNotConfigureOnDryRun() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "configure-acls", "--dry-run", ksqlClusterID}))
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
	cmd.SetArgs(append([]string{"app", "create", ksqlClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.CreateCalled())
}

func (suite *KSQLTestSuite) TestDescribeKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "describe", ksqlClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DescribeCalled())
}

func (suite *KSQLTestSuite) TestListKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "list"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.ListCalled())
}

func (suite *KSQLTestSuite) TestDeleteKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "delete", ksqlClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DeleteCalled())
}

func TestKsqlTestSuite(t *testing.T) {
	suite.Run(t, new(KSQLTestSuite))
}
