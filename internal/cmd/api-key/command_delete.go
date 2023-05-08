package apikey

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <api-key-1> [api-key-2] ... [api-key-n]",
		Short:             "Delete API keys.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	if err := c.confirmDeletion(cmd, args); err != nil {
		return err
	}

	deleted, err := deletion.DeleteResources(args, func(id string) error {
		if r, err := c.V2Client.DeleteApiKey(id); err != nil {
			return errors.CatchApiKeyForbiddenAccessError(err, deleteOperation, r)
		}
		return nil
	}, func(id string) error {
		return c.keystore.DeleteAPIKey(id)
	})
	deletion.PrintSuccessMsg(deleted, resource.ApiKey)

	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.APIKeyNotFoundSuggestions)
	}
	return nil
}

func (c *command) confirmDeletion(cmd *cobra.Command, args []string) error {
	describeFunc := func(id string) error {
		_, _, err := c.V2Client.GetApiKey(id)
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.ApiKey, describeFunc); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ApiKey, args); err != nil || !ok {
		return err
	}

	return nil
}
