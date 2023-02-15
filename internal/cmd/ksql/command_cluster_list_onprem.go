package ksql

import (
	"context"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/spf13/cobra"

	pcluster "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const clusterType = "ksql-cluster"

func (c *ksqlCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered ksqlDB clusters.",
		Long:  "List ksqlDB clusters that are registered with the MDS cluster registry.",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ksqlCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
	ksqlClusterType := &mds.ClusterRegistryListOpts{ClusterType: optional.NewString(clusterType)}

	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(ctx, ksqlClusterType)
	if err != nil {
		return pcluster.HandleClusterError(err, response)
	}

	return pcluster.PrintClusters(cmd, clusterInfos)
}
