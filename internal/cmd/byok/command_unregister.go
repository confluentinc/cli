package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUnregisterCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "unregister <id>",
		Short:             "unregister a self-managed key.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.unregister,
	}
}

func (c *command) unregister(cmd *cobra.Command, args []string) error {
	id := args[0]

	httpResp, err := c.V2Client.DeleteByokKey(id)
	if err != nil {
		return errors.CatchByokKeyNotFoundError(err, httpResp)
	}

	utils.ErrPrintf(cmd, errors.DeletedResourceMsg, resource.ByokKey, id)
	return nil
}
