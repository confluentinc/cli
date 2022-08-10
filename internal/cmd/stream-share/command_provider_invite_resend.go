package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (s *inviteCommand) newResendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resend",
		Short: "Resend email invite.",
		RunE:  s.resend,
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Resend previously sent email invite "ss-12345":`,
				Code: "confluent stream-share provider invite resend ss-12345",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (s *inviteCommand) resend(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	if _, err := s.V2Client.ResendInvite(shareId); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ResendInviteMsg, shareId)
	return nil
}
