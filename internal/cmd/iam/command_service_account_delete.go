package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *serviceAccountCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete service accounts.",
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
	if err := c.confirmDeletion(cmd, args); err != nil {
		return err
	}

	deleted, err := deletion.DeleteResources(args, func(id string) error {
		if err := c.V2Client.DeleteIamServiceAccount(id); err != nil {
			return errors.Errorf(errors.DeleteResourceErrorMsg, resource.ServiceAccount, id, err)
		}
		return nil
	}, deletion.DefaultPostProcess)
	deletion.PrintSuccessMsg(deleted, resource.ServiceAccount)

	return err
}

func (c *serviceAccountCommand) confirmDeletion(cmd *cobra.Command, args []string) error {
	if err := resource.ValidatePrefixes(resource.ServiceAccount, args); err != nil {
		return err
	}

	var displayName string
	describeFunc := func(id string) error {
		serviceAccount, _, err := c.V2Client.GetIamServiceAccount(id)
		if err == nil && id == args[0] {
			displayName = serviceAccount.GetDisplayName()
		}
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.ServiceAccount, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.ServiceAccount, args[0], displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ServiceAccount, args); err != nil || !ok {
			return err
		}
	}

	return nil
}
