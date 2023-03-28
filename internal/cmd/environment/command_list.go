package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud environments.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environments, err := c.V2Client.ListOrgEnvironments()
	if err != nil {
		return err
	}

	environmentId, _ := c.EnvironmentId()
	list := output.NewList(cmd)
	for _, environment := range environments {
		list.Add(&out{
			IsCurrent: environment.GetId() == environmentId,
			Id:        environment.GetId(),
			Name:      environment.GetDisplayName(),
		})
	}
	return list.Print()
}
