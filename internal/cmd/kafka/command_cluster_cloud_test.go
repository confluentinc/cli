package kafka

import (
	"context"
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
	segment "github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	prodv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	v1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"

	test_utils "github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	clusterId   = "lkc-0000"
	clusterName = "testCluster"
	cloudId     = "aws"
	regionId    = "us-west-2"
)

type KafkaClusterTestSuite struct {
	suite.Suite
	conf            *v3.Config
	kafkaMock       *ccsdkmock.Kafka
	envMetadataMock *ccsdkmock.EnvironmentMetadata
	analyticsOutput []segment.Message
	analyticsClient analytics.Client
}

func (suite *KafkaClusterTestSuite) SetupTest() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	suite.kafkaMock = &ccsdkmock.Kafka{
		CreateFunc: func(ctx context.Context, config *v1.KafkaClusterConfig) (cluster *v1.KafkaCluster, e error) {
			return &v1.KafkaCluster{
				Id:         clusterId,
				Name:       clusterName,
				Deployment: &v1.Deployment{Sku: prodv1.Sku_BASIC},
			}, nil
		},
		DeleteFunc: func(ctx context.Context, cluster *v1.KafkaCluster) error {
			return nil
		},
		ListFunc: func(_ context.Context, cluster *v1.KafkaCluster) ([]*v1.KafkaCluster, error) {
			return []*v1.KafkaCluster{
				{
					Id:   clusterId,
					Name: clusterName,
				},
			}, nil
		},
	}
	suite.envMetadataMock = &ccsdkmock.EnvironmentMetadata{
		GetFunc: func(arg0 context.Context) (metadata []*v1.CloudMetadata, e error) {
			cloudMeta := &v1.CloudMetadata{
				Id: cloudId,
				Regions: []*v1.Region{
					{
						Id:            regionId,
						IsSchedulable: true,
					},
				},
			}
			return []*v1.CloudMetadata{
				cloudMeta,
			}, nil
		},
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = test_utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *KafkaClusterTestSuite) newCmd(conf *v3.Config) *clusterCommand {
	client := &ccloud.Client{
		Kafka:               suite.kafkaMock,
		EnvironmentMetadata: suite.envMetadataMock,
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	cmd := NewClusterCommand(prerunner, suite.analyticsClient)
	return cmd
}

func (suite *KafkaClusterTestSuite) TestServerComplete() {
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
			name: "suggest for authenticated user",
			fields: fields{
				Command: suite.newCmd(v3.AuthenticatedCloudConfigMock()),
			},
			want: []prompt.Suggest{
				{
					Text:        clusterId,
					Description: clusterName,
				},
			},
		},
		{
			name: "don't suggest for unauthenticated user",
			fields: fields{
				suite.newCmd(v3.UnauthenticatedCloudConfigMock()),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := tt.fields.Command.ServerComplete()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}

func (suite *KafkaClusterTestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"cluster delete", "cluster describe", "cluster update", "cluster use"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}

func (suite *KafkaClusterTestSuite) TestCreateKafkaCluster() {
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	args := append([]string{"create", clusterName, "--cloud", cloudId, "--region", regionId})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.envMetadataMock.GetCalled())
	req.True(suite.kafkaMock.CreateCalled())
	test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], clusterId, req)
}

func (suite *KafkaClusterTestSuite) TestDeleteKafkaCluster() {
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	args := append([]string{"delete", clusterId})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.kafkaMock.DeleteCalled())
	test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], clusterId, req)
}

func TestKafkaClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClusterTestSuite))
}
