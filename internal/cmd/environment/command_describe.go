package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type out struct {
	IsCurrent bool   `human:"Current" serialized:"is_current"`
	Id        string `human:"ID" serialized:"id"`
	Name      string `human:"Name" serialized:"name"`
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
	environment, httpResp, err := c.V2Client.GetOrgEnvironment(args[0])
	if err != nil {
		return errors.CatchEnvironmentNotFoundError(err, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: *environment.Id == c.EnvironmentId(),
		Id:        *environment.Id,
		Name:      *environment.DisplayName,
	})
	return table.Print()
}
