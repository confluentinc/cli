package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *serviceAccountCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a service account.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
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
	serviceAccountId := args[0]

	if resource.LookupType(serviceAccountId) != resource.ServiceAccount {
		return errors.New(errors.BadServiceAccountIDErrorMsg)
	}

	serviceAccount, httpResp, err := c.V2Client.GetIamServiceAccount(serviceAccountId)
	if err != nil {
		return errors.CatchServiceAccountNotFoundError(err, httpResp, serviceAccountId)
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.ServiceAccount, serviceAccountId, serviceAccount.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, serviceAccount.GetDisplayName()); err != nil {
		return err
	}

	err = c.V2Client.DeleteIamServiceAccount(serviceAccountId)
	if err != nil {
		return errors.Errorf(errors.DeleteResourceErrorMsg, resource.ServiceAccount, serviceAccountId, err)
	}

	utils.ErrPrintf(cmd, errors.DeletedResourceMsg, resource.ServiceAccount, serviceAccountId)
	return nil
}
