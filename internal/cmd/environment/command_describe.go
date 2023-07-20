package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	nameconversions "github.com/confluentinc/cli/internal/pkg/name-conversions"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type out struct {
	IsCurrent bool   `human:"Current" serialized:"is_current"`
	Id        string `human:"ID" serialized:"id"`
	Name      string `human:"Name" serialized:"name"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [id|name]",
		Short:             "Describe a Confluent Cloud environment.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	id := c.Context.GetCurrentEnvironment()
	var err error
	if len(args) > 0 {
		id = args[0]
	}
	if id == "" {
		return errors.NewErrorWithSuggestions("no environment selected", "Select an environment with `confluent environment use` or as an argument.")
	}

	environment, err := c.V2Client.GetOrgEnvironment(id)
	if err != nil {
		if id, err = nameconversions.EnvironmentNameToId(id, c.V2Client, false); err != nil {
			return errors.NewErrorWithSuggestions(err.Error(), errors.NotValidEnvironmentIdSuggestions)
		}
		if environment, err = c.V2Client.GetOrgEnvironment(id); err != nil {
			return errors.NewErrorWithSuggestions(err.Error(), errors.NotValidEnvironmentIdSuggestions)
		}
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: environment.GetId() == c.Context.GetCurrentEnvironment(),
		Id:        environment.GetId(),
		Name:      environment.GetDisplayName(),
	})
	return table.Print()
}
