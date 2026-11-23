package iam

import (
	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type invitationOut struct {
	Id     string `human:"ID" serialized:"id"`
	Name   string `human:"Name" serialized:"name"`
	Email  string `human:"Email" serialized:"email"`
	UserId string `human:"User" serialized:"user_id"`
	Status string `human:"Status" serialized:"status"`
}

func (c invitationCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the organization's invitations.",
		Args:  cobra.NoArgs,
		RunE:  c.listInvitations,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c invitationCommand) listInvitations(cmd *cobra.Command, _ []string) error {
	invitations, err := c.V2Client.ListIamInvitations()
	if err != nil {
		return err
	}

	// Names are best effort: skip the users request when there is nothing to enrich, and keep listing if it fails
	var names map[string]string
	if len(invitations) > 0 {
		var err error
		if names, err = userNames(c.V2Client.ListIamUsers); err != nil {
			output.ErrPrintf(c.Config.EnableColor, "[WARN] Unable to look up user names: %v\n", err)
		}
	}

	list := output.NewList(cmd)
	for _, invitation := range invitations {
		// Invitations whose user is not found get an empty name
		name := names[invitation.User.GetId()]

		list.Add(&invitationOut{
			Id:     invitation.GetId(),
			Name:   name,
			Email:  invitation.GetEmail(),
			UserId: invitation.User.GetId(),
			Status: invitation.GetStatus(),
		})
	}
	return list.Print()
}

// userNames maps user IDs to full names. It returns an empty map and the error if the users cannot be listed.
func userNames(listUsers func() ([]iamv2.IamV2User, error)) (map[string]string, error) {
	users, err := listUsers()
	if err != nil {
		return map[string]string{}, err
	}

	names := make(map[string]string, len(users))
	for _, user := range users {
		names[user.GetId()] = user.GetFullName()
	}
	return names, nil
}
