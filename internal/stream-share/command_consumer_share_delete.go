package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newConsumerShareDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more consumer shares.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConsumerShareArgsMultiple),
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
	existenceFunc := func(id string) bool {
		_, err := c.V2Client.DescribeConsumerShare(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.ConsumerShare); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteConsumerShare(id)
	}

	_, err := deletion.Delete(args, deleteFunc, resource.ConsumerShare)
	return err
}
