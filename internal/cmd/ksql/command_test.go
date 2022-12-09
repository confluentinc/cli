package ksql

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	ksqlmock "github.com/confluentinc/ccloud-sdk-go-v2/ksql/mock"
	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	climock "github.com/confluentinc/cli/mock"
)

const (
	ksqlClusterID         = "lksqlc-12345"
	serviceAcctID         = int32(123)
	serviceAcctResourceID = "sa-12345"
)

type KSQLTestSuite struct {
	suite.Suite
	conf            *v1.Config
	kafkaCluster    *schedv1.KafkaCluster
	ksqlCluster     *ksqlv2.KsqldbcmV2Cluster
	serviceAcct     *orgv1.User
	ksqlc           *ksqlmock.ClustersKsqldbcmV2Api
	kafkac          *mock.Kafka
	userc           *mock.User
	schemaregistryc *mock.SchemaRegistry
	v2Client        *ccloudv2.Client
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
	ksqlClusterId := ksqlClusterID
	outputTopicPrefix := "pksqlc-abcde"
	useDetailedProcessingLog := true
	suite.ksqlCluster = &ksqlv2.KsqldbcmV2Cluster{
		Id: &ksqlClusterId,
		Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
			KafkaCluster: &ksqlv2.ObjectReference{
				Id: suite.conf.Context().KafkaClusterContext.GetActiveKafkaClusterId(),
			},
			CredentialIdentity: &ksqlv2.ObjectReference{
				Id: serviceAcctResourceID,
			},
			UseDetailedProcessingLog: &useDetailedProcessingLog,
		},
		Status: &ksqlv2.KsqldbcmV2ClusterStatus{
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

	suite.ksqlc = &ksqlmock.ClustersKsqldbcmV2Api{
		GetKsqldbcmV2ClusterFunc: func(_ context.Context, _ string) ksqlv2.ApiGetKsqldbcmV2ClusterRequest {
			return ksqlv2.ApiGetKsqldbcmV2ClusterRequest{
				ApiService: suite.ksqlc,
			}
		},
		CreateKsqldbcmV2ClusterFunc: func(_ context.Context) ksqlv2.ApiCreateKsqldbcmV2ClusterRequest {
			return ksqlv2.ApiCreateKsqldbcmV2ClusterRequest{
				ApiService: suite.ksqlc,
			}
		},
		ListKsqldbcmV2ClustersFunc: func(_ context.Context) ksqlv2.ApiListKsqldbcmV2ClustersRequest {
			return ksqlv2.ApiListKsqldbcmV2ClustersRequest{
				ApiService: suite.ksqlc,
			}
		},
		DeleteKsqldbcmV2ClusterFunc: func(_ context.Context, _ string) ksqlv2.ApiDeleteKsqldbcmV2ClusterRequest {
			return ksqlv2.ApiDeleteKsqldbcmV2ClusterRequest{
				ApiService: suite.ksqlc,
			}
		},
		GetKsqldbcmV2ClusterExecuteFunc: func(_ ksqlv2.ApiGetKsqldbcmV2ClusterRequest) (ksqlv2.KsqldbcmV2Cluster, *http.Response, error) {
			return *suite.ksqlCluster, nil, nil
		},
		CreateKsqldbcmV2ClusterExecuteFunc: func(_ ksqlv2.ApiCreateKsqldbcmV2ClusterRequest) (ksqlv2.KsqldbcmV2Cluster, *http.Response, error) {
			return *suite.ksqlCluster, nil, nil
		},
		ListKsqldbcmV2ClustersExecuteFunc: func(_ ksqlv2.ApiListKsqldbcmV2ClustersRequest) (ksqlv2.KsqldbcmV2ClusterList, *http.Response, error) {
			return ksqlv2.KsqldbcmV2ClusterList{
				Data: []ksqlv2.KsqldbcmV2Cluster{*suite.ksqlCluster}}, nil, nil
		},
		DeleteKsqldbcmV2ClusterExecuteFunc: func(_ ksqlv2.ApiDeleteKsqldbcmV2ClusterRequest) (*http.Response, error) {
			return nil, nil
		},
	}
	suite.userc = &mock.User{
		GetServiceAccountsFunc: func(arg0 context.Context) (users []*orgv1.User, e error) {
			return []*orgv1.User{suite.serviceAcct}, nil
		},
	}
	suite.v2Client = &ccloudv2.Client{
		KsqlClient: &ksqlv2.APIClient{
			ClustersKsqldbcmV2Api: suite.ksqlc,
		},
	}
}

func (suite *KSQLTestSuite) newCMD() *cobra.Command {
	kafkaRestProvider := pcmd.KafkaRESTProvider(func() (*pcmd.KafkaREST, error) {
		return nil, nil
	})
	cmd := New(v1.AuthenticatedCloudConfigMock(), climock.NewPreRunnerMock(nil, nil, suite.v2Client, nil, &kafkaRestProvider, suite.conf))
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	return cmd
}

func (suite *KSQLTestSuite) TestShouldNotConfigureAclsWhenUser() {
	cmd := suite.newCMD()
	suite.ksqlCluster.Spec.CredentialIdentity.Id = "u-123"
	cmd.SetArgs([]string{"cluster", "configure-acls", ksqlClusterID})

	err := cmd.Execute()

	req := require.New(suite.T())
	req.EqualError(err, fmt.Sprintf(errors.KsqlDBNoServiceAccountErrorMsg, ksqlClusterID))
	req.Equal(0, len(suite.kafkac.CreateACLsCalls()))
}

func (suite *KSQLTestSuite) TestDescribeKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "describe", ksqlClusterID})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.GetKsqldbcmV2ClusterCalled())
	req.True(suite.ksqlc.GetKsqldbcmV2ClusterExecuteCalled())
	req.Equal(ksqlClusterID, suite.ksqlc.GetKsqldbcmV2ClusterCalls()[0].Id)
}

func (suite *KSQLTestSuite) TestListKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "list"})

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.ListKsqldbcmV2ClustersCalled())
	req.True(suite.ksqlc.ListKsqldbcmV2ClustersExecuteCalled())
}

func (suite *KSQLTestSuite) TestDeleteKSQL() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"cluster", "delete", ksqlClusterID, "--force"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.ksqlc.DeleteKsqldbcmV2ClusterCalled())
	req.True(suite.ksqlc.DeleteKsqldbcmV2ClusterExecuteCalled())
	req.Equal(ksqlClusterID, suite.ksqlc.DeleteKsqldbcmV2ClusterCalls()[0].Id)
}

func TestKsqlTestSuite(t *testing.T) {
	suite.Run(t, new(KSQLTestSuite))
}
