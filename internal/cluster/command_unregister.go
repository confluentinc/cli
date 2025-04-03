package cluster

import (
	"context"
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v4/pkg/cluster"
	pcluster "github.com/confluentinc/cli/v4/pkg/cluster"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type unregisterCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newUnregisterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister",
		Short: "Unregister cluster.",
		Long:  "Unregister cluster from the MDS cluster registry.",
		Args:  cobra.NoArgs,
	}

	c := &unregisterCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	cmd.RunE = c.unregister

	cmd.Flags().String("cluster-name", "", "Cluster Name.")
	cmd.Flags().AddFlagSet(pcmd.OnPremMTLSSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cluster-name"))
	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")

	return cmd
}

func (c *unregisterCommand) unregister(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())

	clusterName, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return err
	}

	clusterInfos, httpResp, err := client.ClusterRegistryApi.ClusterRegistryList(ctx, &mdsv1.ClusterRegistryListOpts{})
	if err != nil {
		return pcluster.HandleClusterError(err, httpResp)
	}

	found := slices.ContainsFunc(clusterInfos, func(cluster mdsv1.ClusterInfo) bool {
		return cluster.ClusterName == clusterName
	})
	if !found {
		return fmt.Errorf(`unknown cluster "%s"`, clusterName)
	}

	httpResp, err = client.ClusterRegistryApi.DeleteNamedCluster(ctx, clusterName)
	if err != nil {
		return cluster.HandleClusterError(err, httpResp)
	}

	output.Printf(c.Config.EnableColor, "Successfully unregistered cluster \"%s\" from the Cluster Registry.\n", clusterName)
	return nil
}
