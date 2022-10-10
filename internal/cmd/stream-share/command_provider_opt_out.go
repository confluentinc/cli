package streamshare

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newOptOutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "opt-out",
		Short: "Opt out of Stream Sharing.",
		RunE:  c.optOut,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Opt out of Stream Sharing:`,
				Code: "confluent stream-share provider opt-out",
			},
		),
	}
}

func (c *command) optOut(cmd *cobra.Command, _ []string) error {
	_, httpResp, err := c.V2Client.OptInOrOut(false)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	utils.Printf(cmd, errors.OptOutMsg)
	return nil
}
