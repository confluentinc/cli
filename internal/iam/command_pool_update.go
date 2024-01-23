package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *poolCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an identity pool.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update the description of identity pool `pool-123456`:",
				Code: `confluent iam pool update pool-123456 --provider op-12345 --description "updated description"`,
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("identity-claim", "", "Claim specifying the external identity using this identity pool.")
	cmd.Flags().String("name", "", "Name of the identity pool.")
	cmd.Flags().String("description", "", "Description of the identity pool.")
	pcmd.AddFilterFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))
	cmd.MarkFlagsOneRequired("identity-claim", "name", "description", "filter")

	return cmd
}

func (c *poolCommand) update(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityClaim, err := cmd.Flags().GetString("identity-claim")
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
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

	updateIdentityPool := identityproviderv2.IamV2IdentityPool{Id: identityproviderv2.PtrString(args[0])}
	if name != "" {
		updateIdentityPool.DisplayName = identityproviderv2.PtrString(name)
	}
	if description != "" {
		updateIdentityPool.Description = identityproviderv2.PtrString(description)
	}
	if identityClaim != "" {
		updateIdentityPool.IdentityClaim = identityproviderv2.PtrString(identityClaim)
	}
	if filter != "" {
		updateIdentityPool.Filter = identityproviderv2.PtrString(filter)
	}

	pool, err := c.V2Client.UpdateIdentityPool(updateIdentityPool, provider)
	if err != nil {
		return err
	}

	return printIdentityPool(cmd, pool)
}
