package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *command) newSubjectCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "subject",
		Short:       "Manage Schema Registry subjects.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newSubjectDescribeCommand())
		cmd.AddCommand(c.newSubjectListCommand())
		cmd.AddCommand(c.newSubjectUpdateCommand())
	} else {
		cmd.AddCommand(c.newSubjectDescribeCommandOnPrem())
		cmd.AddCommand(c.newSubjectListCommandOnPrem())
		cmd.AddCommand(c.newSubjectUpdateCommandOnPrem())
	}

	return cmd
}
