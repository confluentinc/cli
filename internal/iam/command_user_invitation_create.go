package iam

import (
	"regexp"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	if !validateEmail(args[0]) {
		return errors.New("invalid email structure")
	}

	req := iamv2.IamV2Invitation{Email: iamv2.PtrString(args[0])}

	invitation, err := c.V2Client.CreateIamInvitation(req)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "An email invitation has been sent to \"%s\".\n", invitation.GetEmail())
	return nil
}

func validateEmail(email string) bool {
	rgxEmail := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return rgxEmail.MatchString(email)
}
