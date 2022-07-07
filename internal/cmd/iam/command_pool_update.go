package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
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
				Code: `confluent iam pool update op-123456 --description "new description."`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the identity pool.")
	cmd.Flags().String("name", "", "Name of the identity pool.")
	cmd.Flags().String("policy", "", "Policy of the identity pool.")
	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")
	cmd.Flags().String("subject-claim", "", "Subject claim of the identity pool.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("provider")

	return cmd
}

func (c *identityPoolCommand) update(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	policy, err := cmd.Flags().GetString("policy")
	if err != nil {
		return err
	}

	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	subjectClaim, err := cmd.Flags().GetString("subject-claim")
	if err != nil {
		return err
	}

	identityPoolId := args[0]
	updateIdentityPool := identityproviderv2.IamV2IdentityPool{Id: &identityPoolId}
	if name != "" {
		updateIdentityPool.DisplayName = &name
	}
	if description != "" {
		updateIdentityPool.Description = &description
	}
	if subjectClaim != "" {
		updateIdentityPool.SubjectClaim = &subjectClaim
	}
	if policy != "" {
		updateIdentityPool.Policy = &policy
	}

	resp, httpresp, err := c.V2Client.UpdateIdentityPool(updateIdentityPool, provider)
	if err != nil {
		return errors.CatchIdentityPoolNotFoundError(err, httpresp, identityPoolId)
	}

	describeIdentityPool := &identityPool{Id: *resp.Id, DisplayName: *resp.DisplayName, Description: *resp.Description, SubjectClaim: *resp.SubjectClaim, Policy: *resp.Policy}
	return output.DescribeObject(cmd, describeIdentityPool, poolListFields, poolHumanLabelMap, poolStructuredLabelMap)
}
