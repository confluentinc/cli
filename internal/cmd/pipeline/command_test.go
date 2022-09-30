package pipeline

import (
	"context"
	iamMock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"
	"github.com/stretchr/testify/require"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

const (
	ksqlClusterID     = "lksqlc-12345"
	physicalClusterID = "pksqlc-zxcvb"
	outputTopicPrefix = "pksqlc-abcde"
	keyString         = "key"
	keySecretString   = "secret"
	serviceAcctID     = int32(123)
)

type StreamDesignerTestSuite struct {
	suite.Suite
	conf                  *v1.Config
	mockKafkaCluster      *schedv1.KafkaCluster
	mockKsqlCluster       *schedv1.KSQLCluster
	ksqlClient            *mock.KSQL
	kafkaClient           *mock.Kafka
	userClient            *mock.User
	iamServiceAccountMock *iamMock.ServiceAccountsIamV2Api
	sdServiceMock         *iamMock.ServiceAccountsIamV2Api
}

func (suite *StreamDesignerTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.mockKafkaCluster = &schedv1.KafkaCluster{
		Id:   "lkc-123",
		Name: "kafka",
	}
	suite.mockKsqlCluster = &schedv1.KSQLCluster{
		Id:                ksqlClusterID,
		KafkaClusterId:    suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(),
		PhysicalClusterId: physicalClusterID,
		OutputTopicPrefix: outputTopicPrefix,
		ServiceAccountId:  serviceAcctID,
	}
}

func (suite *StreamDesignerTestSuite) SetupTest() {
	suite.kafkaClient = &mock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.mockKafkaCluster, nil
		},
		CreateACLsFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster, binding []*schedv1.ACLBinding) error {
			return nil
		},
		ListFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (clusters []*schedv1.KafkaCluster, e error) {
			return []*schedv1.KafkaCluster{suite.mockKafkaCluster}, nil
		},
	}
	suite.ksqlClient = &mock.KSQL{
		DescribeFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error) {
			return suite.mockKsqlCluster, nil
		},
		CreateFunc: func(arg0 context.Context, arg1 *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error) {
			return suite.mockKsqlCluster, nil
		},
		ListFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error) {
			return []*schedv1.KSQLCluster{suite.mockKsqlCluster}, nil
		},
		DeleteFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) error {
			return nil
		},
	}
}

func (suite *StreamDesignerTestSuite) newPipelineCmd() *cobra.Command {
	client := &ccloud.Client{
		Kafka: suite.kafkaClient,
		User:  suite.userClient,
		KSQL:  suite.ksqlClient,
	}
	cmd := New(v1.AuthenticatedCloudConfigMock(), cliMock.NewPreRunnerMock(client, nil, nil, nil, suite.conf))
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	return cmd
}

func (suite *StreamDesignerTestSuite) TestList() {
	commandName := "app"
	cmd := suite.newPipelineCmd()
	cmd.SetArgs([]string{commandName, "list"})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
}
