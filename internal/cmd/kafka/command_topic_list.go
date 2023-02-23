package kafka

import (
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type topicOut struct {
	Name string `human:"Name" serialized:"name"`
}

func (c *authenticatedTopicCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Kafka topics.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) list(cmd *cobra.Command, _ []string) error {
	topics, err := c.getTopics(cmd)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, topic := range topics {
		list.Add(&topicOut{Name: topic.GetTopicName()})
	}
	return list.Print()
}

func (c *authenticatedTopicCommand) getTopics(cmd *cobra.Command) ([]kafkarestv3.TopicData, error) {
	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	if err := c.provisioningClusterCheck(cmd, kafkaClusterConfig.ID); err != nil {
		return nil, err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil, err
	}

	topics, httpResp, err := kafkaREST.CloudClient.ListKafkaTopics(kafkaClusterConfig.ID)
	if err != nil {
		return nil, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return topics.Data, nil
}
