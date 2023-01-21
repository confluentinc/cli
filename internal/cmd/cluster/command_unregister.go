package cluster

import (
	"context"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cluster"
	pcluster "github.com/confluentinc/cli/internal/pkg/cluster"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
	c.RunE = c.unregister

	c.Flags().String("cluster-name", "", "Cluster Name.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	_ = c.MarkFlagRequired("cluster-name")

	return c.Command
}

func (c *unregisterCommand) unregister(cmd *cobra.Command, _ []string) error {
	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())

	clusterName, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return err
	}

	clusterInfos, httpResp, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(ctx, &mds.ClusterRegistryListOpts{})
	if err != nil {
		return pcluster.HandleClusterError(err, httpResp)
	}
	clusterFound := false
	for _, cluster := range clusterInfos {
		if clusterName == cluster.ClusterName {
			clusterFound = true
		}
	}
	if !clusterFound {
		return errors.Errorf(errors.UnknownClusterErrorMsg, clusterName)
	}

	httpResp, err = c.MDSClient.ClusterRegistryApi.DeleteNamedCluster(ctx, clusterName)
	if err != nil {
		return cluster.HandleClusterError(err, httpResp)
	}

	utils.Printf(cmd, errors.UnregisteredClusterMsg, clusterName)
	return nil
}
