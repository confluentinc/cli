package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

func (c *authenticatedTopicCommand) newDeleteCommandOnPrem() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremDelete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete the topic `my_topic` at specified cluster (providing Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.",
				Code: "confluent kafka topic delete my_topic --url http://localhost:8082",
			}),
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags

	return deleteCmd
}

//Delete a Kafka topic.
//
//Usage:
//confluent kafka topic delete <topic> [flags]
//
//Flags:
//--url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
//--ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
//--client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
//--client-key-path string    Path to client private key, include for mTLS authentication.
//--no-auth                   Include if requests should be made without authentication headers, and user will not be prompted for credentials.
//--context string            CLI Context name.
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
	resp, err := restClient.TopicApi.ClustersClusterIdTopicsTopicNameDelete(restContext, clusterId, topicName)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	utils.Printf(cmd, errors.DeletedTopicMsg, topicName) // topic successfully deleted
	return nil
}
