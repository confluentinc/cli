package iam

import (
	"github.com/hashicorp/go-multierror"
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

	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteIamUser(id); err != nil {
			errs = multierror.Append(errs, errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, id, err))
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.User)

	return errs.ErrorOrNil()
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

	if err := deletion.ValidateArgsForDeletion(cmd, args, resource.User, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.User, args[0], fullName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.User, args); err != nil || !ok {
			return err
		}
	}

	return nil
}
