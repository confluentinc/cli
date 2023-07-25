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

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newSchemaDeleteCommand())
		cmd.AddCommand(c.newSchemaDescribeCommand())
		cmd.AddCommand(c.newSchemaListCommand())
	} else {
		cmd.AddCommand(c.newSchemaDeleteCommandOnPrem())
		cmd.AddCommand(c.newSchemaDescribeCommandOnPrem())
		cmd.AddCommand(c.newSchemaListCommandOnPrem())
	}

	return cmd
}

func catchSchemaNotFoundError(err error, subject, version string) error {
	if err != nil && strings.Contains(err.Error(), "Not Found") {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`subject "%s" or version "%s" not found`, subject, version),
			"List available subjects with `confluent schema-registry subject list`.\nList available versions with `confluent schema-registry subject describe`.",
		)
	}

	return err
}
