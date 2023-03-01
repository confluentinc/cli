package environment

import (
	"context"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

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
	account := &ccloudv1.Account{
		Name:           args[0],
		OrganizationId: c.Context.GetOrganization().GetId(),
	}

	environment, err := c.Client.Account.Create(context.Background(), account)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: environment.Id == c.EnvironmentId(cmd),
		Id:        environment.Id,
		Name:      environment.Name,
	})
	return table.Print()
}
