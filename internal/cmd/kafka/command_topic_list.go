package kafka

import (
	"context"
	"fmt"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (a *authenticatedTopicCommand) newListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka topics.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(a.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all topics.",
				Code: "confluent kafka topic list",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	pcmd.AddContextFlag(listCmd, a.CLICommand)
	pcmd.AddOutputFlag(listCmd)

	return listCmd
}

func (a *authenticatedTopicCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, _ := a.GetKafkaREST()
	if kafkaREST != nil {
		kafkaClusterConfig, err := a.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		topicGetResp, httpResp, err := kafkaREST.Client.TopicApi.ClustersClusterIdTopicsGet(kafkaREST.Context, lkc)

		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST is available and there was no error
			outputWriter, err := output.NewListOutputWriter(cmd, []string{"TopicName"}, []string{"Name"}, []string{"name"})
			if err != nil {
				return err
			}
			for _, topicData := range topicGetResp.Data {
				outputWriter.AddElement(&topicData)
			}
			return outputWriter.Out()
		}
	}

	// Kafka REST is not available, fall back to KafkaAPI

	resp, err := a.getTopics()
	if err != nil {
		return err
	}
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Name"}, []string{"Name"}, []string{"name"})
	if err != nil {
		return err
	}
	for _, topic := range resp {
		outputWriter.AddElement(topic)
	}
	return outputWriter.Out()
}

func (a *authenticatedTopicCommand) getTopics() ([]*schedv1.TopicDescription, error) {
	cluster, err := pcmd.KafkaCluster(a.Context)
	if err != nil {
		return []*schedv1.TopicDescription{}, err
	}

	resp, err := a.Client.Kafka.ListTopics(context.Background(), cluster)
	return resp, errors.CatchClusterNotReadyError(err, cluster.Id)
}
