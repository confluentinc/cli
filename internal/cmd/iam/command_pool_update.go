package iam

import (
	"fmt"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *identityPoolCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an identity pool.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the description of identity pool "op-123456".`,
				Code: `confluent iam pool update op-123456 --description "Update demo identity pool information."`,
			},
		),
	}

	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")
	cmd.Flags().String("description", "", "Description of the identity pool.")

	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func (c *identityPoolCommand) update(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	if resource.LookupType(args[0]) != resource.IdentityPool {
		return fmt.Errorf(errors.BadResourceIDErrorMsg, resource.IdentityPoolPrefix)
	}
	identityPoolId := args[0]
	updateIdentityPool := identityproviderv2.IamV2IdentityPool{
		Id:          &identityPoolId,
		Description: identityproviderv2.PtrString(description),
	}
	_, httpresp, err := c.V2Client.UpdateIdentityPool(updateIdentityPool, provider)
	if err != nil {
		return errors.CatchIdentityPoolNotFoundError(err, httpresp, identityPoolId)
	}

	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "identity pool", identityPoolId, description)
	return nil
}
