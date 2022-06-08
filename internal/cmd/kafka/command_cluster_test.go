package kafka

import (
	"context"
	"net/http"
	"testing"
	"time"

	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"

	corev1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	clusterId     = "lkc-0000"
	clusterName   = "testCluster"
	cloudId       = "aws"
	regionId      = "us-west-2"
	environmentId = "abc"
)

var shouldError bool
var shouldPrompt bool

var cmkByokCluster = cmkv2.CmkV2Cluster{
	Spec: &cmkv2.CmkV2ClusterSpec{
		Environment: &cmkv2.ObjectReference{
			Id: environmentId,
		},
		DisplayName:  cmkv2.PtrString("gcp-byok-test"),
		Cloud:        cmkv2.PtrString("gcp"),
		Region:       cmkv2.PtrString("us-central1"),
		Config:       setCmkClusterConfig("dedicated", 1, "xyz"),
		Availability: cmkv2.PtrString(lowAvailability),
	},
	Id: cmkv2.PtrString("lkc-xyz"),
	Status: &cmkv2.CmkV2ClusterStatus{
		Cku:   cmkv2.PtrInt32(1),
		Phase: "PROVISIONING",
	},
}

var cmkExpandCluster = cmkv2.CmkV2Cluster{
	Spec: &cmkv2.CmkV2ClusterSpec{
		Environment: &cmkv2.ObjectReference{
			Id: environmentId,
		},
		DisplayName:  cmkv2.PtrString("gcp-shrink-test"),
		Cloud:        cmkv2.PtrString("gcp"),
		Region:       cmkv2.PtrString("us-central1"),
		Config:       setCmkClusterConfig("dedicated", 3, ""),
		Availability: cmkv2.PtrString(lowAvailability),
	},
	Id: cmkv2.PtrString("lkc-xyz"),
	Status: &cmkv2.CmkV2ClusterStatus{
		Cku:   cmkv2.PtrInt32(3),
		Phase: "PROVISIONED",
	},
}

type KafkaClusterTestSuite struct {
	suite.Suite
	conf            *v1.Config
	kafkaMock       *ccsdkmock.Kafka
	envMetadataMock *ccsdkmock.EnvironmentMetadata
	metricsApi      *ccsdkmock.MetricsApi
	usageLimits     *ccsdkmock.UsageLimits
	cmkClusterApi   *cmkmock.ClustersCmkV2Api
}

func (suite *KafkaClusterTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return &schedv1.KafkaCluster{
				ApiEndpoint: "api-endpoint",
			}, nil
		},
	}
	suite.cmkClusterApi = &cmkmock.ClustersCmkV2Api{
		CreateCmkV2ClusterFunc: func(ctx context.Context) cmkv2.ApiCreateCmkV2ClusterRequest {
			return cmkv2.ApiCreateCmkV2ClusterRequest{}
		},
		CreateCmkV2ClusterExecuteFunc: func(req cmkv2.ApiCreateCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkByokCluster, nil, nil
		},
		GetCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
			return cmkv2.ApiGetCmkV2ClusterRequest{}
		},
		GetCmkV2ClusterExecuteFunc: func(req cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkByokCluster, nil, nil
		},
		DeleteCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiDeleteCmkV2ClusterRequest {
			return cmkv2.ApiDeleteCmkV2ClusterRequest{}
		},
		DeleteCmkV2ClusterExecuteFunc: func(req cmkv2.ApiDeleteCmkV2ClusterRequest) (*http.Response, error) {
			return nil, nil
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

func (suite *KafkaClusterTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	client := &ccloud.Client{
		Kafka:               suite.kafkaMock,
		EnvironmentMetadata: suite.envMetadataMock,
		MetricsApi:          suite.metricsApi,
		UsageLimits:         suite.usageLimits,
	}
	cmkClient := &cmkv2.APIClient{
		ClustersCmkV2Api: suite.cmkClusterApi,
	}
	prerunner := cliMock.NewPreRunnerMock(client, &ccloudv2.Client{CmkClient: cmkClient, AuthToken: "auth-token"}, nil, nil, conf)
	return newClusterCommand(conf, prerunner)
}

func (suite *KafkaClusterTestSuite) TestClusterShrinkShouldPrompt() {
	req := require.New(suite.T())
	suite.cmkClusterApi = &cmkmock.ClustersCmkV2Api{
		GetCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
			return cmkv2.ApiGetCmkV2ClusterRequest{}
		},
		GetCmkV2ClusterExecuteFunc: func(req cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkExpandCluster, nil, nil
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
	suite.cmkClusterApi = &cmkmock.ClustersCmkV2Api{
		GetCmkV2ClusterFunc: func(ctx context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
			return cmkv2.ApiGetCmkV2ClusterRequest{}
		},
		GetCmkV2ClusterExecuteFunc: func(req cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
			return cmkExpandCluster, nil, nil
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
	req.True(suite.cmkClusterApi.CreateCmkV2ClusterCalled())
}

func (suite *KafkaClusterTestSuite) TestDeleteKafkaCluster() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"delete", clusterId})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.cmkClusterApi.DeleteCmkV2ClusterCalled())
}

func (suite *KafkaClusterTestSuite) TestGetLkcForDescribe() {
	req := require.New(suite.T())
	cmd := new(cobra.Command)
	cfg := v1.AuthenticatedCloudConfigMock()
	prerunner := &pcmd.PreRun{Config: cfg}
	c := &clusterCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.Config = dynamicconfig.New(cfg, nil, nil)
	lkc, err := c.getLkcForDescribe([]string{"lkc-123"})
	req.Equal("lkc-123", lkc)
	req.NoError(err)
	lkc, err = c.getLkcForDescribe([]string{})
	req.Equal(c.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId(), lkc)
	req.NoError(err)
	c.Config.Context().KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster = ""
	lkc, err = c.getLkcForDescribe([]string{})
	req.Equal("", lkc)
	req.Equal(errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaForDescribeSuggestions).Error(), err.Error())
}

func TestKafkaClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClusterTestSuite))
}
