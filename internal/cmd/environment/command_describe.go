package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type out struct {
	Id   string `human:"ID" json:"id" yaml:"id"`
	Name string `human:"Name" json:"name" yaml:"name"`
}

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

	table := output.NewTable(cmd)
	table.Add(&out{
		Id:   *environment.Id,
		Name: *environment.DisplayName,
	})
	return table.Print()
}
