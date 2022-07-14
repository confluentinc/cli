package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *identityPoolCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an identity pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an identity pool named "DemoIdentityPool".`,
				Code: "confluent iam pool create DemoIdentityPool --provider op-12345 --description new-description --identity-claim claims.sub --filter 'claims.iss==\"https://my.issuer.com\"'",
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the identity pool.")
	cmd.Flags().String("filter", "", "Filters which identities can authenticate using your identity pool.")
	cmd.Flags().String("identity-claim", "", "Claim specifying the external identity using this identity pool.")
	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("filter")
	_ = cmd.MarkFlagRequired("identity-claim")
	_ = cmd.MarkFlagRequired("provider")

	return cmd
}

func (c *identityPoolCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	identityClaim, err := cmd.Flags().GetString("identity-claim")
	if err != nil {
		return err
	}

	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	createIdentityPool := identityproviderv2.IamV2IdentityPool{
		DisplayName:  identityproviderv2.PtrString(name),
		Description:  identityproviderv2.PtrString(description),
		SubjectClaim: identityproviderv2.PtrString(identityClaim),
		Policy:       identityproviderv2.PtrString(filter),
	}
	resp, httpResp, err := c.V2Client.CreateIdentityPool(createIdentityPool, provider)
	if err != nil {
		return errors.CatchServiceNameInUseError(err, httpResp, name)
	}

	identityPool := &identityPool{
		Id:            *resp.Id,
		DisplayName:   *resp.DisplayName,
		Description:   *resp.Description,
		IdentityClaim: *resp.SubjectClaim,
		Filter:        *resp.Policy,
	}

	return output.DescribeObject(cmd, identityPool, identityPoolListFields, poolHumanLabelMap, poolStructuredLabelMap)
}
