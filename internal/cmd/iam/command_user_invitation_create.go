package iam

import (
	"context"
	"fmt"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
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

	req := &flowv1.CreateInvitationRequest{
		User:           &orgv1.User{Email: email},
		SendInvitation: true,
	}

	user, err := c.Client.User.CreateInvitation(context.Background(), req)
	if err != nil {
		return err
	}

	utils.Println(cmd, fmt.Sprintf(errors.EmailInviteSentMsg, user.Email))
	return nil
}
