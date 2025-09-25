package unifiedstreammanager

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newConnectDeregisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "deregister <usm-cluster-id-1> [usm-cluster-id-2] [usm-cluster-id-3] ... [usm-cluster-id-n]",
		Short:             "Deregister Connect clusters.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectArgsMultiple),
		RunE:              c.deregisterConnect,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Deregister a Confluent Platform Connect cluster.",
				Code: "confluent unified-stream-manager connect deregister usmcc-abc123",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) deregisterConnect(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetUsmConnectCluster(id, environmentId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.UsmConnectCluster); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteUsmConnectCluster(id, environmentId)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.UsmConnectCluster)
	return err
}
