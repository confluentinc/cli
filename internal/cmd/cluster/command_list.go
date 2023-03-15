package cluster

import (
	"context"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcluster "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

	c := &listCommand{pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	c.RunE = c.list

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(c.Command)

	return c.Command
}

func (c *listCommand) list(cmd *cobra.Command, _ []string) error {
	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(ctx, &mds.ClusterRegistryListOpts{})
	if err != nil {
		return pcluster.HandleClusterError(err, response)
	}

	return pcluster.PrintClusters(cmd, clusterInfos)
}
