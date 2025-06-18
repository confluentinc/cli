package kafka

import (
	"fmt"
	"github.com/confluentinc/cli/v4/pkg/kafka"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"

	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newEndpointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		RunE:  c.endpointList,
		Short: "List Kafka Cluster endpoint.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the available Kafka Cluster endpoints with current cloud provider and region.",
				Code: "confluent kafka cluster endpoint list --cluster lkc-123456",
			},
		),
	}

	// have a --cluster flag to display endpoint of one specific cluster (use default active cluster if not specified)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointList(cmd *cobra.Command, args []string) error {
	// Add logic for displaying endpoint, layer by layer
	// check current environment... cloud... region...
	// PNI -> privatelink -> public

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	clusterId := cluster.GetId()
	if err != nil {
		return err
	}
	if clusterId == "" {
		if clusterId = c.Context.KafkaClusterContext.GetActiveKafkaClusterId(); clusterId == "" {
			return fmt.Errorf("No cluster specified and no active cluster found, please specify a Kafka cluster Id.")
		}
	}

	// display the endpoints corresponding to the specified cluster

	clusterConfigs, _, err := c.V2Client.DescribeKafkaCluster(clusterId, c.Context.GetCurrentEnvironment())
	if err != nil {
		log.CliLogger.Debugf("Error describing Kafka Cluster: %v", err)
		return fmt.Errorf("error retrieving configs for cluster %q", cluster)
	}

	clusterEndpoints := clusterConfigs.Spec.GetEndpoints()

	list := output.NewList(cmd)
	for accessPointId, attributes := range clusterEndpoints {

		// basically in the format of:
		// * | ap1pni123 | http... | PNI
		out := &endpointOut{
			IsCurrent: attributes.GetHttpEndpoint() == c.Context.KafkaClusterContext.GetActiveKafkaClusterEndpoint(),
			Endpoint:  accessPointId,
			// don't wanna display multiple long strings (KafkaBootstrapEndpoint & HttpEndpoint) as one row in output
			// between kafka_bootstrap_endpoint and http_endpoint, which one we want to display to customer? Which is more useful to display
			//KafkaBootstrapEndpoint: attributes.KafkaBootstrapEndpoint,
			HttpEndpoint:   attributes.GetHttpEndpoint(),
			ConnectionType: attributes.GetConnectionType(),
		}

		list.Add(out)
	}

	return list.Print()
}
