package apikey

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
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

	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
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

	deletedIDs, err := resource.Delete(args, deleteFunc, resource.ApiKey)
	if err2 := c.deleteKeysFromKeyStore(deletedIDs); err2 != nil {
		err = multierror.Append(err, err2)
	}
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.APIKeyNotFoundSuggestions)
	}

	return nil
}

func (c *command) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	describeFunc := func(id string) error {
		_, _, err := c.V2Client.GetApiKey(id)
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.ApiKey, describeFunc); err != nil {
		return false, err
	}

	return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.ApiKey, args))
}

func (c *command) deleteKeysFromKeyStore(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		if err := c.keystore.DeleteAPIKey(id); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
