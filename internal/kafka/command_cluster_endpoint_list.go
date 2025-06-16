package kafka

import (
	"fmt"

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

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	if cluster == "" {
		cluster = c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
		if cluster == "" {
			return fmt.Errorf("No cluster specified and no active cluster found, please specify a Kafka cluster Id.")
		}
	}

	// display the endpoints corresponding to the specified cluster

	clusterConfigs, _, err := c.V2Client.DescribeKafkaCluster(cluster, c.Context.GetCurrentEnvironment())
	if err != nil {
		log.CliLogger.Debugf("Error describing Kafka Cluster: %v", err)
		return fmt.Errorf("Error retrieving configs for cluster %q", cluster)
	}

	clusterEndpoints := clusterConfigs.GetEndpoints()

	list := output.NewList(cmd)
	for accessPointId, attributes := range clusterEndpoints {

		out := &endpointOut{
			IsCurrent:              attributes.HttpEndpoint == c.Context.KafkaClusterContext.GetActiveKafkaClusterEndpoint(),
			Endpoint:               accessPointId,
			KafkaBootstrapEndpoint: attributes.KafkaBootstrapEndpoint,
			HttpEndpoint:           attributes.HttpEndpoint,
			ConnectionType:         attributes.ConnectionType,
		}

		list.Add(out)
	}

	return list.Print()
}
