package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more networks.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetNetwork(environmentId, id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.Network); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteNetwork(environmentId, id); err != nil {
			return fmt.Errorf(errors.DeleteResourceErrorMsg, resource.Network, id, err)
		}
		return nil
	}

	_, err = deletion.Delete(args, deleteFunc, resource.Network)
	return err
}
