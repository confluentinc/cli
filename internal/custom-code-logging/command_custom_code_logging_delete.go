package customcodelogging

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *customCodeLoggingCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more custom code loggings.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	return cmd
}

func (c *customCodeLoggingCommand) delete(cmd *cobra.Command, args []string) error {
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.DescribeCustomCodeLogging(id, environment)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.CustomCodeLogging); err != nil {
		return err
	}
	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCustomCodeLogging(id, environment)
	}
	_, err = deletion.Delete(cmd, args, deleteFunc, resource.CustomCodeLogging)
	return err
}
