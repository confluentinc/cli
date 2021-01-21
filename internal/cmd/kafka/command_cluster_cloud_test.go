package kafka

import (
	"context"
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
	v1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	clusterId   = "lkc-0000"
	clusterName = "testCluster"
)

type KafkaClusterTestSuite struct {
	suite.Suite
}

func (suite *KafkaClusterTestSuite) newCmd(conf *v3.Config) *clusterCommand {
	client := &ccloud.Client{
		Kafka: &ccsdkmock.Kafka{
			ListFunc: func(_ context.Context, cluster *v1.KafkaCluster) ([]*v1.KafkaCluster, error) {
				return []*v1.KafkaCluster{
					{
						Id:   clusterId,
						Name: clusterName,
					},
				}, nil
			},
		},
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, conf)
	cmd := NewClusterCommand(prerunner)
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

func TestKafkaClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClusterTestSuite))
}
