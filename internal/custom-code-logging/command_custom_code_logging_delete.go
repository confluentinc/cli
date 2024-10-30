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

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	return cmd
}

func (c *customCodeLoggingCommand) delete(cmd *cobra.Command, args []string) error {
	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCustomCodeLogging(id)
	}
	if err := deletion.ConfirmPrompt(cmd, deletion.DefaultYesNoDeletePromptString(resource.CustomCodeLogging, args, "")); err != nil {
		return err
	}
	_, err := deletion.Delete(args, deleteFunc, resource.CustomCodeLogging)
	return err
}
