package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newProviderShareDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete provider shares.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validProviderShareArgsMultiple),
		RunE:              c.deleteProviderShare,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete provider share "ss-12345":`,
				Code: "confluent stream-share provider share delete ss-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) deleteProviderShare(cmd *cobra.Command, args []string) error {
	if err := c.confirmDeletionProviderShare(cmd, args); err != nil {
		return err
	}

	deleted, err := deletion.DeleteResources(args, func(id string) error {
		if err := c.V2Client.DeleteProviderShare(id); err != nil {
			return err
		}
		return nil
	}, deletion.DefaultPostProcess)
	deletion.PrintSuccessMsg(deleted, resource.ProviderShare)

	return err
}

func (c *command) confirmDeletionProviderShare(cmd *cobra.Command, args []string) error {
	describeFunc := func(id string) error {
		_, err := c.V2Client.DescribeProviderShare(id)
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.ProviderShare, describeFunc); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, deletion.DefaultYesNoPromptString(resource.ProviderShare, args)); err != nil || !ok {
		return err
	}

	return nil
}
