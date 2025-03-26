package connect

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v4/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *clusterCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List registered Connect clusters.",
		Long:        "List Connect clusters that are registered with the MDS cluster registry.",
		Args:        cobra.NoArgs,
		RunE:        c.listOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremMTLSSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")

	return cmd
}

func (c *clusterCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
	opts := &mdsv1.ClusterRegistryListOpts{ClusterType: optional.NewString(clusterType)}

	clusterInfos, response, err := client.ClusterRegistryApi.ClusterRegistryList(ctx, opts)
	if err != nil {
		return cluster.HandleClusterError(err, response)
	}

	return cluster.PrintClusters(cmd, clusterInfos)
}
