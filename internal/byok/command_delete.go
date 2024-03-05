package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
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
	existenceFunc := func(id string) bool {
		_, _, err := c.V2Client.GetByokKey(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.ByokKey); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if httpResp, err := c.V2Client.DeleteByokKey(id); err != nil {
			return errors.CatchByokKeyNotFoundError(err, httpResp)
		}
		return nil
	}

	if _, err := deletion.Delete(args, deleteFunc, resource.ByokKey); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.ByokKeyNotFoundSuggestions)
	}

	return nil
}
