package streamshare

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newConsumerShareDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete consumer shares.",
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
	if err := c.confirmDeletionConsumerShare(cmd, args); err != nil {
		return err
	}

	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteConsumerShare(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessMsg(deleted, resource.ConsumerShare)

	return errs.ErrorOrNil()
}

func (c *command) confirmDeletionConsumerShare(cmd *cobra.Command, args []string) error {
	describeFunc := func(id string) error {
		_, err := c.V2Client.DescribeConsumerShare(id)
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.ConsumerShare, describeFunc); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ConsumerShare, args); err != nil || !ok {
		return err
	}

	return nil
}
