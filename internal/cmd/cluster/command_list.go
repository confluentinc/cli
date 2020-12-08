package cluster

import (
	"context"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	print "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type listCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewListCommand returns the sub-command object for listing clusters
func NewListCommand(prerunner pcmd.PreRunner) *cobra.Command {
	listCmd := &listCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(
			&cobra.Command{
				Use:   "list",
				Short: "List registered clusters.",
				Long:  "List clusters that are registered with the MDS cluster registry.",
			},
			prerunner),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	listCmd.RunE = pcmd.NewCLIRunE(listCmd.list)
	return listCmd.Command
}

func (c *listCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *listCommand) list(cmd *cobra.Command, _ []string) error {
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), &mds.ClusterRegistryListOpts{})
	if err != nil {
		return print.HandleClusterError(cmd, err, response)
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	return print.PrintCluster(cmd, clusterInfos, format)
}
