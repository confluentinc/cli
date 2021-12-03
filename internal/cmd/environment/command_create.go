package environment

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	createFields           = []string{"Name", "Id"}
	createHumanLabels      = map[string]string{"Name": "Environment Name", "Id": "ID"}
	createStructuredLabels = map[string]string{"Name": "name", "Id": "id"}
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Confluent Cloud environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
	}

	output.AddFlag(cmd)

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	account := &orgv1.Account{
		Name:           args[0],
		OrganizationId: c.State.Auth.Account.OrganizationId,
	}

	environment, err := c.Client.Account.Create(context.Background(), account)
	if err != nil {
		return err
	}

	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, environment.Id)

	return output.DescribeObject(cmd, environment, createFields, createHumanLabels, createStructuredLabels)
}
