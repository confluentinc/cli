package iam

import (
	"strings"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type ipGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type ipGroupHumanOut struct {
	ID         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
	CidrBlocks string `human:"CIDR blocks" serialized:"cidr_blocks"`
}

type ipGroupSerializedOut struct {
	ID         string   `human:"ID" serialized:"id"`
	Name       string   `human:"Name" serialized:"name"`
	CidrBlocks []string `human:"CIDR blocks" serialized:"cidr_blocks"`
}

func newIpGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip-group",
		Short: "Manage IP groups.",
		Long:  "Manage IP groups and their permissions.",
	}

	c := &ipGroupCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printIpGroup(cmd *cobra.Command, ipGroup iamv2.IamV2IpGroup) error {
	table := output.NewTable(cmd)
	if output.GetFormat(cmd) == output.Human {
		table.Add(&ipGroupHumanOut{
			ID:         ipGroup.GetId(),
			Name:       ipGroup.GetGroupName(),
			CidrBlocks: strings.Join(ipGroup.GetCidrBlocks(), ", "),
		})
	} else {
		table.Add(&ipGroupSerializedOut{
			ID:         ipGroup.GetId(),
			Name:       ipGroup.GetGroupName(),
			CidrBlocks: ipGroup.GetCidrBlocks(),
		})
	}
	return table.Print()
}

func (c *ipGroupCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteIpGroups(c.V2Client)
}
