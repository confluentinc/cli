package organization

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
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

func formatTimestamp(ts *time.Time) string {
	if ts == nil {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}
