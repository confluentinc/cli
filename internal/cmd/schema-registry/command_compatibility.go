package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) newCompatibilityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "compatibility",
		Short:       "Validate schema compatibility.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newCompatibilityValidateCommand())

	return cmd
}
