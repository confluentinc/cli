package kafka

import (
	"context"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/spf13/cobra"

	print "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const kafkaClusterTypeName = "kafka-cluster"

func (c *clusterCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List registered Kafka clusters.",
		Long:  "List Kafka clusters that are registered with the MDS cluster registry.",
		RunE:  c.onPremList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) onPremList(cmd *cobra.Command, _ []string) error {
	clustertype := &mds.ClusterRegistryListOpts{ClusterType: optional.NewString(kafkaClusterTypeName)}

	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), clustertype)
	if err != nil {
		return print.HandleClusterError(err, response)
	}

	return print.PrintClusters(cmd, clusterInfos)
}

func (c *clusterCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
}
