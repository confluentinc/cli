package kafka

import (
	"bytes"
	"context"
	"testing"
	"time"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	corev1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	clusterId   = "lkc-0000"
	clusterName = "testCluster"
	cloudId     = "aws"
	regionId    = "us-west-2"
)

var shouldError bool
var shouldPrompt bool

type KafkaClusterTestSuite struct {
	suite.Suite
	conf            *v1.Config
	kafkaMock       *ccsdkmock.Kafka
	envMetadataMock *ccsdkmock.EnvironmentMetadata
	metricsApi      *ccsdkmock.MetricsApi
	usageLimits     *ccsdkmock.UsageLimits
}

func (suite *KafkaClusterTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.kafkaMock = &ccsdkmock.Kafka{
		CreateFunc: func(ctx context.Context, config *schedv1.KafkaClusterConfig) (cluster *schedv1.KafkaCluster, e error) {
			return &schedv1.KafkaCluster{
				Id:         clusterId,
				Name:       clusterName,
				Deployment: &schedv1.Deployment{Sku: corev1.Sku_BASIC},
			}, nil
		},
		DeleteFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) error {
			return nil
		},
		ListFunc: func(_ context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
			return []*schedv1.KafkaCluster{
				{
					Id:   clusterId,
					Name: clusterName,
				},
			}, nil
		},
	}
	suite.envMetadataMock = &ccsdkmock.EnvironmentMetadata{
		GetFunc: func(arg0 context.Context) (metadata []*schedv1.CloudMetadata, e error) {
			cloudMeta := &schedv1.CloudMetadata{
				Id: cloudId,
				Regions: []*schedv1.Region{
					{
						Id:            regionId,
						IsSchedulable: true,
					},
				},
			}
			return []*schedv1.CloudMetadata{
				cloudMeta,
			}, nil
		},
	}
	suite.metricsApi = &ccsdkmock.MetricsApi{
		QueryV2Func: func(ctx context.Context, view string, query *ccloud.MetricsApiRequest, jwt string) (*ccloud.MetricsApiQueryReply, error) {
			if query.Aggregations[0].Metric != ClusterLoadMetricName {
				value := 10.0
				if shouldError {
					value = 5000
				}
				return &ccloud.MetricsApiQueryReply{
					Result: []ccloud.ApiData{
						{
							Timestamp: time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC),
							Value:     value,
							Labels:    map[string]interface{}{"metric.topic": "test-topic"},
						},
					},
				}, nil
			}
			value := 0.1
			if shouldPrompt {
				value = 0.8
			}
			return &ccloud.MetricsApiQueryReply{
				Result: []ccloud.ApiData{
					{
						Timestamp: time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC),
						Value:     value,
						Labels:    map[string]interface{}{"metric.topic": "test-topic"},
					},
				},
			}, nil
		},
	}
	suite.usageLimits = &ccsdkmock.UsageLimits{
		GetUsageLimitsFunc: func(ctx context.Context, provider ...string) (*schedv1.GetUsageLimitsReply, error) {
			return &schedv1.GetUsageLimitsReply{UsageLimits: &corev1.UsageLimits{
				TierLimits: map[string]*corev1.TierFixedLimits{
					"BASIC": {
						PartitionLimits: &corev1.KafkaPartitionLimits{},
						ClusterLimits:   &corev1.KafkaClusterLimits{},
					},
				},
				CkuLimits: map[uint32]*corev1.CKULimits{
					uint32(2): {
						NumBrokers: &corev1.IntegerUsageLimit{Limit: &corev1.IntegerUsageLimit_Value{Value: 5}},
						Storage: &corev1.IntegerUsageLimit{
							Limit: &corev1.IntegerUsageLimit_Value{Value: 500},
							Unit:  corev1.LimitUnit_GB,
						},
						NumPartitions: &corev1.IntegerUsageLimit{Limit: &corev1.IntegerUsageLimit_Value{Value: 2000}},
					},
					uint32(3): {
						NumBrokers: &corev1.IntegerUsageLimit{Limit: &corev1.IntegerUsageLimit_Value{Value: 5}},
						Storage: &corev1.IntegerUsageLimit{
							Limit: &corev1.IntegerUsageLimit_Value{Value: 1000},
							Unit:  corev1.LimitUnit_GB,
						},
						NumPartitions: &corev1.IntegerUsageLimit{Limit: &corev1.IntegerUsageLimit_Value{Value: 3000}},
					},
				},
			}}, nil
		},
	}
}

func (suite *KafkaClusterTestSuite) newCmd(conf *v1.Config) *clusterCommand {
	client := &ccloud.Client{
		Kafka:               suite.kafkaMock,
		EnvironmentMetadata: suite.envMetadataMock,
		MetricsApi:          suite.metricsApi,
		UsageLimits:         suite.usageLimits,
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	return newClusterCommand(conf, prerunner)
}

func (suite *KafkaClusterTestSuite) TestCreateGCPBYOK() {
	req := require.New(suite.T())
	root := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	kafkaMock := &ccsdkmock.Kafka{
		CreateFunc: func(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
			return &schedv1.KafkaCluster{
				Id:              "lkc-xyz",
				Name:            "gcp-byok-test",
				Region:          "us-central1",
				ServiceProvider: "gcp",
				Deployment: &schedv1.Deployment{
					Sku: corev1.Sku_DEDICATED,
				},
			}, nil
		},
	}
	idMock := &ccsdkmock.ExternalIdentity{
		CreateExternalIdentityFunc: func(_ context.Context, cloud, accountID string) (string, error) {
			return "id-xyz", nil
		},
	}
	client := &ccloud.Client{
		Kafka:            kafkaMock,
		ExternalIdentity: idMock,
		EnvironmentMetadata: &ccsdkmock.EnvironmentMetadata{
			GetFunc: func(ctx context.Context) ([]*schedv1.CloudMetadata, error) {
				return []*schedv1.CloudMetadata{{
					Id:       "gcp",
					Accounts: []*schedv1.AccountMetadata{{Id: "account-xyz"}},
					Regions:  []*schedv1.Region{{IsSchedulable: true, Id: "us-central1"}},
				}}, nil
			},
		},
	}
	root.AuthenticatedCLICommand.State = &v1.ContextState{
		Auth: &v1.AuthConfig{
			Account: &orgv1.Account{
				Id: "abc",
			},
		},
	}
	root.Client = client
	var buf bytes.Buffer
	root.SetOut(&buf)
	cmd, args, err := root.Command.Find([]string{
		"create",
		"gcp-byok-test",
	})
	req.NoError(err)
	err = cmd.ParseFlags([]string{
		"--cloud=gcp",
		"--region=us-central1",
		"--type=dedicated",
		"--cku=1",
		"--encryption-key=xyz",
	})
	req.NoError(err)
	err = root.create(cmd, args, mock.NewPromptMock(
		"y", // yes customer has granted key access
	))
	req.NoError(err)
	got, want := buf.Bytes(), []byte(`Create a role with these permissions, add the identity as a member of your key, and grant your role to the member:

Permissions:
  - cloudkms.cryptoKeyVersions.useToDecrypt
  - cloudkms.cryptoKeyVersions.useToEncrypt
  - cloudkms.cryptoKeys.get

Identity:
  id-xyz


Please confirm you've authorized the key for this identity: id-xyz (y/n): It may take up to 1 hour for the Kafka cluster to be ready. The organization admin will receive an email once the dedicated cluster is provisioned.
+---------------+---------------+
| ID            | lkc-xyz       |
| Name          | gcp-byok-test |
| Type          | DEDICATED     |
| Ingress       |             0 |
| Egress        |             0 |
| Storage       |             0 |
| Provider      | gcp           |
| Availability  | single-zone   |
| Region        | us-central1   |
| Status        | PROVISIONING  |
| Endpoint      |               |
| API Endpoint  |               |
| REST Endpoint |               |
| Cluster Size  |             0 |
+---------------+---------------+
`)
	req.True(cmp.Equal(got, want), cmp.Diff(got, want))
	req.Equal("abc", idMock.CreateExternalIdentityCalls()[0].AccountID)
	req.Equal("gcp", idMock.CreateExternalIdentityCalls()[0].Cloud)
	req.Equal("abc", kafkaMock.CreateCalls()[0].Config.AccountId)
	req.Equal("gcp", kafkaMock.CreateCalls()[0].Config.ServiceProvider)
	req.Equal("us-central1", kafkaMock.CreateCalls()[0].Config.Region)
	req.Equal("xyz", kafkaMock.CreateCalls()[0].Config.EncryptionKeyId)
	req.Equal(int32(1), kafkaMock.CreateCalls()[0].Config.Cku)
	req.Equal(corev1.Sku_DEDICATED, kafkaMock.CreateCalls()[0].Config.Deployment.Sku)
	req.False(suite.metricsApi.QueryV2Called())
}

func (suite *KafkaClusterTestSuite) TestClusterShrinkShouldPrompt() {
	req := require.New(suite.T())
	mockKafkaCluster := &schedv1.KafkaCluster{
		Id:              "lkc-xyz",
		Name:            "gcp-shrink-test",
		Region:          "us-central1",
		ServiceProvider: "gcp",
		Deployment: &schedv1.Deployment{
			Sku:      corev1.Sku_DEDICATED,
			Provider: &schedv1.Provider{Cloud: schedv1.Provider_GCP},
		},
		Cku:    3,
		Status: schedv1.ClusterStatus_UP,
	}
	suite.kafkaMock = &ccsdkmock.Kafka{
		CreateFunc: func(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
			return mockKafkaCluster, nil
		},
		UpdateFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return &schedv1.KafkaCluster{
				Id:              "lkc-xyz",
				Name:            "gcp-shrink-test",
				Region:          "us-central1",
				ServiceProvider: "gcp",
				Deployment: &schedv1.Deployment{
					Sku:      corev1.Sku_DEDICATED,
					Provider: &schedv1.Provider{Cloud: schedv1.Provider_GCP},
				},
				Cku:        3,
				PendingCku: 2,
			}, nil
		},
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return mockKafkaCluster, nil
		},
	}
	// Set variable for Metrics API mock
	shouldError = false
	shouldPrompt = true
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"update", clusterName, "--cku", "2"})
	err := cmd.Execute()
	req.Contains(err.Error(), "Cluster resize error: failed to read your confirmation")
	req.True(suite.metricsApi.QueryV2Called())
}

func (suite *KafkaClusterTestSuite) TestClusterShrinkValidationError() {
	req := require.New(suite.T())
	mockKafkaCluster := &schedv1.KafkaCluster{
		Id:              "lkc-xyz",
		Name:            "gcp-shrink-test",
		Region:          "us-central1",
		ServiceProvider: "gcp",
		Deployment: &schedv1.Deployment{
			Sku:      corev1.Sku_DEDICATED,
			Provider: &schedv1.Provider{Cloud: schedv1.Provider_GCP},
		},
		Cku:    3,
		Status: schedv1.ClusterStatus_UP,
	}
	suite.kafkaMock = &ccsdkmock.Kafka{
		CreateFunc: func(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
			return mockKafkaCluster, nil
		},
		UpdateFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return &schedv1.KafkaCluster{
				Id:              "lkc-xyz",
				Name:            "gcp-shrink-test",
				Region:          "us-central1",
				ServiceProvider: "gcp",
				Deployment: &schedv1.Deployment{
					Sku:      corev1.Sku_DEDICATED,
					Provider: &schedv1.Provider{Cloud: schedv1.Provider_GCP},
				},
				Cku:        3,
				PendingCku: 2,
			}, nil
		},
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return mockKafkaCluster, nil
		},
	}
	// Set variable for Metrics API mock
	shouldError = true
	shouldPrompt = false
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"update", clusterName, "--cku", "2"})
	err := cmd.Execute()
	req.True(suite.metricsApi.QueryV2Called())
	req.Contains(err.Error(), "cluster shrink validation error")
}

func (suite *KafkaClusterTestSuite) TestCreateKafkaCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"create", clusterName, "--cloud", cloudId, "--region", regionId})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.envMetadataMock.GetCalled())
	req.True(suite.kafkaMock.CreateCalled())
}

func (suite *KafkaClusterTestSuite) TestDeleteKafkaCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"delete", clusterId})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.kafkaMock.DeleteCalled())
}

func (suite *KafkaClusterTestSuite) TestGetLkcForDescribe() {
	req := require.New(suite.T())
	conf := v1.AuthenticatedCloudConfigMock()
	cmd := suite.newCmd(conf)
	cmd.Config = pcmd.NewDynamicConfig(conf, nil, nil, nil)
	lkc, err := cmd.getLkcForDescribe([]string{"lkc-123"})
	req.Equal("lkc-123", lkc)
	req.NoError(err)
	lkc, err = cmd.getLkcForDescribe([]string{})
	req.Equal(cmd.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId(), lkc)
	req.NoError(err)
	cmd.Config.Context().KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster = ""
	lkc, err = cmd.getLkcForDescribe([]string{})
	req.Equal("", lkc)
	req.Equal(errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaForDescribeSuggestions).Error(), err.Error())
}

func TestKafkaClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClusterTestSuite))
}
