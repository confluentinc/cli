package apikey

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <api-key-1> [api-key-2] ... [api-key-n]",
		Short:             "Delete one or more API keys.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	existenceFunc := func(id string) bool {
		_, _, err := c.V2Client.GetApiKey(id)
		return err == nil
	}

	if confirm, err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.ApiKey); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if r, err := c.V2Client.DeleteApiKey(id); err != nil {
			return errors.CatchApiKeyForbiddenAccessError(err, deleteOperation, r)
		}
		return nil
	}

	deletedIDs, err := deletion.Delete(args, deleteFunc, resource.ApiKey)

	errs := multierror.Append(err, c.deleteKeysFromKeyStore(deletedIDs))
	if errs.ErrorOrNil() != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.APIKeyNotFoundSuggestions)
	}

	return nil
}

func (c *command) deleteKeysFromKeyStore(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		errs = multierror.Append(errs, c.keystore.DeleteAPIKey(id))
	}

	return errs.ErrorOrNil()
}
