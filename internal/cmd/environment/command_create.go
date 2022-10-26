package environment

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	fields            = []string{"Id", "Name"}
	humanRenames      = map[string]string{"Id": "ID"}
	structuredRenames = map[string]string{"Id": "id", "Name": "name"}
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

	return output.DescribeObject(cmd, environment, fields, humanRenames, structuredRenames)
}
