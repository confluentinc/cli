package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *serviceAccountCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a service account.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the description of service account "sa-123456".`,
				Code: `confluent iam service-account update sa-123456 --description "updated description"`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the service account.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("description"))

	return cmd
}

func (c *serviceAccountCommand) update(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	if resource.LookupType(args[0]) != resource.ServiceAccount {
		return fmt.Errorf(errors.BadServiceAccountIdErrorMsg)
	}
	serviceAccountId := args[0]

	update := iamv2.IamV2ServiceAccount{Description: &description}
	if _, httpResp, err := c.V2Client.UpdateIamServiceAccount(serviceAccountId, update); err != nil {
		return errors.CatchServiceAccountNotFoundError(err, httpResp, serviceAccountId)
	}

	output.ErrPrintf(c.Config.EnableColor, errors.UpdateSuccessMsg, "description", "service account", serviceAccountId, description)
	return nil
}
