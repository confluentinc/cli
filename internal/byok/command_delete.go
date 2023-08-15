package byok

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a self-managed key.",
		Long:              "Delete a self-managed key from Confluent Cloud.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.ByokKey, id)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	httpResp, err := c.V2Client.DeleteByokKey(id)
	if err != nil {
		return errors.CatchByokKeyNotFoundError(err, httpResp)
	}

	output.ErrPrintf(errors.DeletedResourceMsg, resource.ByokKey, id)
	return nil
}
