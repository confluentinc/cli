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

func (c *command) newProviderShareDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a provider share.",
		Args:              cobra.ExactArgs(1),
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

	return cmd
}

func (c *command) deleteProviderShare(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	_, err := c.V2Client.DescribeProviderShare(shareId)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.ProviderShare, shareId)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	err = c.V2Client.DeleteProviderShare(shareId)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.ProviderShare, shareId)
	return nil
}
