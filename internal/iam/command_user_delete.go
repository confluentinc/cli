package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *userCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more users from your organization.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *userCommand) delete(cmd *cobra.Command, args []string) error {
	user, err := c.V2Client.GetIamUserById(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.User, args[0])
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetIamUserById(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.User, user.GetFullName()); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteIamUser(id); err != nil {
			return errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, id, err)
		}
		return nil
	}

	_, err = deletion.Delete(args, deleteFunc, resource.User)
	return err
}
