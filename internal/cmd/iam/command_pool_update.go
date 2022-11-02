package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

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
				Text: `Update the description of identity pool "pool-123456":`,
				Code: `confluent iam pool update pool-123456 --provider op-12345 --description "New description."`,
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("name", "", "Name of the identity pool.")
	cmd.Flags().String("description", "", "Description of the identity pool.")
	cmd.Flags().String("filter", "", "Filter which identities can authenticate with the identity pool.")
	cmd.Flags().String("identity-claim", "", "Claim specifying the external identity using this identity pool.")
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

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityClaim, err := cmd.Flags().GetString("identity-claim")
	if err != nil {
		return err
	}

	if description == "" && filter == "" && identityClaim == "" && name == "" {
		return errors.New(errors.IdentityPoolNoOpUpdateErrorMsg)
	}

	identityPoolId := args[0]
	updateIdentityPool := identityproviderv2.IamV2IdentityPool{Id: &identityPoolId}
	if name != "" {
		updateIdentityPool.DisplayName = &name
	}
	if description != "" {
		updateIdentityPool.Description = &description
	}
	if identityClaim != "" {
		updateIdentityPool.IdentityClaim = &identityClaim
	}
	if filter != "" {
		updateIdentityPool.Filter = &filter
	}

	pool, err := c.V2Client.UpdateIdentityPool(updateIdentityPool, provider)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&identityPoolOut{
		Id:            pool.GetId(),
		DisplayName:   pool.GetDisplayName(),
		Description:   pool.GetDescription(),
		IdentityClaim: pool.GetIdentityClaim(),
		Filter:        pool.GetFilter(),
	})
	return table.Print()
}
