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
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	if err := c.validateArgs(cmd, args); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ApiKey, args); err != nil || !ok {
		return err
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if r, err := c.V2Client.DeleteApiKey(id); err != nil {
			errs = errors.Join(errs, errors.CatchApiKeyForbiddenAccessError(err, deleteOperation, id, r))
		} else {
			deleted = append(deleted, id)
			if err := c.keystore.DeleteAPIKey(id); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.ApiKey)

	if errs != nil {
		return errors.NewErrorWithSuggestions(errs.Error(), errors.APIKeyNotFoundSuggestions)
	}

	return nil
}

func (c *command) validateArgs(cmd *cobra.Command, args []string) error {
	describeFunc := func(id string) error {
		_, _, err := c.V2Client.GetApiKey(id)
		return err
	}

	return deletion.ValidateArgsForDeletion(cmd, args, resource.ApiKey, describeFunc)
}
