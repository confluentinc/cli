package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pconv "github.com/confluentinc/cli/internal/pkg/name-conversions"
)

type identityPoolCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type identityPoolOut struct {
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

	c := &identityPoolCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newUseCommand())

	return cmd
}

func (c *identityPoolCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	provider, _ := cmd.Flags().GetString("provider")
	return pcmd.AutocompleteIdentityPools(c.V2Client, provider)
}

func (c *identityPoolCommand) poolAndProviderNamesToIds(pool string, provider string) (string, string, error) {
	provider, err := pconv.ConvertIamProviderNameToId(provider, c.V2Client)
	if err != nil {
		return pool, provider, err
	}
	pool, err = pconv.ConvertIamPoolNameToId(pool, provider, c.V2Client)
	if err != nil {
		return pool, provider, err
	}
	return pool, provider, err
}
