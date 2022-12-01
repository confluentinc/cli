package kafka

import (
	"context"
	"fmt"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type topicOut struct {
	Name string `human:"Name" serialized:"name"`
}

func (c *authenticatedTopicCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka topics.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all topics.",
				Code: "confluent kafka topic list",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	err = c.provisioningClusterCheck(kafkaClusterConfig.ID)
	if err != nil {
		return err
	}

	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil {
		topicGetResp, httpResp, err := kafkaREST.CloudClient.ListKafkaTopics(kafkaClusterConfig.ID)

		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}

			// Kafka REST is available and there was no error
			list := output.NewList(cmd)
			for _, topic := range topicGetResp.Data {
				list.Add(&topicOut{Name: topic.GetTopicName()})
			}
			return list.Print()
		}
	}

	// Kafka REST is not available, fall back to KafkaAPI
	topics, err := c.getTopics()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, topic := range topics {
		list.Add(&topicOut{Name: topic.GetName()})
	}
	return list.Print()
}

func (c *authenticatedTopicCommand) getTopics() ([]*schedv1.TopicDescription, error) {
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return []*schedv1.TopicDescription{}, err
	}

	resp, err := c.PrivateClient.Kafka.ListTopics(context.Background(), cluster)
	return resp, errors.CatchClusterNotReadyError(err, cluster.Id)
}
