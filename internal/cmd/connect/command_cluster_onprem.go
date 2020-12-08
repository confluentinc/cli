package connect

import (
	"context"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	print "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var clusterType = "connect-cluster"

type clusterCommandOnPrem struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

// NewClusterCommand returns the Cobra command for Kafka cluster.
func NewClusterCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedWithMDSStateFlagCommand(
		&cobra.Command{
			Use:   "cluster",
			Short: "Manage Connect clusters.",
		},
		prerunner, ClusterSubcommandFlags)
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
		Short: "List registered Connect clusters.",
		Long:  "List Connect clusters that are registered with the MDS cluster registry.",
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
	connectClustertype := &mds.ClusterRegistryListOpts{
		ClusterType: optional.NewString(clusterType),
	}
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), connectClustertype)
	if err != nil {
		return print.HandleClusterError(cmd, err, response)
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	return print.PrintCluster(cmd, clusterInfos, format)
}
