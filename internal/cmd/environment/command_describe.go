package environment

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a Confluent Cloud environment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	environment, r, err := c.V2Client.GetOrgEnvironment(args[0])
	if err != nil {
		return errors.CatchEnvironmentNotFoundError(err, r)
	}

	account := &orgv1.Account{
		Id:   *environment.Id,
		Name: *environment.DisplayName,
	}

	return output.DescribeObject(cmd, account, fields, humanRenames, structuredRenames)
}
