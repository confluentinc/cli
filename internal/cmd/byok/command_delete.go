package byok

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	perrors "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/set"
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

	promptMsg := fmt.Sprintf(perrors.DeleteResourceConfirmYesNoMsg, resource.ByokKey, args[0])
	if len(args) > 1 {
		promptMsg = fmt.Sprintf(perrors.DeleteResourcesConfirmYesNoMsg, resource.ByokKey, utils.ArrayToCommaDelimitedStringWithAnd(args))
	}
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	var errs error
	for _, id := range args {
		if httpResp, err := c.V2Client.DeleteByokKey(id); err != nil {
			errs = errors.Join(errs, perrors.CatchByokKeyNotFoundError(err, id, httpResp))
		} else {
			output.ErrPrintf(perrors.DeletedResourceMsg, resource.ByokKey, id)
		}
	}
	if errs != nil {
		errs = perrors.NewErrorWithSuggestions(errs.Error(), perrors.ByokKeyNotFoundSuggestions)
	}

	return errs
}

func (c *command) checkExistence(cmd *cobra.Command, args []string) error {
	// Single
	if len(args) == 1 {
		if _, httpResp, err := c.V2Client.GetByokKey(args[0]); err != nil {
			return perrors.CatchByokKeyNotFoundError(err, args[0], httpResp)
		}
		return nil
	}

	// Multiple
	keys, err := c.V2Client.ListByokKeys("", "")
	if err != nil {
		return err
	}

	keySet := set.New()
	for _, key := range keys {
		keySet.Add(key.GetId())
	}

	invalidKeys := keySet.Difference(args)
	if len(invalidKeys) > 0 {
		return perrors.NewErrorWithSuggestions("self-managed keys not found or access forbidden: " + utils.ArrayToCommaDelimitedStringWithAnd(invalidKeys), perrors.ByokKeyNotFoundSuggestions)
	}

	return nil
}
