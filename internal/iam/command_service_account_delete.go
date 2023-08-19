package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/deletion"
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
	serviceAccount, _, err := c.V2Client.GetIamServiceAccount(args[0])
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, _, err := c.V2Client.GetIamServiceAccount(id)
		return err == nil
	}

	if confirm, err := deletion.ValidateAndConfirmDeletionWithName(cmd, args, existenceFunc, resource.ServiceAccount, serviceAccount.GetDisplayName()); err != nil {
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

	_, err = deletion.Delete(args, deleteFunc, resource.ServiceAccount)
	return err
}
