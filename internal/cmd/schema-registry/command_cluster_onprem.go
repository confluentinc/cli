package schema_registry

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

type clusterCommandOnPrem struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

// NewClusterCommand returns the Cobra command for Kafka cluster.
func NewClusterCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedWithMDSStateFlagCommand(
		&cobra.Command{
			Use:   "cluster",
			Short: "Manage Schema Registry clusters.",
		},
		prerunner, OnPremClusterSubcommandFlags)
	cmd := &clusterCommandOnPrem{
		AuthenticatedStateFlagCommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *clusterCommandOnPrem) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List registered Schema Registry clusters.",
		Long:  "List Schema Registry clusters that are registered with the MDS cluster registry.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)
}

func (c *clusterCommandOnPrem) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *clusterCommandOnPrem) list(cmd *cobra.Command, _ []string) error {
	schemaClustertype := &mds.ClusterRegistryListOpts{
		ClusterType: optional.NewString(clusterType),
	}
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), schemaClustertype)
	if err != nil {
		return print.HandleClusterError(cmd, err, response)
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	return print.PrintCluster(cmd, clusterInfos, format)
}
