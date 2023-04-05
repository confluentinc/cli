package streamshare

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newProviderShareDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete provider shares.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validProviderShareArgs),
		RunE:              c.deleteProviderShare,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete provider share "ss-12345":`,
				Code: "confluent stream-share provider share delete ss-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddSkipInvalidFlag(cmd)

	return cmd
}

func (c *command) deleteProviderShare(cmd *cobra.Command, args []string) error {
	if validArgs, err := c.validateArgsProviderShare(cmd, args); err != nil {
		return err
	} else {
		args = validArgs
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ProviderShare, args); err != nil || !ok {
		return err
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteProviderShare(id); err != nil {
			errs = errors.Join(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.ProviderShare)

	return errs
}

func (c *command) validateArgsProviderShare(cmd *cobra.Command, args []string) ([]string, error) {
	describeFunc := func(id string) error {
		_, err := c.V2Client.DescribeProviderShare(id)
		return err
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, resource.ProviderShare, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.ProviderShare, "stream-share provider share"))

	return validArgs, err
}
