package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type poolCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type poolOut struct {
	Id            string `human:"ID" serialized:"id"`
	DisplayName   string `human:"Name" serialized:"name"`
	Description   string `human:"Description" serialized:"description"`
	IdentityClaim string `human:"Identity Claim" serialized:"identity_claim"`
	Filter        string `human:"Filter" serialized:"filter"`
}

func newPoolCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "pool",
		Short:       "Manage identity pools.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &poolCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printIdentityPool(cmd *cobra.Command, pool identityproviderv2.IamV2IdentityPool) error {
	table := output.NewTable(cmd)
	table.Add(&poolOut{
		Id:            pool.GetId(),
		DisplayName:   pool.GetDisplayName(),
		Description:   pool.GetDescription(),
		IdentityClaim: pool.GetIdentityClaim(),
		Filter:        pool.GetFilter(),
	})
	return table.Print()
}

func (c *poolCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *poolCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	provider, _ := cmd.Flags().GetString("provider")
	return pcmd.AutocompleteIdentityPools(c.V2Client, provider)
}
