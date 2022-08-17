package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newConsumerShareDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a consumer share.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConsumerShareArgs),
		RunE:              c.deleteConsumerShare,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete consumer share "ss-12345":`,
				Code: "confluent stream-share consumer share delete ss-12345",
			},
		),
	}
}

func (c *command) deleteConsumerShare(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	if httpResp, err := c.V2Client.DeleteConsumerShare(shareId); err != nil {
		return errors.CatchV2ErrorDetailWithResponse(err, httpResp)
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.ConsumerShare, shareId)
	return nil
}
