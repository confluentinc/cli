package iam

import (
	"github.com/spf13/cobra"
	"strings"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *ipGroupCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IP Groups.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipGroupCommand) list(cmd *cobra.Command, _ []string) error {

	ipGroups, err := c.V2Client.ListIamIpGroups()
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, group := range ipGroups {
			list.Add(&ipGroupHumanOut{
				ID:         group.GetId(),
				Name:       group.GetGroupName(),
				CidrBlocks: strings.Join(group.GetCidrBlocks(), ", "),
			})
		}
		return list.Print()
	}

	list := output.NewList(cmd)
	for _, group := range ipGroups {
		list.Add(&ipGroupSerializedOut{
			ID:         group.GetId(),
			Name:       group.GetGroupName(),
			CidrBlocks: group.GetCidrBlocks(),
		})
	}
	return list.Print()

}
