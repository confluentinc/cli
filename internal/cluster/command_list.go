package cluster

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcluster "github.com/confluentinc/cli/v4/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
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
	cmd.RunE = c.list

	cmd.Flags().AddFlagSet(pcmd.OnPremMTLSSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")

	return cmd
}

func (c *listCommand) list(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
	clusterInfos, response, err := client.ClusterRegistryApi.ClusterRegistryList(ctx, &mdsv1.ClusterRegistryListOpts{})
	if err != nil {
		return pcluster.HandleClusterError(err, response)
	}

	return pcluster.PrintClusters(cmd, clusterInfos)
}
