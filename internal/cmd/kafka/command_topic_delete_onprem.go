package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

func (c *authenticatedTopicCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremDelete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the topic "my_topic" at specified cluster (providing Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.`,
				Code: "confluent kafka topic delete my_topic --url http://localhost:8082",
			}),
	}
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags

	return cmd
}

func (c *authenticatedTopicCommand) onPremDelete(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Delete Topic
	resp, err := restClient.TopicV3Api.DeleteKafkaTopic(restContext, clusterId, topicName)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	utils.Printf(cmd, errors.DeletedTopicMsg, topicName) // topic successfully deleted
	return nil
}
