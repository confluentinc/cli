package ksql

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/ccloudapis/ksql/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	kafkaClusterID    = "lkc-12345"
	ksqlClusterID     = "lksqlc-12345"
	outputTopicPrefix = "pksqlc-abcde"
	serviceAcctID     = int32(123)
	expectedACLs      = `  ServiceAccountId | Permission |    Operation     | Resource |             Name             |   Type    
+------------------+------------+------------------+----------+------------------------------+----------+
  User:123         | ALLOW      | DESCRIBE         | CLUSTER  | kafka-cluster                | LITERAL   
  User:123         | ALLOW      | DESCRIBE_CONFIGS | CLUSTER  | kafka-cluster                | LITERAL   
  User:123         | ALLOW      | CREATE           | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | CREATE           | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | CREATE           | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER            | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | ALTER            | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER            | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE_CONFIGS | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | DESCRIBE_CONFIGS | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE_CONFIGS | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER_CONFIGS    | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | ALTER_CONFIGS    | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | ALTER_CONFIGS    | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | READ             | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | READ             | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | READ             | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | WRITE            | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | WRITE            | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | WRITE            | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DELETE           | TOPIC    | pksqlc-abcde                 | PREFIXED  
  User:123         | ALLOW      | DELETE           | TOPIC    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DELETE           | GROUP    | _confluent-ksql-pksqlc-abcde | PREFIXED  
  User:123         | ALLOW      | DESCRIBE         | TOPIC    | *                            | LITERAL   
  User:123         | ALLOW      | DESCRIBE         | GROUP    | *                            | LITERAL   
  User:123         | ALLOW      | DESCRIBE_CONFIGS | TOPIC    | *                            | LITERAL   
  User:123         | ALLOW      | DESCRIBE_CONFIGS | GROUP    | *                            | LITERAL   
`
)

type KSQLTestSuite struct {
	suite.Suite
	conf         *config.Config
	kafkaCluster *kafkav1.KafkaCluster
	ksqlCluster  *v1.KSQLCluster
	serviceAcct  *orgv1.User
	ksqlc        *mock.MockKSQL
	kafkac       *mock.Kafka
	userc        *mock.User
}

func (suite *KSQLTestSuite) SetupSuite() {
	suite.initConf()
}

func (suite *KSQLTestSuite) initConf() {
	suite.conf = config.New()
	suite.conf.Logger = log.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.Auth = &config.AuthConfig{
		User:    new(orgv1.User),
		Account: &orgv1.Account{Id: "testAccount"},
	}
	user := suite.conf.Auth
	name := fmt.Sprintf("login-%s-%s", user.User.Email, suite.conf.AuthURL)

	suite.conf.Platforms[name] = &config.Platform{
		Server: suite.conf.AuthURL,
	}

	suite.conf.Credentials[name] = &config.Credential{
		Username: user.User.Email,
	}

	suite.conf.Contexts[name] = &config.Context{
		Platform:      name,
		Credential:    name,
		Kafka:         kafkaClusterID,
		KafkaClusters: map[string]*config.KafkaClusterConfig{kafkaClusterID: {}},
	}

	suite.conf.CurrentContext = name

	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         kafkaClusterID,
		Enterprise: true,
	}

	suite.ksqlCluster = &v1.KSQLCluster{
		Id:                ksqlClusterID,
		KafkaClusterId:    kafkaClusterID,
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
		CreateACLFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster, binding []*kafkav1.ACLBinding) error {
			return nil
		},
	}
	suite.ksqlc = &mock.MockKSQL{
		DescribeFunc: func(arg0 context.Context, arg1 *v1.KSQLCluster) (*v1.KSQLCluster, error) {
			return suite.ksqlCluster, nil
		},
	}
	suite.userc = &mock.User{
		GetServiceAccountsFunc: func(arg0 context.Context) (users []*orgv1.User, e error) {
			return []*orgv1.User{suite.serviceAcct}, nil
		},
	}
}

func (suite *KSQLTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.ksqlc, suite.kafkac, suite.userc, &pcmd.ConfigHelper{Config: suite.conf, Client: &ccloud.Client{Kafka: suite.kafkac}})
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	return cmd
}

func (suite *KSQLTestSuite) TestShouldConfigureACLs() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "configure-acls", ksqlClusterID}))

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.Equal(1, len(suite.kafkac.CreateACLCalls()))
	bindings := suite.kafkac.CreateACLCalls()[0].Binding
	buf := new(bytes.Buffer)
	acl.PrintAcls(bindings, buf)
	req.Equal(expectedACLs, buf.String())
}

func (suite *KSQLTestSuite) TestShouldNotConfigureForPro() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "configure-acls", ksqlClusterID}))
	suite.kafkac.DescribeFunc = func(ctx context.Context, cluster *kafkav1.KafkaCluster) (cluster2 *kafkav1.KafkaCluster, e error) {
		return &kafkav1.KafkaCluster{Id: kafkaClusterID, Enterprise: false}, nil
	}
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.False(suite.kafkac.CreateACLCalled())
	req.Equal("The Kafka cluster is not an enterprise cluster. ACLs cannot be set.", buf.String())
}

func (suite *KSQLTestSuite) TestShouldNotConfigureOnDryRun() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"app", "configure-acls", "--dry-run", ksqlClusterID}))
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	err := cmd.Execute()

	req := require.New(suite.T())
	req.Nil(err)
	req.False(suite.kafkac.CreateACLCalled())
	req.Equal(expectedACLs, buf.String())
}

func TestKsqlTestSuite(t *testing.T) {
	suite.Run(t, new(KSQLTestSuite))
}
