package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	identityPoolListFields           = []string{"Id", "DisplayName", "Description", "IdentityClaim", "Filter"}
	identityPoolListHumanLabels      = []string{"ID", "Display Name", "Description", "Identity Claim", "Filter"}
	identityPoolListStructuredLabels = []string{"id", "display_name", "description", "identity_claim", "filter"}
)

func (c *identityPoolCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List identity pools.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	_ = cmd.MarkFlagRequired("provider")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *identityPoolCommand) list(cmd *cobra.Command, _ []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityPools, err := c.V2Client.ListIdentityPools(provider)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, identityPoolListFields, identityPoolListHumanLabels, identityPoolListStructuredLabels)
	if err != nil {
		return err
	}
	for _, pool := range identityPools {
		element := &identityPool{
			Id:            *pool.Id,
			DisplayName:   *pool.DisplayName,
			IdentityClaim: *pool.IdentityClaim,
			Filter:        *pool.Filter,
		}
		if pool.Description != nil {
			element.Description = *pool.Description
		}
		outputWriter.AddElement(element)
	}
	return outputWriter.Out()
}
