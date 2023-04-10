package environment

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Confluent Cloud environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	environment := orgv2.OrgV2Environment{DisplayName: orgv2.PtrString(args[0])}
	environment, err := c.V2Client.CreateOrgEnvironment(environment)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: environment.GetId() == c.Context.GetCurrentEnvironment(),
		Id:        environment.GetId(),
		Name:      environment.GetDisplayName(),
	})
	if err := table.Print(); err != nil {
		return err
	}

	c.Context.AddEnvironment(environment.GetId())
	_ = c.Config.Save()

	return nil
}
