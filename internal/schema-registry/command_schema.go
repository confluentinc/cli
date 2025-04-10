package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

func (c *command) newSchemaCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage Schema Registry schemas.",
	}

	cmd.AddCommand(c.newSchemaCompatibilityCommand(cfg))
	cmd.AddCommand(c.newSchemaCreateCommand(cfg))
	cmd.AddCommand(c.newSchemaDeleteCommand(cfg))
	cmd.AddCommand(c.newSchemaDescribeCommand(cfg))
	cmd.AddCommand(c.newSchemaListCommand(cfg))

	return cmd
}

func catchSchemaNotFoundError(err error, subject, version string) error {
	if err != nil && strings.Contains(err.Error(), "Not Found") {
		message := fmt.Sprintf(`subject "%s" `, subject)
		if version != "" {
			message += fmt.Sprintf(`or version "%s" `, version)
		}
		message += "not found"

		return errors.NewErrorWithSuggestions(
			message,
			"List available subjects with `confluent schema-registry subject list`.\nList available versions with `confluent schema-registry subject describe`.",
		)
	}

	return err
}
