package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more self-managed keys.",
		Long:              "Delete one or more self-managed keys from Confluent Cloud.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if r, err := c.V2Client.DeleteByokKey(id); err != nil {
			return errors.CatchByokKeyNotFoundError(err, r)
		}
		return nil
	}

	if _, err := resource.Delete(args, deleteFunc, resource.ByokKey); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.ByokKeyNotFoundSuggestions)
	}

	return nil
}

func (c *command) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	describeFunc := func(id string) error {
		_, _, err := c.V2Client.GetByokKey(id)
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.ByokKey, describeFunc); err != nil {
		return false, err
	}

	return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.ByokKey, args))
}
