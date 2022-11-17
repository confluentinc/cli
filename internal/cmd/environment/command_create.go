package environment

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

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

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	account := &orgv1.Account{
		Name:           args[0],
		OrganizationId: c.Context.GetOrganization().GetId(),
	}

	environment, err := c.Client.Account.Create(context.Background(), account)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: environment.Id == c.EnvironmentId(),
		Id:        environment.Id,
		Name:      environment.Name,
	})
	return table.Print()
}
