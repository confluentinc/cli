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
				Code: "confluent iam pool create DemoIdentityPool --provider op-12345 --description new-description --subject-claim claims.sub --policy claims.iss",
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the identity pool.")
	cmd.Flags().String("policy", "", "Filters which identities can authenticate using your identity pool.")
	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")
	cmd.Flags().String("subject-claim", "", "Principal that is the subject of the identity pool.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("policy")
	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("subject-claim")

	return cmd
}

func (c *identityPoolCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	description, err := cmd.Flags().GetString("description")
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

	createIdentityPool := identityproviderv2.IamV2IdentityPool{
		DisplayName:  identityproviderv2.PtrString(name),
		Description:  identityproviderv2.PtrString(description),
		SubjectClaim: identityproviderv2.PtrString(subjectClaim),
		Policy:       identityproviderv2.PtrString(policy),
	}
	resp, httpResp, err := c.V2Client.CreateIdentityPool(createIdentityPool, provider)
	if err != nil {
		return errors.CatchServiceNameInUseError(err, httpResp, name)
	}

	identityPool := &identityPool{
		Id:           *resp.Id,
		DisplayName:  *resp.DisplayName,
		Description:  *resp.Description,
		SubjectClaim: *resp.SubjectClaim,
		Policy:       *resp.Policy,
	}

	return output.DescribeObject(cmd, identityPool, poolListFields, poolHumanLabelMap, poolStructuredLabelMap)
}
