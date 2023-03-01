package iam

import (
	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *userCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
	}

	cmd.Flags().String("full-name", "", "The user's full name.")
	_ = cmd.MarkFlagRequired("full-name")

	return cmd
}

func (c *userCommand) update(cmd *cobra.Command, args []string) error {
	fullName, err := cmd.Flags().GetString("full-name")
	if err != nil {
		return err
	}

	resourceId := args[0]
	if resource.LookupType(resourceId) != resource.User {
		return errors.Errorf(errors.BadResourceIDErrorMsg, "u")
	}

	update := iamv2.IamV2UserUpdate{FullName: iamv2.PtrString(fullName)}
	if _, err := c.V2Client.UpdateIamUser(resourceId, update); err != nil {
		return errors.Errorf(errors.UpdateResourceErrorMsg, resource.User, resourceId, err)
	}

	utils.ErrPrintf(errors.UpdateSuccessMsg, "full name", "user", resourceId, fullName)
	return nil
}
