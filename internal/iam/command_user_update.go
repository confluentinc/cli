package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
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

	id := args[0]
	if resource.LookupType(id) != resource.User {
		return fmt.Errorf(badResourceIdErrorMsg, "u")
	}

	update := iamv2.IamV2UserUpdate{FullName: iamv2.PtrString(fullName)}
	if _, err := c.V2Client.UpdateIamUser(id, update); err != nil {
		return fmt.Errorf(`failed to update %s "%s": %w`, resource.User, id, err)
	}

	output.ErrPrintf(c.Config.EnableColor, errors.UpdateSuccessMsg, "full name", "user", id, fullName)
	return nil
}
