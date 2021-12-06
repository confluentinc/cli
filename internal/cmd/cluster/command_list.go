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

func newListCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered clusters.",
		Long:  "List clusters that are registered with the MDS cluster registry.",
	}

	c := &listCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}

	c.RunE = pcmd.NewCLIRunE(c.list)

	output.AddFlag(c.Command)

	return c.Command
}

func (c *listCommand) list(cmd *cobra.Command, _ []string) error {
	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(ctx, &mds.ClusterRegistryListOpts{})
	if err != nil {
		return print.HandleClusterError(err, response)
	}

	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	return print.PrintCluster(clusterInfos, format)
}
