package organization

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type scimTokenListOut struct {
	Id        string `human:"ID" serialized:"id"`
	CreatedAt string `human:"Created At" serialized:"created_at"`
	ExpiresAt string `human:"Expires At" serialized:"expires_at"`
}

func (c *scimTokenCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SCIM tokens.",
		Long:  "List SCIM tokens for the current organization.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *scimTokenCommand) list(cmd *cobra.Command, _ []string) error {
	tokens, err := c.V2Client.ListScimTokens()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, token := range tokens {
		list.Add(&scimTokenListOut{
			Id:        token.GetId(),
			CreatedAt: formatTimestamp(token.CreatedAt),
			ExpiresAt: formatTimestamp(token.ExpiresAt),
		})
	}
	return list.Print()
}
