package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newDekSubjectCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "subject",
		Short:       "Manage Schema Registry Data Encryption Key (DEK) subjects.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newDekSubjectListCommand(cfg))

	return cmd
}
