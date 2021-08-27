package kafka

import (
	"context"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	print "github.com/confluentinc/cli/internal/pkg/cluster"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var kafkaClusterTypeName = "kafka-cluster"

func (c *clusterCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *clusterCommand) onPremList(cmd *cobra.Command, _ []string) error {
	clustertype := &mds.ClusterRegistryListOpts{
		ClusterType: optional.NewString(kafkaClusterTypeName),
	}
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), clustertype)
	if err != nil {
		return print.HandleClusterError(cmd, err, response)
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	return print.PrintCluster(cmd, clusterInfos, format)
}
