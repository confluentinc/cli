package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *clusterCommand) newEndpointListCommand() *cobra.Command {
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

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) endpointList(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("error retrieving configs for cluster %q", clusterId)
	}

	clusterEndpoints := clusterConfigs.Spec.GetEndpoints()

	list := output.NewList(cmd)
	for accessPointId, attributes := range clusterEndpoints {
		out := &endpointOut{
			IsCurrent:              attributes.GetHttpEndpoint() == c.Context.KafkaClusterContext.GetActiveKafkaClusterEndpoint(),
			Endpoint:               accessPointId,
			KafkaBootstrapEndpoint: attributes.GetKafkaBootstrapEndpoint(),
			HttpEndpoint:           attributes.GetHttpEndpoint(),
			ConnectionType:         attributes.GetConnectionType(),
		}

		list.Add(out)
	}

	return list.Print()
}
