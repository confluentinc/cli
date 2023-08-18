package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/resource"
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

	pcmd.AddContextFlag(cmd, c.CLICommand)
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

	_, err := resource.Delete(args, deleteFunc, resource.ServiceAccount)
	return err
}

func (c *serviceAccountCommand) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	if err := resource.ValidatePrefixes(resource.ServiceAccount, args); err != nil {
		return false, err
	}

	var displayName string
	existenceFunc := func(id string) bool {
		serviceAccount, _, err := c.V2Client.GetIamServiceAccount(id)
		if err != nil {
			return false
		}
		if id == args[0] {
			displayName = serviceAccount.GetDisplayName()
		}

		return true
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.ServiceAccount, existenceFunc); err != nil {
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
