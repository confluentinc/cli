package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a network.",
		Args:  cobra.ExactArgs(1),
		// TODO: Implement autocompletion after List Network is implemented.
		// ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE: c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete Confluent network "n-abcde1".`,
				Code: `confluent network delete n-abcde1`,
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.Network, id)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	if err := c.V2Client.DeleteNetwork(environmentId, id); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.Network, id)
	return nil
}
