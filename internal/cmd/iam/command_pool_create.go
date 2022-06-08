package iam

import (
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *identityPoolCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an identity pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an identity pool named `DemoIdentityPool`.",
				Code: `confluent iam pool create DemoIdentityPool --provider op-12345 --description new-description --subject-claim sub`,
			},
		),
	}

	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().String("description", "", "Description of the identity provider.")
	_ = cmd.MarkFlagRequired("description")
	cmd.Flags().String("subject-claim", "", "Subject claim of the identity pool.")
	_ = cmd.MarkFlagRequired("subject-claim")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *identityPoolCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := requireLen(name, nameLength, "pool name"); err != nil {
		return err
	}

	providerID, err := cmd.Flags().GetString("provider")
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

	subjectClaim, err := cmd.Flags().GetString("subject-claim")
	if err != nil {
		return err
	}

	createIdentityPool := identityproviderv2.IamV2IdentityPool{
		DisplayName:  identityproviderv2.PtrString(name),
		Description:  identityproviderv2.PtrString(description),
		SubjectClaim: identityproviderv2.PtrString(subjectClaim),
	}
	resp, httpResp, err := c.V2Client.CreateIdentityPool(createIdentityPool, providerID)
	if err != nil {
		return errors.CatchServiceNameInUseError(err, httpResp, name)
	}

	describeIdentityPool := &identityPool{Id: *resp.Id, DisplayName: *resp.DisplayName, Description: *resp.Description, SubjectClaim: *resp.SubjectClaim}

	return output.DescribeObject(cmd, describeIdentityPool, poolListFields, poolHumanLabelMap, poolStructuredLabelMap)
}
