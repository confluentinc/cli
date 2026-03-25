package organization

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Use a Confluent Cloud organization in subsequent commands.",
		Long:              "Choose a Confluent Cloud organization to be used in subsequent commands. Switching organizations clears your active environment and Kafka cluster selections.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	return cmd
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	organizations, err := c.V2Client.ListOrgOrganizations()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(organizations))
	for i, org := range organizations {
		suggestions[i] = fmt.Sprintf("%s\t%s", org.GetId(), org.GetDisplayName())
	}
	return suggestions
}

func (c *command) use(_ *cobra.Command, args []string) error {
	id := args[0]

	// Use the list endpoint instead of get-by-ID because the org-scoped JWT
	// only authorizes reading the *current* organization. The list endpoint
	// returns every organization the user belongs to regardless of JWT scope.
	organizations, err := c.V2Client.ListOrgOrganizations()
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	found := false
	for _, org := range organizations {
		if org.GetId() == id {
			found = true
			break
		}
	}
	if !found {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`organization "%s" not found or access forbidden`, id),
			"List available organizations with `organization list`.",
		)
	}

	if id == c.Context.GetCurrentOrganization() {
		output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.Organization, id)
		return nil
	}

	if err := c.Context.SwitchOrganization(c.Client, id); err != nil {
		return fmt.Errorf("failed to switch to organization %q: %w", id, err)
	}

	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.Organization, id)
	return nil
}
