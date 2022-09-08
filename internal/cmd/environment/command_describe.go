package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	fields            = []string{"Id", "Name"}
	humanRenames      = map[string]string{"Id": "ID"}
	structuredRenames = map[string]string{"Id": "id", "Name": "name"}
)

func (c *command) newDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a Confluent Cloud environment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	environment, r, err := c.V2Client.GetOrgEnvironment(args[0])
	if err != nil {
		return errors.CatchEnvironmentNotFoundError(err, r)
	}

	return output.DescribeObject(cmd, environment, fields, humanRenames, structuredRenames)
}
