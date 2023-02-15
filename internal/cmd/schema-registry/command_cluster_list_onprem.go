package schemaregistry

import (
	"context"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/spf13/cobra"

	print "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const clusterType = "schema-registry-cluster"

func (c *clusterCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List registered Schema Registry clusters.",
		Long:        "List Schema Registry clusters that are registered with the MDS cluster registry.",
		Args:        cobra.NoArgs,
		RunE:        c.onPremList,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) onPremList(cmd *cobra.Command, _ []string) error {
	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
	opts := &mds.ClusterRegistryListOpts{ClusterType: optional.NewString(clusterType)}

	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(ctx, opts)
	if err != nil {
		return print.HandleClusterError(err, response)
	}

	return print.PrintClusters(cmd, clusterInfos)
}
