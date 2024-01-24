package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *poolCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an identity pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an identity pool named "demo-identity-pool" with identity provider "op-12345":`,
				Code: `confluent iam pool create demo-identity-pool --provider op-12345 --description "new description" --identity-claim claims.sub --filter 'claims.iss=="https://my.issuer.com"'`,
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("identity-claim", "", "Claim specifying the external identity using this identity pool.")
	cmd.Flags().String("description", "", "Description of the identity pool.")
	pcmd.AddFilterFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))
	cobra.CheckErr(cmd.MarkFlagRequired("identity-claim"))

	return cmd
}

func (c *poolCommand) create(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityClaim, err := cmd.Flags().GetString("identity-claim")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	createIdentityPool := identityproviderv2.IamV2IdentityPool{
		DisplayName:   identityproviderv2.PtrString(args[0]),
		Description:   identityproviderv2.PtrString(description),
		IdentityClaim: identityproviderv2.PtrString(identityClaim),
		Filter:        identityproviderv2.PtrString(filter),
	}
	pool, err := c.V2Client.CreateIdentityPool(createIdentityPool, provider)
	if err != nil {
		return err
	}

	return printIdentityPool(cmd, pool)
}
