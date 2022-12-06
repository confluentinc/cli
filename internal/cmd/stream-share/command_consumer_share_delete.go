package streamshare

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newConsumerShareDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
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

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) deleteConsumerShare(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	if _, err := c.V2Client.DescribeConsumerShare(shareId); err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.ConsumerShare, shareId)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	err := c.V2Client.DeleteConsumerShare(shareId)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.ConsumerShare, shareId)
	return nil
}
