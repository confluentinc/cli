package organization

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

type scimTokenCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func (c *command) newScimTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scim-token",
		Short: "Manage SCIM tokens.",
		Long:  "Manage SCIM tokens for the current organization.",
	}

	scimCmd := &scimTokenCommand{c.AuthenticatedCLICommand}

	cmd.AddCommand(scimCmd.newCreateCommand())
	cmd.AddCommand(scimCmd.newListCommand())
	cmd.AddCommand(scimCmd.newDeleteCommand())

	return cmd
}

func (c *scimTokenCommand) validateSSOConfigured() error {
	// Get organization from current user
	user, err := c.Client.Auth.User()
	if err != nil {
		return err
	}
	org := user.GetOrganization()

	// Get connection name from organization's SSO configuration
	sso := org.GetSso()
	if sso == nil || sso.GetAuth0ConnectionName() == "" {
		return errors.NewErrorWithSuggestions(
			"No SSO connection found for organization.",
			"SCIM tokens require SSO to be configured for the organization.",
		)
	}

	return nil
}

func formatTimestamp(ts *time.Time) string {
	if ts == nil {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}
