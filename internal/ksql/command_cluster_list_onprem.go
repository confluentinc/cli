package ksql

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcluster "github.com/confluentinc/cli/v4/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
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

	cmd.Flags().AddFlagSet(pcmd.OnPremMTLSSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")

	return cmd
}

func (c *ksqlCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
	ksqlClusterType := &mdsv1.ClusterRegistryListOpts{ClusterType: optional.NewString(clusterType)}

	clusterInfos, response, err := client.ClusterRegistryApi.ClusterRegistryList(ctx, ksqlClusterType)
	if err != nil {
		return pcluster.HandleClusterError(err, response)
	}

	return pcluster.PrintClusters(cmd, clusterInfos)
}
