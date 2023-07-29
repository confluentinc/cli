package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (c *command) newSchemaCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "schema",
		Short:       "Manage Schema Registry schemas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

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
