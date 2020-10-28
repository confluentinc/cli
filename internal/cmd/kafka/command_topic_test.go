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
	topicName  = "topic"
	isInternal = false
)

type KafkaTopicTestSuite struct {
	suite.Suite
}

func (suite *KafkaTopicTestSuite) newCmd(conf *v3.Config) *kafkaTopicCommand {
	client := &ccloud.Client{
		Kafka: &ccsdkmock.Kafka{
			ListTopicsFunc: func(_ context.Context, cluster *v1.KafkaCluster) ([]*v1.TopicDescription, error) {
				return []*v1.TopicDescription{
					{
						Name:     topicName,
						Internal: isInternal,
					},
				}, nil
			},
		},
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, conf)
	cmd := NewTopicCommand(false, prerunner, nil, "id")
	return cmd
}

func (suite *KafkaTopicTestSuite) TestServerComplete() {
	req := suite.Require()
	type fields struct {
		Command *kafkaTopicCommand
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
					Text:        topicName,
					Description: "",
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

func (suite *KafkaTopicTestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"topic describe", "topic update", "topic delete"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}

func TestKafkaTopicTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaTopicTestSuite))
}
