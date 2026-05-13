package org

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *scimTokenCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List org scim tokens.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	// Required flags

	// Optional flags

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *scimTokenCommand) list(cmd *cobra.Command, _ []string) error {

	scimTokens, err := c.V2Client.ListOrgScimTokens()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, scimToken := range scimTokens {
		out := &scimTokenOut{
			ID:             scimToken.GetId(),
			ConnectionName: scimToken.GetConnectionName(),
			Token:          scimToken.GetToken(),
			CreatedAt:      scimToken.GetCreatedAt().String(),
			ExpiresAt:      scimToken.GetExpiresAt().String(),
		}
		list.Add(out)
	}
	return list.Print()
}
