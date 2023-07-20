package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *serviceAccountCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more service accounts.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete service account "sa-123456".`,
				Code: "confluent iam service-account delete sa-123456",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *serviceAccountCommand) delete(cmd *cobra.Command, args []string) error {
	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteIamServiceAccount(id); err != nil {
			return errors.Errorf(errors.DeleteResourceErrorMsg, resource.ServiceAccount, id, err)
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	resource.PrintDeleteSuccessMsg(deleted, resource.ServiceAccount)

	return err
}

func (c *serviceAccountCommand) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	if err := resource.ValidatePrefixes(resource.ServiceAccount, args); err != nil {
		return false, err
	}

	var displayName string
	describeFunc := func(id string) error {
		serviceAccount, _, err := c.V2Client.GetIamServiceAccount(id)
		if err != nil {
			return err
		}
		if id == args[0] {
			displayName = serviceAccount.GetDisplayName()
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.ServiceAccount, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.ServiceAccount, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.ServiceAccount, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}
