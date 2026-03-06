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

	if _, httpResp, err := c.V2Client.GetOrgOrganization(id); err != nil {
		return errors.CatchCCloudV2ResourceNotFoundError(err, resource.Organization, httpResp)
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
