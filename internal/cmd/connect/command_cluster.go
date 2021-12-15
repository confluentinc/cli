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

const clusterType = "connect-cluster"

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func newClusterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage Connect clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &clusterCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner, nil),
		prerunner:                     prerunner,
	}

	c.AddCommand(c.newListCommand())

	return c.Command
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
	opts := &mds.ClusterRegistryListOpts{ClusterType: optional.NewString(clusterType)}

	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(ctx, opts)
	if err != nil {
		return print.HandleClusterError(err, response)
	}

	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	return print.PrintCluster(clusterInfos, format)
}
