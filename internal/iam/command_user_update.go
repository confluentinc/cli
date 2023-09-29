package iam

import (
	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *userCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a user.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
	}

	cmd.Flags().String("full-name", "", "The user's full name.")

	cobra.CheckErr(cmd.MarkFlagRequired("full-name"))

	return cmd
}

func (c *userCommand) update(cmd *cobra.Command, args []string) error {
	fullName, err := cmd.Flags().GetString("full-name")
	if err != nil {
		return err
	}

	resourceId := args[0]
	if resource.LookupType(resourceId) != resource.User {
		return errors.Errorf(badResourceIdErrorMsg, "u")
	}

	update := iamv2.IamV2UserUpdate{FullName: iamv2.PtrString(fullName)}
	if _, err := c.V2Client.UpdateIamUser(resourceId, update); err != nil {
		return errors.Errorf(`failed to update %s "%s": %v`, resource.User, resourceId, err)
	}

	output.ErrPrintf(errors.UpdateSuccessMsg, "full name", "user", resourceId, fullName)
	return nil
}
