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
		Use: "update <id>",
		Short: "Update a user.",
		Args: cobra.ExactArgs(1),
		RunE: c.update,
	}

	cmd.Flags().String("full-name", "", "The user's full name.")
	_ = cmd.MarkFlagRequired("full-name")

	return cmd
}

func (c *userCommand) update(cmd *cobra.Command, args []string) error {
	full_name, err := cmd.Flags().GetString("full-name")
	if err != nil {
		return err
	}

	resourceId := args[0]
	if resource.LookupType(resourceId) != resource.User {
		return errors.New(errors.BadResourceIDErrorMsg)
	}

	update := iamv2.IamV2UserUpdate{FullName: &full_name}

	_, httpResp, err := c.V2Client.UpdateIamUser(resourceId, update)
	if err != nil {
		return errors.Errorf(`failed to update user "%s": %v`, resourceId, errors.CatchV2ErrorDetailWithResponse(err, httpResp))
	}

	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "full-name", "user", resourceId, full_name)
	return nil
}
