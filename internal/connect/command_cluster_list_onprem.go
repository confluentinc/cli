package connect

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v3/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
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

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	ctx := context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
	opts := &mdsv1.ClusterRegistryListOpts{ClusterType: optional.NewString(clusterType)}

	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(ctx, opts)
	if err != nil {
		return cluster.HandleClusterError(err, response)
	}

	return cluster.PrintClusters(cmd, clusterInfos)
}
