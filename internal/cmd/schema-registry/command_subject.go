package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

func (c *command) newSubjectCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "subject",
		Short:       "Manage Schema Registry subjects.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newSubjectDescribeCommand(cfg))
	cmd.AddCommand(c.newSubjectListCommand(cfg))
	cmd.AddCommand(c.newSubjectUpdateCommand(cfg))

	return cmd
}
