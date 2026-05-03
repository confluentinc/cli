package organization

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

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
	if err := c.validateSSOConfigured(); err != nil {
		return err
	}

	token := orgv2.InlineObject{}

	// Only set expiration duration if explicitly provided
	if cmd.Flags().Changed("expire-duration-mins") {
		expireDurationMins, err := cmd.Flags().GetInt32("expire-duration-mins")
		if err != nil {
			return err
		}
		token.ExpireDurationMins = orgv2.PtrInt32(expireDurationMins)
	}

	createdToken, err := c.V2Client.CreateScimToken(token)
	if err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "Save the token as it is not retrievable later.")

	table := output.NewTable(cmd)
	table.Add(&scimTokenCreateOut{
		Id:        createdToken.GetId(),
		Token:     createdToken.GetToken(),
		CreatedAt: formatTimestamp(createdToken.CreatedAt),
		ExpiresAt: formatTimestamp(createdToken.ExpiresAt),
	})
	return table.Print()
}
