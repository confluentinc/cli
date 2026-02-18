package organization

import (
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"

	ccloud "github.com/confluentinc/ccloud-sdk-go-v1"
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

func (c *scimTokenCommand) getSCIMClient() (ccloud.SCIM, string, string, error) {
	// Get organization from current user
	user, err := c.Client.Auth.User()
	if err != nil {
		return nil, "", "", err
	}
	org := user.GetOrganization()
	orgId := org.GetResourceId()

	// Get connection name from organization's SSO configuration
	sso := org.GetSso()
	if sso == nil || sso.GetAuth0ConnectionName() == "" {
		return nil, "", "", errors.NewErrorWithSuggestions(
			"No SSO connection found for organization.",
			"SCIM tokens require SSO to be configured for the organization.",
		)
	}
	connectionName := sso.GetAuth0ConnectionName()

	// Create v1 (non-public) client for SCIM operations
	params := &ccloud.Params{
		BaseURL:    c.Context.GetPlatformServer(),
		HttpClient: c.Client.HttpClient,
		Logger:     c.Client.Logger,
		UserAgent:  c.Config.Version.UserAgent,
	}
	v1Client := ccloud.NewClient(params)

	return v1Client.SCIM, orgId, connectionName, nil
}

func formatTimestamp(ts *types.Timestamp) string {
	if ts == nil {
		return ""
	}
	t := time.Unix(ts.Seconds, int64(ts.Nanos))
	return t.UTC().Format(time.RFC3339)
}
