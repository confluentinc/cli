package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (c *command) newConfigCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "Manage Schema Registry configuration.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newConfigDescribeCommand(cfg))
	cmd.AddCommand(c.newConfigDeleteCommand(cfg))

	return cmd
}

func catchSubjectLevelConfigNotFoundError(err error, subject string) error {
	if err != nil && strings.Contains(err.Error(), "Not Found") {
		return errors.New(fmt.Sprintf(`subject "%s" does not have subject-level compatibility configured`, subject))
	}

	return err
}

func prettyJson(str []byte) string {
	return strings.TrimSpace(string(pretty.Pretty(str)))
}
