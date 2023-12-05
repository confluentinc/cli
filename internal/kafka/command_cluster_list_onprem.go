package kafka

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v3/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

const kafkaClusterTypeName = "kafka-cluster"

func (c *clusterCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List registered Kafka clusters.",
		Long:  "List Kafka clusters that are registered with the MDS cluster registry.",
		RunE:  c.listOnPrem,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	clustertype := &mdsv1.ClusterRegistryListOpts{ClusterType: optional.NewString(kafkaClusterTypeName)}

	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), clustertype)
	if err != nil {
		return cluster.HandleClusterError(err, response)
	}

	return cluster.PrintClusters(cmd, clusterInfos)
}

func (c *clusterCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
}
