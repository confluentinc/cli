package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *userCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete users from your organization.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *userCommand) delete(cmd *cobra.Command, args []string) error {
	if err := c.confirmDeletion(cmd, args); err != nil {
		return err
	}

	deleted, err := deletion.DeleteResources(args, func(id string) error {
		if err := c.V2Client.DeleteIamUser(id); err != nil {
			return errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, id, err)
		}
		return nil
	}, deletion.DefaultPostProcess)
	deletion.PrintSuccessMsg(deleted, resource.User)

	return err
}

func (c *userCommand) confirmDeletion(cmd *cobra.Command, args []string) error {
	if err := resource.ValidatePrefixes(resource.User, args); err != nil {
		return err
	}

	var fullName string
	describeFunc := func(id string) error {
		user, err := c.V2Client.GetIamUserById(id)
		if err == nil && id == args[0] {
			fullName = user.GetFullName()
		}
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.User, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, deletion.DefaultPromptString(resource.User, args[0], fullName), fullName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, deletion.DefaultYesNoPromptString(resource.User, args)); err != nil || !ok {
			return err
		}
	}

	return nil
}
