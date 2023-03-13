package byok

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-N]",
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
	if err := c.checkExistence(cmd, args); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ByokKey, args); err != nil || !ok {
		return err
	}

	var errs error
	for _, id := range args {
		if httpResp, err := c.V2Client.DeleteByokKey(id); err != nil {
			errs = errors.Join(errs, errors.CatchByokKeyNotFoundError(err, id, httpResp))
		} else {
			output.ErrPrintf(errors.DeletedResourceMsg, resource.ByokKey, id)
		}
	}
	if errs != nil {
		errs = errors.NewErrorWithSuggestions(errs.Error(), errors.ByokKeyNotFoundSuggestions)
	}

	return errs
}

func (c *command) checkExistence(cmd *cobra.Command, args []string) error {
	// Single
	if len(args) == 1 {
		if _, httpResp, err := c.V2Client.GetByokKey(args[0]); err != nil {
			return errors.CatchByokKeyNotFoundError(err, args[0], httpResp)
		}
		return nil
	}

	// Multiple
	keys, err := c.V2Client.ListByokKeys("", "")
	if err != nil {
		return err
	}

	keySet := types.NewSet()
	for _, key := range keys {
		keySet.Add(key.GetId())
	}

	invalidKeys := keySet.Difference(args)
	if len(invalidKeys) > 0 {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.AccessForbiddenErrorMsg, resource.ByokKey, utils.ArrayToCommaDelimitedStringWithAnd(invalidKeys)), errors.ByokKeyNotFoundSuggestions)
	}

	return nil
}
