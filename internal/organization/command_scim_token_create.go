package organization

import (
	"context"

	"github.com/spf13/cobra"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type scimTokenCreateOut struct {
	Id        string `human:"ID" serialized:"id"`
	Token     string `human:"Token" serialized:"token"`
	CreatedAt string `human:"Created At" serialized:"created_at"`
	ExpiresAt string `human:"Expires At" serialized:"expires_at"`
}

func (c *scimTokenCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a SCIM token.",
		Long:  "Create a SCIM token for the current organization.\n\nSave the token as it is not retrievable later.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	cmd.Flags().Int32("expire-duration-mins", 0, "Token expiration duration in minutes. Defaults to 6 months if not specified.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *scimTokenCommand) create(cmd *cobra.Command, _ []string) error {
	scimClient, orgId, connectionName, err := c.getSCIMClient()
	if err != nil {
		return err
	}

	req := &flowv1.CreateSCIMTokenRequest{
		OrgResourceId:  orgId,
		ConnectionName: connectionName,
	}

	// Only set expiration duration if explicitly provided
	if cmd.Flags().Changed("expire-duration-mins") {
		expireDurationMins, err := cmd.Flags().GetInt32("expire-duration-mins")
		if err != nil {
			return err
		}
		req.ExpireDurationMins = expireDurationMins
	}

	token, err := scimClient.CreateToken(context.Background(), req)
	if err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "Save the token as it is not retrievable later.")

	table := output.NewTable(cmd)
	table.Add(&scimTokenCreateOut{
		Id:        token.Id,
		Token:     token.Token,
		CreatedAt: formatTimestamp(token.CreatedAt),
		ExpiresAt: formatTimestamp(token.ExpiresAt),
	})
	return table.Print()
}
