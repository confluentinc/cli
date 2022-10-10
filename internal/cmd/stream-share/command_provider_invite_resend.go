package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newResendEmailInviteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resend",
		Short: "Resend an email invite.",
		RunE:  c.resendEmailInvite,
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Resend the previously sent email invite for stream share "ss-12345":`,
				Code: "confluent stream-share provider invite resend ss-12345",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) resendEmailInvite(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	httpResp, err := c.V2Client.ResendInvite(shareId)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	utils.Printf(cmd, errors.ResendInviteMsg, shareId)
	return nil
}
