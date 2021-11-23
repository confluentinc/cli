package schemaregistry

import (
	"context"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	print "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var clusterType = "schema-registry-cluster"

func (c *clusterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List registered Schema Registry clusters.",
		Long:        "List Schema Registry clusters that are registered with the MDS cluster registry.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.list),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

	return cmd
}

func (c *clusterCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	schemaClustertype := &mds.ClusterRegistryListOpts{
		ClusterType: optional.NewString(clusterType),
	}
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), schemaClustertype)
	if err != nil {
		return print.HandleClusterError(err, response)
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	return print.PrintCluster(clusterInfos, format)
}
