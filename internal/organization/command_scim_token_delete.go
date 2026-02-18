package organization

import (
	"context"

	"github.com/spf13/cobra"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *scimTokenCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a SCIM token.",
		Long:  "Delete a SCIM token for the current organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *scimTokenCommand) delete(cmd *cobra.Command, args []string) error {
	scimClient, orgId, connectionName, err := c.getSCIMClient()
	if err != nil {
		return err
	}

	if err := deletion.ValidateAndConfirm(cmd, args, func(id string) bool { return true }, resource.ScimToken); err != nil {
		return err
	}

	_, err = deletion.Delete(cmd, args, func(tokenId string) error {
		req := &flowv1.DeleteSCIMTokenRequest{
			OrgResourceId:  orgId,
			ConnectionName: connectionName,
			TokenId:        tokenId,
		}
		return scimClient.DeleteToken(context.Background(), req)
	}, resource.ScimToken)

	return err
}
