package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteIamUser(id); err != nil {
			return errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, id, err)
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	resource.PrintDeleteSuccessMsg(deleted, resource.User)

	return err
}

func (c *userCommand) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	if err := resource.ValidatePrefixes(resource.User, args); err != nil {
		return false, err
	}

	var fullName string
	describeFunc := func(id string) error {
		user, err := c.V2Client.GetIamUserById(id)
		if err != nil {
			return err
		}
		if id == args[0] {
			fullName = user.GetFullName()
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.User, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.User, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.User, args[0], fullName), fullName); err != nil {
		return false, err
	}

	return true, nil
}
