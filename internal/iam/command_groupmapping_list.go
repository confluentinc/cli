package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *groupMappingCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List group mappings.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *groupMappingCommand) list(cmd *cobra.Command, _ []string) error {
	groupMappings, err := c.V2Client.ListGroupMappings()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, groupMapping := range groupMappings {
		list.Add(&groupMappingOut{
			Id:          groupMapping.GetId(),
			Name:        groupMapping.GetDisplayName(),
			Description: groupMapping.GetDescription(),
			Filter:      groupMapping.GetFilter(),
			Principal:   groupMapping.GetPrincipal(),
			State:       groupMapping.GetState(),
		})
	}
	return list.Print()
}
