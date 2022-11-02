package iam

import (
	"fmt"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c invitationCommand) newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <email>",
		Short: "Invite a user to join your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createInvitation,
	}
}

func (c invitationCommand) createInvitation(cmd *cobra.Command, args []string) error {
	email := args[0]

	if ok := utils.ValidateEmail(email); !ok {
		return errors.New(errors.BadEmailFormatErrorMsg)
	}

	req := iamv2.IamV2Invitation{Email: iamv2.PtrString(email)}

	invitation, err := c.V2Client.CreateIamInvitation(req)
	if err != nil {
		return err
	}

	utils.Println(cmd, fmt.Sprintf(errors.EmailInviteSentMsg, invitation.GetEmail()))
	return nil
}
