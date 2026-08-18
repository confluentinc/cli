package endpoint

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more switchover endpoints.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete switchover endpoint "se-123456" in the current environment.`,
				Code: "confluent switchover endpoint delete se-123456",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetSwitchoverEndpoint(id, environmentId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.SwitchoverEndpoint); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteSwitchoverEndpoint(id, environmentId)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.SwitchoverEndpoint)
	return err
}
