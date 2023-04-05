package byok

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete self-managed keys.",
		Long:              "Delete self-managed keys from Confluent Cloud.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	if err := c.validateArgs(cmd, args); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ByokKey, args); err != nil || !ok {
		return err
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if r, err := c.V2Client.DeleteByokKey(id); err != nil {
			errs = errors.Join(errs, errors.CatchByokKeyNotFoundError(err, id, r))
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.ByokKey)

	if errs != nil {
		return errors.NewErrorWithSuggestions(errs.Error(), errors.ByokKeyNotFoundSuggestions)
	}

	return nil
}

func (c *command) validateArgs(cmd *cobra.Command, args []string) error {
	describeFunc := func(id string) error {
		_, _, err := c.V2Client.GetByokKey(id)
		return err
	}

	err := deletion.ValidateArgsForDeletion(cmd, args, resource.ByokKey, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.ByokKey, "byok"))

	return err
}
