package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
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

	list := output.NewList(cmd)
	for _, environment := range environments {
		list.Add(&out{
			IsCurrent:               environment.GetId() == c.Context.GetCurrentEnvironment(),
			Id:                      environment.GetId(),
			Name:                    environment.GetDisplayName(),
			StreamGovernancePackage: environment.StreamGovernanceConfig.GetPackage(),
		})
	}
	return list.Print()
}
