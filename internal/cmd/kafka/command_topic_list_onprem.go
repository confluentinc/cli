package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *authenticatedTopicCommand) newListCommandOnPrem() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.onPremList),
		Short: "List Kafka topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				// on-prem examples are ccloud examples + "of a specified cluster (providing Kafka REST Proxy endpoint)."
				Text: "List all topics of a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic list --url http://localhost:8082",
			},
		),
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	pcmd.AddOutputFlag(listCmd)

	return listCmd
}

//List Kafka topics.
//
//Usage:
//confluent kafka topic list [flags]
//
//Flags:
//--url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
//--ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
//--client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
//--client-key-path string    Path to client private key, include for mTLS authentication.
//--no-auth                   Include if requests should be made without authentication headers, and user will not be prompted for credentials.
//-o, --output string         Specify the output format as "human", "json", or "yaml". (default "human")
//--context string            CLI Context name.
func (c *authenticatedTopicCommand) onPremList(cmd *cobra.Command, _ []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Get Topics
	topicGetResp, resp, err := restClient.TopicApi.ClustersClusterIdTopicsGet(restContext, clusterId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	topicDatas := topicGetResp.Data

	// Create and populate output writer
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"TopicName"}, []string{"Name"}, []string{"name"})
	if err != nil {
		return err
	}
	for _, topicData := range topicDatas {
		outputWriter.AddElement(&topicData)
	}

	return outputWriter.Out()
}
