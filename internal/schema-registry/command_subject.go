package schemaregistry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func (c *command) newSubjectCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subject",
		Short: "Manage Schema Registry subjects.",
	}

	cmd.AddCommand(c.newSubjectDescribeCommand(cfg))
	cmd.AddCommand(c.newSubjectListCommand(cfg))
	cmd.AddCommand(c.newSubjectUpdateCommand(cfg))

	return cmd
}
